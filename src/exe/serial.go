package exe

import (
	"context"

	"github.com/SakuraSa/ge/src/concept"
)

var (
	_ concept.Task = Serial{}
)

// Serial is a task that executes its children in order.
type Serial struct {
	children []concept.Task
}

func (s Serial) Do(ctx context.Context) error {
	for _, child := range s.children {
		if err := child.Do(ctx); err != nil {
			return err
		}
	}
	return nil
}

func NewSerial(children ...concept.Task) Serial {
	return Serial{children: children}
}
