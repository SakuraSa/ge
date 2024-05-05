package exe

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/SakuraSa/ge/src/concept"
)

var (
	_ concept.Task = Serial{}
)

// Serial is a task that executes its children in order.
type Serial struct {
	children []concept.Task
}

func (s Serial) Do(ctx context.Context) (err error) {
	var current concept.Task
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("task %s panic: %v\n%s", current, e, debug.Stack())
		}
	}()
	for _, current = range s.children {
		if err = current.Do(ctx); err != nil {
			return
		}
	}
	return nil
}

func NewSerial(children ...concept.Task) Serial {
	return Serial{children: children}
}
