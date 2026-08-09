// Package http поднимает HTTP-сервер и корректно его останавливает.
//
// Вынесено из main, чтобы composition root остался плоским: main только
// собирает зависимости, а как именно живёт и умирает сервер — знает этот пакет.
package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server — HTTP-сервер сервиса вместе с его жизненным циклом.
type Server struct {
	server          *http.Server
	logger          *slog.Logger
	shutdownTimeout time.Duration
}

// NewServer собирает сервер с обязательными таймаутами.
//
// Без них медленный клиент держит соединение вечно. Подробный разбор —
// Модуль 12.
func NewServer(addr string, handler http.Handler, logger *slog.Logger, shutdownTimeout time.Duration) *Server {
	return &Server{
		server: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
		logger:          logger,
		shutdownTimeout: shutdownTimeout,
	}
}

// Run запускает сервер и блокируется до сигнала остановки или ошибки.
func (s *Server) Run() error {
	// Слушаем сигналы остановки: SIGINT от Ctrl+C, SIGTERM от Docker и k8s.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("сервис запущен", slog.String("addr", s.server.Addr))
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.logger.Info("получен сигнал остановки, дожидаемся активных запросов",
			slog.Duration("timeout", s.shutdownTimeout))
	}

	// Graceful shutdown: перестаём принимать новые соединения и даём
	// активным запросам доработать, но не дольше таймаута.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		return err
	}

	s.logger.Info("сервис остановлен")
	return nil
}
