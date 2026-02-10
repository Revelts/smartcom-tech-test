package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/smartcom/integration-platform/pkg/config"
	"github.com/smartcom/integration-platform/pkg/httpclient"
	"github.com/smartcom/integration-platform/pkg/logger"
	"github.com/smartcom/integration-platform/services/middleware/internal/handler"
	"github.com/smartcom/integration-platform/services/middleware/internal/infrastructure"
	"github.com/smartcom/integration-platform/services/middleware/internal/repository"
	"github.com/smartcom/integration-platform/services/middleware/internal/usecase"
	"github.com/smartcom/integration-platform/services/middleware/internal/worker"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "application error: %v\n", err)
		os.Exit(1)
	}
}

func run() (err error) {
	log := logger.NewDefault()
	log.Info("starting middleware integration service")

	port := config.GetEnv("PORT", "8080")
	externalURL := config.GetEnv("EXTERNAL_ENDPOINT_URL", "http://localhost:8081/external/alerts")
	queueSize := config.GetEnvInt("QUEUE_SIZE", 1000)
	workerCount := config.GetEnvInt("WORKER_COUNT", 10)
	httpTimeout := config.GetEnvDuration("HTTP_TIMEOUT", 3*time.Second)
	maxRetries := config.GetEnvInt("MAX_RETRIES", 3)
	baseDelay := config.GetEnvDuration("BASE_DELAY", 500*time.Millisecond)

	httpClientConfig := httpclient.Config{
		Timeout:    httpTimeout,
		MaxRetries: maxRetries,
		BaseDelay:  baseDelay,
	}
	httpClient := httpclient.New(httpClientConfig)

	idGenerator := infrastructure.NewUUIDGenerator()
	eventMapper := usecase.NewEventMapper(idGenerator)
	eventProcessor := usecase.NewEventProcessor(httpClient, externalURL, log)
	eventQueue := repository.NewEventQueue(queueSize)

	workerPool := worker.NewPool(workerCount, eventQueue, eventProcessor, log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workerPool.Start(ctx)

	eventHandler := handler.NewEventHandler(eventQueue, eventMapper, log)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	eventHandler.RegisterRoutes(router)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Info("http server listening", "port", port)
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err = <-serverErrors:
		err = fmt.Errorf("server error: %w", err)
		return
	case sig := <-shutdown:
		log.Info("received shutdown signal", "signal", sig.String())

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		err = server.Shutdown(shutdownCtx)
		if err != nil {
			err = server.Close()
			err = fmt.Errorf("failed to gracefully shutdown server: %w", err)
			return
		}

		cancel()

		workerShutdownCtx, workerShutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer workerShutdownCancel()

		workerPool.Shutdown(workerShutdownCtx)

		log.Info("middleware service shutdown complete")
	}

	return
}
