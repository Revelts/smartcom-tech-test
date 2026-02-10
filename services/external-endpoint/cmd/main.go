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
	"github.com/smartcom/integration-platform/pkg/logger"
	"github.com/smartcom/integration-platform/services/external-endpoint/internal/handler"
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
	log.Info("starting external alert endpoint service")

	port := config.GetEnv("PORT", "8081")

	alertHandler := handler.NewAlertHandler(log)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	alertHandler.RegisterRoutes(router)

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

		log.Info("external endpoint service shutdown complete")
	}

	return
}
