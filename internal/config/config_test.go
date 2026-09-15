package config

import (
	"log/slog"
	"os"
	"testing"
	"time"
)

// Пример того, как выглядит table-driven тест в этом репозитории.
// Именно такой стиль ждём в домашних заданиях.
//
// Тест читает окружение, поэтому первым делом он это окружение себе готовит:
// иначе экспортированный в шелле ORDERS_LOG_LEVEL красит его в красный, хотя
// с кодом всё в порядке. Тест, который зависит от того, где его запустили, —
// не тест.
func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    Config
		wantErr bool
	}{
		{
			name: "значения по умолчанию",
			env:  nil,
			want: Config{
				HTTPAddr:        ":8080",
				ShutdownTimeout: 10 * time.Second,
				LogLevel:        slog.LevelInfo,
			},
		},
		{
			name: "переопределение из окружения",
			env: map[string]string{
				"ORDERS_HTTP_ADDR":        ":9090",
				"ORDERS_SHUTDOWN_TIMEOUT": "3s",
				"ORDERS_LOG_LEVEL":        "debug",
			},
			want: Config{
				HTTPAddr:        ":9090",
				ShutdownTimeout: 3 * time.Second,
				LogLevel:        slog.LevelDebug,
			},
		},
		{
			name: "уровень info задан явно",
			env:  map[string]string{"ORDERS_LOG_LEVEL": "info"},
			want: Config{
				HTTPAddr:        ":8080",
				ShutdownTimeout: 10 * time.Second,
				LogLevel:        slog.LevelInfo,
			},
		},
		{
			name: "уровень warn",
			env:  map[string]string{"ORDERS_LOG_LEVEL": "warn"},
			want: Config{
				HTTPAddr:        ":8080",
				ShutdownTimeout: 10 * time.Second,
				LogLevel:        slog.LevelWarn,
			},
		},
		{
			name: "warning — синоним warn",
			env:  map[string]string{"ORDERS_LOG_LEVEL": "warning"},
			want: Config{
				HTTPAddr:        ":8080",
				ShutdownTimeout: 10 * time.Second,
				LogLevel:        slog.LevelWarn,
			},
		},
		{
			name: "регистр и пробелы не важны",
			env:  map[string]string{"ORDERS_LOG_LEVEL": "  ERROR "},
			want: Config{
				HTTPAddr:        ":8080",
				ShutdownTimeout: 10 * time.Second,
				LogLevel:        slog.LevelError,
			},
		},
		{
			name: "пустой адрес — значение по умолчанию",
			env:  map[string]string{"ORDERS_HTTP_ADDR": ""},
			want: Config{
				HTTPAddr:        ":8080",
				ShutdownTimeout: 10 * time.Second,
				LogLevel:        slog.LevelInfo,
			},
		},
		{
			name:    "некорректный таймаут — ошибка на старте",
			env:     map[string]string{"ORDERS_SHUTDOWN_TIMEOUT": "быстро"},
			wantErr: true,
		},
		{
			name:    "неизвестный уровень логирования — ошибка на старте",
			env:     map[string]string{"ORDERS_LOG_LEVEL": "trace"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEnv(t, tt.env)

			got, err := Load()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Load() = %+v, ожидалась ошибка", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() вернул неожиданную ошибку: %v", err)
			}
			if got != tt.want {
				t.Errorf("Load() = %+v, ожидалось %+v", got, tt.want)
			}
		})
	}
}

// setEnv готовит окружение подтеста: сначала снимает все переменные сервиса,
// потом выставляет те, что нужны кейсу.
//
// t.Setenv здесь нужен ради его cleanup — он запоминает прежнее значение и
// вернёт его после теста; os.Unsetenv сразу за ним убирает переменную, чтобы
// Load увидел её незаданной (пустая строка для ORDERS_LOG_LEVEL — это уже
// другой случай, ошибка разбора).
func setEnv(t *testing.T, env map[string]string) {
	t.Helper()

	for _, key := range []string{"ORDERS_HTTP_ADDR", "ORDERS_SHUTDOWN_TIMEOUT", "ORDERS_LOG_LEVEL"} {
		t.Setenv(key, "")
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("не удалось снять %s: %v", key, err)
		}
	}

	for k, v := range env {
		t.Setenv(k, v)
	}
}
