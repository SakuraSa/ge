package exe

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/SakuraSa/ge/src/concept"
)

type TestTask struct {
	f func(context.Context) error
}

func (t TestTask) Do(ctx context.Context) error {
	return t.f(ctx)
}

func T(f func(context.Context) error) concept.Task {
	return TestTask{f: f}
}

type TestKeyType string

const (
	testKey TestKeyType = "this_is_a_test_key"
)

type TestValue struct {
	Values []string
}

func (v *TestValue) String() string {
	return strings.Join(v.Values, ",")
}

type TestAOP struct {
	f func(concept.TaskFunc) concept.TaskFunc
}

func (a TestAOP) Apply(t concept.TaskFunc) concept.TaskFunc {
	return a.f(t)
}

func TestSerial(t *testing.T) {
	tests := []struct {
		name     string
		children []concept.Task
		ctx      context.Context
		checker  func(context.Context) error
		wantErr  bool
	}{
		{
			name:     "empty",
			children: nil,
			ctx:      context.Background(),
			checker: func(ctx context.Context) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "normal",
			ctx:  context.WithValue(context.Background(), testKey, &TestValue{}),
			children: []concept.Task{
				T(func(ctx context.Context) error {
					v := ctx.Value(testKey).(*TestValue)
					v.Values = append(v.Values, "1")
					return nil
				}),
				T(func(ctx context.Context) error {
					v := ctx.Value(testKey).(*TestValue)
					v.Values = append(v.Values, "2")
					return nil
				}),
				T(func(ctx context.Context) error {
					v := ctx.Value(testKey).(*TestValue)
					v.Values = append(v.Values, "3")
					return nil
				}),
			},
			checker: func(ctx context.Context) error {
				v := ctx.Value(testKey).(*TestValue)
				if v.String() != "1,2,3" {
					return fmt.Errorf("unexpected value: %s", v.String())
				}
				return nil
			},
			wantErr: false,
		},
		{
			name: "error",
			ctx:  context.WithValue(context.Background(), testKey, &TestValue{}),
			children: []concept.Task{
				T(func(ctx context.Context) error {
					return fmt.Errorf("error")
				}),
			},
			checker: func(ctx context.Context) error { return nil },
			wantErr: true,
		},
		{
			name: "panic",
			ctx:  context.WithValue(context.Background(), testKey, &TestValue{}),
			children: []concept.Task{
				T(func(ctx context.Context) error {
					panic("panic")
				}),
			},
			checker: func(ctx context.Context) error { return nil },
			wantErr: true,
		},
		{
			name: "aop",
			ctx: SetAOP(context.WithValue(context.Background(), testKey, &TestValue{}), concept.AOPs{&TestAOP{
				f: func(t concept.TaskFunc) concept.TaskFunc {
					return func(ctx context.Context) error {
						v := ctx.Value(testKey).(*TestValue)
						v.Values = append(v.Values, "1")
						return t(ctx)
					}
				},
			}}),
			children: []concept.Task{
				T(func(ctx context.Context) error {
					return nil
				}),
			},
			checker: func(ctx context.Context) error {
				v := ctx.Value(testKey).(*TestValue)
				if v.String() != "1" {
					return fmt.Errorf("unexpected value: %s", v.String())
				}
				return nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSerial(tt.children...)
			err := s.Do(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Serial.Do() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err := tt.checker(tt.ctx); err != nil {
				t.Errorf("Serial.Do() checker = %v", err)
			}
		})
	}

	t.Run("ctx done", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		s := NewSerial(T(func(ctx context.Context) error {
			time.Sleep(time.Millisecond * 100)
			return nil
		}))
		err := s.Do(ctx)
		if err != context.Canceled {
			t.Errorf("Serial.Do() error = %v, wantErr %v", err, context.Canceled)
		}
	})
}
