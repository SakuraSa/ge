package exe

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/SakuraSa/ge/src/concept"
)

func Test_checkDuplicate(t *testing.T) {
	type args struct {
		d DAG
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "empty",
			args: args{
				d: DAG{},
			},
			wantErr: false,
		},
		{
			name: "normal",
			args: args{
				d: DAG{
					nodes: []concept.Task{nil, nil},
					edges: [][]int{{1}, {}},
				},
			},
			wantErr: false,
		},
		{
			name: "duplicate",
			args: args{
				d: DAG{
					nodes: []concept.Task{nil, nil},
					edges: [][]int{{1, 1}, {}},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkDuplicate(tt.args.d); (err != nil) != tt.wantErr {
				t.Errorf("checkDuplicate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_checkCycle(t *testing.T) {
	type args struct {
		d DAG
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "empty",
			args: args{
				d: DAG{},
			},
			wantErr: false,
		},
		{
			name: "normal",
			args: args{
				d: DAG{
					nodes: []concept.Task{nil, nil, nil, nil},
					edges: [][]int{{}, {0}, {0, 1}, {2}},
				},
			},
			wantErr: false,
		},
		{
			name: "cycle",
			args: args{
				d: DAG{
					nodes: []concept.Task{nil, nil, nil, nil},
					edges: [][]int{{3}, {0}, {0, 1}, {2}},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkCycle(tt.args.d); (err != nil) != tt.wantErr {
				t.Errorf("checkCycle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_checkUnknownDep(t *testing.T) {
	type args struct {
		d DAG
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "empty",
			args: args{
				d: DAG{},
			},
			wantErr: false,
		},
		{
			name: "normal",
			args: args{
				d: DAG{
					nodes: []concept.Task{nil, nil},
					edges: [][]int{{1}, {}},
				},
			},
			wantErr: false,
		},
		{
			name: "unkown",
			args: args{
				d: DAG{
					nodes: []concept.Task{nil, nil},
					edges: [][]int{{1}, {9}},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkUnknownDep(tt.args.d); (err != nil) != tt.wantErr {
				t.Errorf("checkUnknownDep() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDAG(t *testing.T) {
	type buildArgs struct {
		task concept.Task
		name string
		deps []string
	}
	tests := []struct {
		name     string
		children []buildArgs
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
			children: []buildArgs{
				{
					task: T(func(ctx context.Context) error {
						v := ctx.Value(testKey).(*TestValue)
						v.Values = append(v.Values, "1")
						return nil
					}),
					name: "1",
					deps: nil,
				},
				{
					task: T(func(ctx context.Context) error {
						v := ctx.Value(testKey).(*TestValue)
						v.Values = append(v.Values, "2")
						return nil
					}),
					name: "2",
					deps: []string{"1"},
				},
				{
					task: T(func(ctx context.Context) error {
						v := ctx.Value(testKey).(*TestValue)
						v.Values = append(v.Values, "3")
						return nil
					}),
					name: "3",
					deps: []string{"2"},
				},
				{
					task: T(func(ctx context.Context) error {
						v := ctx.Value(testKey).(*TestValue)
						v.Values = append(v.Values, "4")
						return nil
					}),
					name: "4",
					deps: []string{"1", "3"},
				},
			},
			checker: func(ctx context.Context) error {
				v := ctx.Value(testKey).(*TestValue)
				if v.String() != "4,3,2,1" {
					return fmt.Errorf("unexpected value: %s", v.String())
				}
				return nil
			},
			wantErr: false,
		},
		{
			name: "wildcard",
			ctx:  context.WithValue(context.Background(), testKey, &TestValue{}),
			children: []buildArgs{
				{
					task: T(func(ctx context.Context) error {
						v := ctx.Value(testKey).(*TestValue)
						v.Values = append(v.Values, "1")
						return nil
					}),
					name: "task-1",
					deps: nil,
				},
				{
					task: T(func(ctx context.Context) error {
						v := ctx.Value(testKey).(*TestValue)
						v.Values = append(v.Values, "2")
						return nil
					}),
					name: "task-2",
					deps: []string{"task-1"},
				},
				{
					task: T(func(ctx context.Context) error {
						v := ctx.Value(testKey).(*TestValue)
						v.Values = append(v.Values, "3")
						return nil
					}),
					name: "task-3",
					deps: []string{"task-2"},
				},
				{
					task: T(func(ctx context.Context) error {
						v := ctx.Value(testKey).(*TestValue)
						v.Values = append(v.Values, "4")
						return nil
					}),
					name: "task-4",
					deps: []string{"task-[13]"},
				},
			},
			checker: func(ctx context.Context) error {
				v := ctx.Value(testKey).(*TestValue)
				if v.String() != "4,3,2,1" {
					return fmt.Errorf("unexpected value: %s", v.String())
				}
				return nil
			},
			wantErr: false,
		},
		{
			name: "error",
			ctx:  context.WithValue(context.Background(), testKey, &TestValue{}),
			children: []buildArgs{
				{
					task: T(func(ctx context.Context) error {
						return fmt.Errorf("error")
					}),
					name: "error",
					deps: nil,
				},
			},
			checker: func(ctx context.Context) error { return nil },
			wantErr: true,
		},
		{
			name: "panic",
			ctx:  context.WithValue(context.Background(), testKey, &TestValue{}),
			children: []buildArgs{
				{
					task: T(func(ctx context.Context) error {
						panic("panic")
					}),
					name: "panic",
					deps: nil,
				},
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
			children: []buildArgs{
				{
					task: T(func(ctx context.Context) error {
						return nil
					}),
					name: "do-nothing",
					deps: nil,
				},
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
			b := NewDAGBuilder()
			for _, c := range tt.children {
				b.AddNode(c.name, c.task, c.deps...)
			}
			s, err := b.Build()
			if err != nil {
				t.Errorf("DAG.Build() error = %v", err)
				return
			}
			err = s.Do(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("DAG.Do() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err := tt.checker(tt.ctx); err != nil {
				t.Errorf("DAG.Do() checker = %v", err)
			}
		})
	}

	t.Run("ctx done", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		b := NewDAGBuilder()
		b.AddNode("sleep", T(func(ctx context.Context) error {
			time.Sleep(time.Millisecond * 100)
			return nil
		}))
		s, err := b.Build()
		if err != nil {
			t.Errorf("DAG.Build() error = %v", err)
			return
		}
		err = s.Do(ctx)
		if err != context.Canceled {
			t.Errorf("DAG.Do() error = %v, wantErr %v", err, context.Canceled)
		}
	})

	t.Run("build err:cycle", func(t *testing.T) {
		b := NewDAGBuilder()
		b.AddNode("1", T(func(ctx context.Context) error {
			return nil
		}), "1")
		_, err := b.Build()
		if err == nil {
			t.Errorf("DAG.Build() error = %v, wantErr %v", err, true)
		}
	})

	t.Run("build err:bad wildcard", func(t *testing.T) {
		b := NewDAGBuilder()
		b.AddNode("1", T(func(ctx context.Context) error {
			return nil
		}), "(")
		_, err := b.Build()
		if err == nil {
			t.Errorf("DAG.Build() error = %v, wantErr %v", err, true)
		}
	})
}
