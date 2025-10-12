package main

import (
	"context"
	"dispatcherd/handler"
	"dispatcherd/logging"
	"dispatcherd/middleware"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
)

type ServerOptions struct {
	ListenAddress string
	Logger        *slog.Logger
	CorsOrigin    string
}

type Server struct {
	ListenAddress string
	logger        *slog.Logger
	router        chi.Router
	corsOrigin    string
}

func NewServer(opts ServerOptions) *Server {
	return &Server{
		ListenAddress: opts.ListenAddress,
		router:        chi.NewRouter(),
		logger:        opts.Logger,
		corsOrigin:    opts.CorsOrigin,
	}
}

func (s *Server) Start() {
	corsOptions := cors.Options{
		AllowedOrigins: []string{s.corsOrigin},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE"},
	}

	// register middleware
	s.router.Use(cors.New(corsOptions).Handler)
	s.router.Use(middleware.SecurityHeaders())
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RequestLogger(s.logger))

	s.router.Use(chiMiddleware.AllowContentType("application/json"))
	s.router.Use(chiMiddleware.Recoverer)

	// register public routes
	s.router.Get("/health", handler.Make(handler.HandleHealth))

	// setup default handlers
	s.router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		handler.RespondError(w, r, http.StatusNotFound, fmt.Errorf("%s not found", r.URL.Path))
	})
	s.router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		handler.RespondError(w, r, http.StatusMethodNotAllowed,
			fmt.Errorf("method %s not allowed on %s", r.Method, r.URL.Path))
	})

	// setup graceful shutdown
	server := &http.Server{
		Addr:    s.ListenAddress,
		Handler: s.router,
		//nolint:mnd // just a default to prevent slow loris
		ReadHeaderTimeout: 5 * time.Second,
	}
	serverCtx, serverStopCtx := context.WithCancel(context.Background())
	// Listen for syscall signals for the process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		//nolint:govet,mnd // does not matter if we cancel, as the application is terminated anyway
		//goland:noinspection ALL
		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
				s.logger.Warn("context deadline exceeded, forcing shutdown")
			}
		}()

		// Trigger graceful shutdown
		s.logger.Info("received signal to shut down server gracefully")
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			s.logger.Error("failed to shutdown server gracefully", logging.LoggerFieldError, err)
		}
		serverStopCtx()
	}()

	// start listening for connections
	s.logger.Info("listening on " + s.ListenAddress)
	err := server.ListenAndServe()

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error("failed to start server on "+s.ListenAddress, logging.LoggerFieldError, err)
		panic(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
