package exe

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/SakuraSa/ge/src/concept"
	"github.com/SakuraSa/ge/src/util/gslice"
)

var (
	_ concept.Task = DAG{}

	ErrCycle     = fmt.Errorf("cycle detected in DAG")
	ErrDuplicate = fmt.Errorf("duplicate node in DAG")
)

type DAG struct {
	nodes []concept.Task
	edges [][]int
}

func (d DAG) Do(ctx context.Context) error {
	if len(d.nodes) == 0 {
		return nil
	}

	type Result struct {
		err   error
		index int
	}
	var (
		closed    = 0
		conds     = make([]int, 0, len(d.nodes))
		onReady   = make(chan int, len(d.nodes))
		onFinnish = make(chan Result, len(d.nodes))
	)

	conds = d.getConds()

	for i, cond := range conds {
		if cond == 0 {
			onReady <- i
		}
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case index := <-onReady:
			go func() {
				var (
					err   error
					child = d.nodes[index]
				)
				defer func() {
					if e := recover(); e != nil {
						err = fmt.Errorf("task %s panic: %v\n%s", child, e, debug.Stack())
					}
					onFinnish <- Result{err, index}
				}()
				err = child.Do(ctx)
			}()
		case result := <-onFinnish:
			if result.err != nil {
				return result.err
			}
			closed++
			for _, index := range d.edges[result.index] {
				conds[index]--
				if conds[index] == 0 {
					onReady <- index
				}
			}
			if closed == len(d.nodes) {
				return nil
			}
		}
	}
}

func (d DAG) getConds() []int {
	conds := make([]int, len(d.nodes))
	for _, edges := range d.edges {
		for _, edge := range edges {
			conds[edge]++
		}
	}
	return conds
}

type DAGBuilder struct {
	nodeMap map[string]concept.Task
	edgeMap map[string][]string
}

func NewDAGBuilder() *DAGBuilder {
	return &DAGBuilder{
		nodeMap: make(map[string]concept.Task),
		edgeMap: make(map[string][]string),
	}
}

func (d *DAGBuilder) AddNode(name string, task concept.Task, deps ...string) {
	d.nodeMap[name] = task
	d.edgeMap[name] = deps
}

func (d *DAGBuilder) Build() (DAG, error) {
	nodes := make([]concept.Task, 0, len(d.nodeMap))
	edges := make([][]int, len(d.nodeMap))
	nodeIndex := make(map[string]int)

	for name := range d.nodeMap {
		nodeIndex[name] = len(nodes)
		nodes = append(nodes, d.nodeMap[name])
	}

	for name, deps := range d.edgeMap {
		index := nodeIndex[name]
		for _, dep := range deps {
			edges[index] = append(edges[index], nodeIndex[dep])
		}
	}

	dag := DAG{
		nodes: nodes,
		edges: edges,
	}

	for _, f := range []func(DAG) error{checkCycle, checkDuplicate} {
		if err := f(dag); err != nil {
			return dag, err
		}
	}

	return dag, nil
}

func checkCycle(d DAG) error {
	var path []int
	var visited = make([]int, len(d.nodes))
	for i := range d.nodes {
		if visited[i] == 0 {
			if err := dfs(d, i, &path, visited); err != nil {
				return err
			}
		}

	}
	return nil
}

func dfs(d DAG, index int, path *[]int, visited []int) error {
	visited[index] = 1
	*path = append(*path, index)
	for _, edge := range d.edges[index] {
		if visited[edge] == 1 {
			return ErrCycle
		}
		if visited[edge] == 0 {
			if err := dfs(d, edge, path, visited); err != nil {
				return err
			}
		}
	}
	visited[index] = 2
	*path = (*path)[:len(*path)-1]
	return nil
}

func checkDuplicate(d DAG) error {
	for _, edges := range d.edges {
		if !gslice.IsUniqe(edges) {
			return ErrDuplicate
		}
	}
	return nil
}
