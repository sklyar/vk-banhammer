package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/SevereCloud/vksdk/v2/api"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/sklyar/vk-banhammer/internal/config"
	"github.com/sklyar/vk-banhammer/internal/entity"
	"github.com/sklyar/vk-banhammer/internal/server"
	"github.com/sklyar/vk-banhammer/internal/service"
)

func main() {
	cfg, err := config.ParseConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	logger, err := newLogger(cfg.LoggerLever)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer logger.Sync() //nolint:errcheck

	heuristicRules, err := loadHeuristicRules(cfg.HeuristicsPath)
	if err != nil {
		logger.Fatal("failed to load heuristic rules", zap.Error(err))
	}
	log.Printf("heuristic rules: %+v", heuristicRules)
	if err := validateHeuristicRules(heuristicRules); err != nil {
		logger.Fatal("failed to validate heuristic rules", zap.Error(err))
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	vkClient := api.NewVK(cfg.APIToken)

	banhammerService := service.NewService(vkClient, heuristicRules)
	httpServer := server.NewServer(logger, cfg.HTTPAddr, banhammerService, cfg.CallbackConfirmationCode)

	if err := httpServer.ListenAndServe(ctx); err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}
}

func newLogger(cfgLevel string) (*zap.Logger, error) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfgLevel)); err != nil {
		return nil, fmt.Errorf("failed to unmarshal logger level: %w", err)
	}

	zapCfg := zap.NewDevelopmentConfig()
	zapCfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	zapCfg.Level = zap.NewAtomicLevelAt(level)

	logger, err := zapCfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	// listen SIGUSR1 signal to reconfigure logger
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGUSR1)

		for range c {
			if err := zapCfg.Level.UnmarshalText([]byte(cfgLevel)); err != nil {
				logger.Error("failed to unmarshal logger level", zap.Error(err))
				continue
			}

			if err := logger.Sync(); err != nil {
				logger.Error("failed to sync logger", zap.Error(err))
				continue
			}

			logger.Info("logger reconfigured")
		}
	}()

	return logger, nil
}

func loadHeuristicRules(path string) (entity.HeuristicRules, error) {
	var rules entity.HeuristicRules

	if _, err := toml.DecodeFile(path, &rules); err != nil {
		return entity.HeuristicRules{}, fmt.Errorf("failed to decode heuristics rules: %w", err)
	}

	return rules, nil
}

func validateHeuristicRules(rules entity.HeuristicRules) error {
	if len(rules.PersonNonGrata) == 0 {
		return fmt.Errorf("heuristic rules must contain at least one rule")
	}

	// birthday regexp without leading zeros in month and day. like "19.9.1921"
	birthdateRegexp := regexp.MustCompile(`^([1-9]|[12]\d|3[01]).([1-9]|1[012]).\d{4}$`)
	for _, rule := range rules.PersonNonGrata {
		if rule.Name != nil && *rule.Name == "" {
			return fmt.Errorf("empty name in person non grata rule")
		}
		if rule.BirthDate != nil {
			if !birthdateRegexp.MatchString(*rule.BirthDate) {
				return fmt.Errorf("invalid birthdate format in person non grata rule")
			}
			if _, err := time.Parse("2.1.2006", *rule.BirthDate); err != nil {
				return fmt.Errorf("invalid birthdate format in person non grata rule")
			}
		}
	}

	return nil
}
