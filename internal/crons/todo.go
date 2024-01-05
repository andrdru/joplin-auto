package crons

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type (
	todoManager interface {
		UpdateTodo(ctx context.Context) error
	}

	JobTodo struct {
		manager todoManager
	}
)

func NewJobTodo(manager todoManager) *JobTodo {
	return &JobTodo{
		manager: manager,
	}
}

func (c *JobTodo) Run() {
	logger := slog.Default().
		With("cron", "todo_cron").
		With("job_id", uuid.New().String())

	timeStart := time.Now()

	logger.Info("cron status start")

	defer func() {
		logger.Info(fmt.Sprintf("cron status finish in %s", time.Since(timeStart).String()))
	}()

	err := c.manager.UpdateTodo(context.Background())
	if err != nil {
		logger.Error("cron error", slog.String("err", err.Error()))
	}
}
