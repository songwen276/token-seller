package utils

import "log/slog"

type JobWrapper struct {
	name   string
	runner func() error
}

func WrapJob(name string, runner func() error) *JobWrapper {
	return &JobWrapper{
		name:   name,
		runner: runner,
	}
}

func (w *JobWrapper) Run() {
	if err := w.runner(); err != nil {
		slog.Error("exec job failed", "name", w.name, "err", err)
	}
}
