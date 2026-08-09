// Package config собирает конфигурацию сервиса из переменных окружения.
//
// Правило курса: никаких флагов и файлов настроек в коде сервиса — только
// окружение (12-factor). Дальше по курсу сюда добавятся DSN базы (Модуль 5),
// адрес брокера (Модуль 8) и параметры таймаутов (Модуль 12).
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

// Config — все настройки сервиса в одном месте.
type Config struct {
	// HTTPAddr — адрес, который слушает HTTP-сервер, например ":8080".
	HTTPAddr string
	// ShutdownTimeout — сколько ждём завершения активных запросов при остановке.
	ShutdownTimeout time.Duration
	// LogLevel — уровень логирования.
	LogLevel slog.Level
}

// Load читает конфигурацию из окружения, подставляя значения по умолчанию.
// Возвращает ошибку, если значение задано, но некорректно: сервис должен
// падать на старте, а не работать с непонятной конфигурацией.
func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:        env("ORDERS_HTTP_ADDR", ":8080"),
		ShutdownTimeout: 10 * time.Second,
		LogLevel:        slog.LevelInfo,
	}

	if raw, ok := os.LookupEnv("ORDERS_SHUTDOWN_TIMEOUT"); ok {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return Config{}, fmt.Errorf("ORDERS_SHUTDOWN_TIMEOUT=%q: %w", raw, err)
		}
		cfg.ShutdownTimeout = d
	}

	if raw, ok := os.LookupEnv("ORDERS_LOG_LEVEL"); ok {
		lvl, err := parseLevel(raw)
		if err != nil {
			return Config{}, err
		}
		cfg.LogLevel = lvl
	}

	return cfg, nil
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func parseLevel(raw string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("ORDERS_LOG_LEVEL=%q: ожидается debug|info|warn|error", raw)
	}
}
