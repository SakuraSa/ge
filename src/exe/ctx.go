package exe

import (
	"context"

	"github.com/SakuraSa/ge/src/concept"
)

type AOPKeyType string

const (
	AOPKey AOPKeyType = "AOP"
)

var (
	nilAOPs concept.AOPs = nil
)

// GetAOP returns the AOP in the context.
func GetAOP(ctx context.Context) concept.AOP {
	aop, ok := ctx.Value(AOPKey).(concept.AOP)
	if !ok {
		return nilAOPs
	}
	return aop
}

// SetAOP sets the AOP in the context.
func SetAOP(ctx context.Context, aop concept.AOP) context.Context {
	return context.WithValue(ctx, AOPKey, aop)
}
