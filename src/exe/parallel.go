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
	var aops = GetAOP(ctx)
	for _, child := range p.children {
		f := aops.Apply(child.Do)
		go func(child concept.Task) {
			defer func() {
				if e := recover(); e != nil {
					errs <- fmt.Errorf("task %s panic: %v\n%s", child, e, debug.Stack())
				}
			}()
			errs <- f(ctx)
		}(child)
	}
	for range p.children {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errs:
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func NewParallel(children ...concept.Task) Parallel {
	return Parallel{children: children}
}
