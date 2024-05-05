package exe

import (
	"context"
	"fmt"
	"testing"

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
}
