package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpadapter "github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/http"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/memory"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/pix"
	clinicapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/clinic"
	dentistapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/dentist"
	paymentapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/payment"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	clinicRepo := memory.NewClinicRepository()
	dentistRepo := memory.NewDentistRepository()
	paymentRepo := memory.NewPaymentRepository()
	pixProvider := pix.NewDefaultSimulator()

	clinicService := clinicapp.NewService(clinicRepo)
	dentistService := dentistapp.NewService(dentistRepo, clinicRepo)
	paymentService := paymentapp.NewService(paymentRepo, clinicRepo, dentistRepo, pixProvider)

	router := httpadapter.NewRouter(
		httpadapter.NewClinicHandler(clinicService),
		httpadapter.NewDentistHandler(dentistService),
		httpadapter.NewPaymentHandler(paymentService),
		cfg.OpenAPIPath,
	)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		logger.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}
