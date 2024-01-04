package app

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/andrdru/go-template/graceful"
	"github.com/andrdru/joplin-auto/internal/configs"
	"github.com/andrdru/joplin-auto/internal/crons"
	"github.com/andrdru/joplin-auto/internal/managers"
	"github.com/andrdru/joplin-auto/joplin_provider"
	"github.com/andrdru/joplin-auto/joplin_provider/s3client"
	"github.com/robfig/cron/v3"
)

type (
	bootstrap struct {
		cron        *cron.Cron
		todoManager *managers.Todo

		closers []graceful.Closer
	}
)

func Run(logger *slog.Logger, configPath string) (code int) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Info("app staring")
	defer logger.Info("app finished")

	conf, err := configs.NewConfig(configPath)
	if err != nil {
		logger.Error("init config", slog.String("error", err.Error()))
		return 1
	}

	boot := &bootstrap{
		cron: cron.New(cron.WithSeconds()),
	}

	defer func() {
		ctxCloser, cancelCloser := context.WithTimeout(context.Background(), 15*time.Second)
		defer func() {
			cancelCloser()
		}()

		graceful.Stop(ctxCloser, logger, boot.closers)
	}()

	err = initApp(boot, conf)
	if err != nil {
		logger.Error("init app", slog.String("error", err.Error()))
		return 1
	}

	initCronSchedule(boot)

	go boot.cron.Run()
	boot.closers = append(boot.closers,
		func(ctx context.Context) (description string, err error) {
			boot.cron.Stop()
			return "cron", nil
		})

	logger.Info("app started successfully")
	<-ctx.Done()

	return 0
}

func initCronSchedule(boot *bootstrap) {
	jobCreate := crons.NewJobTodo(boot.todoManager)

	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	schedule, _ := parser.Parse("0 * * * * *") // every 0th second of a minute

	boot.cron.Schedule(schedule, jobCreate)
}

func initApp(boot *bootstrap, conf configs.Config) (err error) {
	s3, err := s3client.NewS3Client(&s3client.Config{
		Region:   conf.S3.Region,
		Endpoint: conf.S3.Host,
		Key:      conf.S3.Key,
		Secret:   conf.S3.Secret,
		Bucket:   conf.S3.Bucket,
	})

	if err != nil {
		return fmt.Errorf("s3 client: %w", err)
	}

	s3provider := joplin_provider.NewS3(s3)

	boot.todoManager = managers.NewTodo(s3provider, conf.AppID, conf.NoteID, conf.ParentID)

	return nil
}
