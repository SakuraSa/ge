package concept

import "context"

// Task is an interface that defines a task that can be executed.
type Task interface {
	Do(ctx context.Context) error
}

// TaskFunc is a function type that defines a task that can be executed.
type TaskFunc func(ctx context.Context) error
