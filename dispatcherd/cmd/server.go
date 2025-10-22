package main

import (
	"context"
	"dispatcherd/handler"
	"dispatcherd/logging"
	"dispatcherd/middleware"
	"dispatcherd/service"
	"errors"
	"fmt"
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
	ListenAddress  string
	CorsOrigin     string
	MessageService service.MessageService
}

type Server struct {
	ListenAddress  string
	router         chi.Router
	corsOrigin     string
	messageService service.MessageService
}

func NewServer(opts ServerOptions) *Server {
	return &Server{
		ListenAddress:  opts.ListenAddress,
		router:         chi.NewRouter(),
		corsOrigin:     opts.CorsOrigin,
		messageService: opts.MessageService,
	}
}

func (s *Server) Start() {
	logger := logging.GetLogger(logging.API)

	corsOptions := cors.Options{
		AllowedOrigins: []string{s.corsOrigin},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE"},
	}

	// register middleware
	requestIDMiddleware := middleware.NewUUIDv4RequestIDMiddleWare()
	requestLoggerMiddleware := middleware.NewRequestLoggerMiddleware()

	s.router.Use(cors.New(corsOptions).Handler)
	s.router.Use(middleware.SecurityHeaders())
	s.router.Use(requestIDMiddleware.OnRequest)
	s.router.Use(requestLoggerMiddleware.OnRequest)

	s.router.Use(chiMiddleware.AllowContentType("application/json"))
	s.router.Use(chiMiddleware.Recoverer)

	dispatchHandler := handler.NewDispatchHandler(s.messageService)

	// register public routes
	s.router.Get("/health", handler.Make(handler.HandleHealth))
	s.router.Post("/message", handler.Make(dispatchHandler.HandlePost))

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
				logger.Warn("context deadline exceeded, forcing shutdown")
			}
		}()

		// Trigger graceful shutdown
		logger.Info("received signal to shut down server gracefully")
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			logger.Error("failed to shutdown server gracefully", logging.FieldError, err)
		}
		serverStopCtx()
	}()

	// start listening for connections
	logger.Info("listening on " + s.ListenAddress)
	err := server.ListenAndServe()

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("failed to start server on "+s.ListenAddress, logging.FieldError, err)
		panic(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
