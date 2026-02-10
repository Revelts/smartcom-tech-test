package worker

import (
	"context"
	"sync"

	"github.com/smartcom/integration-platform/services/middleware/internal/domain"
	"github.com/smartcom/integration-platform/services/middleware/internal/repository"
)

type Pool struct {
	workerCount int
	queue       *repository.EventQueue
	processor   domain.EventProcessor
	logger      WorkerLogger
	wg          sync.WaitGroup
}

type WorkerLogger interface {
	InfoContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
}

const DefaultWorkerCount = 10

func NewPool(workerCount int, queue *repository.EventQueue, processor domain.EventProcessor, logger WorkerLogger) (pool *Pool) {
	if workerCount <= 0 {
		workerCount = DefaultWorkerCount
	}

	pool = &Pool{
		workerCount: workerCount,
		queue:       queue,
		processor:   processor,
		logger:      logger,
	}
	return
}

func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		workerID := i + 1
		go p.worker(ctx, workerID)
	}

	p.logger.InfoContext(ctx, "worker pool started", "worker_count", p.workerCount)
}

func (p *Pool) worker(ctx context.Context, workerID int) {
	defer p.wg.Done()

	workerCtx := context.WithValue(ctx, "worker_id", workerID)
	p.logger.InfoContext(workerCtx, "worker started")

	for {
		event, ok := p.queue.Dequeue(ctx)
		if !ok {
			p.logger.InfoContext(workerCtx, "worker shutting down")
			return
		}

		err := p.processor.ProcessEvent(event)
		if err != nil {
			p.logger.ErrorContext(workerCtx, "failed to process event",
				"event_id", event.ID,
				"error", err.Error(),
			)
		}
	}
}

func (p *Pool) Shutdown(ctx context.Context) {
	p.logger.InfoContext(ctx, "shutting down worker pool")

	p.queue.Close()

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		p.logger.InfoContext(ctx, "worker pool shutdown complete")
	case <-ctx.Done():
		p.logger.ErrorContext(ctx, "worker pool shutdown timeout")
	}
}
