package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

type AppConfig struct {
	App struct {
		Version     string `json:"version"`
		Environment string `json:"environment"`
	} `json:"app"`
	Logging struct {
		Level  string `json:"level"`
		Format string `json:"format"`
	} `json:"logging"`
	Message struct {
		Hello string `json:"hello"`
	} `json:"message"`
}

type Secret struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}

type Response struct {
	Message string     `json:"hello"`
	Config  *AppConfig `json:"config"`
	Secret  *Secret    `json:"secrets,omitempty"`
}

func loadConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func loadSecretsFromEnv() *Secret {
	return &Secret{
		User: os.Getenv("USER"),
		Pass: os.Getenv("PASS"),
	}
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func setupSlog(cfg *AppConfig) *slog.Logger {
	var handler slog.Handler

	level := parseLogLevel(cfg.Logging.Level)

	switch strings.ToLower(cfg.Logging.Format) {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	default:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

func main() {
	const configPath = "./config/config.json"

	cfg, err := loadConfig(configPath)
	if err != nil {
		log := slog.New(slog.NewTextHandler(os.Stderr, nil))
		log.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	secrets := loadSecretsFromEnv()

	logger := setupSlog(cfg)
	logger.Info("Config loaded", "enviroment", cfg.App.Environment, "version", cfg.App.Version)

	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Handling /config request", "remote", r.RemoteAddr)
		resp := Response{
			Message: "Configuration",
			Config:  cfg,
			Secret:  secrets,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	http.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Handling /version request", "remote", r.RemoteAddr)
		resp := map[string]string{
			"version": cfg.App.Version,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	http.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Handling /message request", "remote", r.RemoteAddr)
		resp := map[string]string{
			"message": cfg.Message.Hello,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	http.HandleFunc("/secrets", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Handling /secrets", "remote", r.RemoteAddr)
		json.NewEncoder(w).Encode(secrets)
	})

	port := "8080"
	logger.Info("Starting server", "port", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		logger.Error("Server failed", "error", err)
	}
}
