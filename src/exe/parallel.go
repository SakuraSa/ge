package exe

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/SakuraSa/ge/src/concept"
)

var (
	_ concept.Task = Parallel{}
)

// Parallel is a task that executes its children concurrently.
type Parallel struct {
	children []concept.Task
}

func (p Parallel) Do(ctx context.Context) error {
	errs := make(chan error, len(p.children))
	for _, child := range p.children {
		go func(child concept.Task) {
			defer func() {
				if e := recover(); e != nil {
					errs <- fmt.Errorf("task %s panic: %v\n%s", child, e, debug.Stack())
				}
			}()
			errs <- child.Do(ctx)
		}(child)
	}
	for range p.children {
		if err := <-errs; err != nil {
			return err
		}
	}
	return nil
}

func NewParallel(children ...concept.Task) Parallel {
	return Parallel{children: children}
}
