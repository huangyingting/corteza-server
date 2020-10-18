package workflow

import (
	"context"
	"fmt"
	"github.com/PaesslerAG/gval"
	"sync"
)

type (
	GatewayPath struct {
		expr string
		eval gval.Evaluable
		to   Node
	}

	joinGateway struct {
		nodeRef string

		// all parent nodes we'll wait to be executed before going to next node
		paths []Iterator
		index map[Node]bool
		next  Node

		l sync.Mutex

		scope map[Node]Variables
	}

	forkGateway struct {
		nodeRef string
		paths   Nodes
		index   map[Node]bool
	}

	inclGateway struct {
		nodeRef string
		paths   []*GatewayPath
	}

	exclGateway struct {
		nodeRef string
		paths   []*GatewayPath
	}
)

var (
	_ Joiner = &joinGateway{}
	_ Tester = &forkGateway{}
	_ Tester = &inclGateway{}
	_ Tester = &exclGateway{}
)

func GwPathWithCondition(expr string, to Node) *GatewayPath {
	return &GatewayPath{expr: expr, to: to}
}

func GwPath(to Node) *GatewayPath {
	return &GatewayPath{to: to}
}

func initGatewayPaths(paths ...*GatewayPath) ([]*GatewayPath, error) {
	var (
		err error
	)

	for _, p := range paths {
		if p.expr == "" {
			continue
		}

		if p.eval, err = gval.Full().NewEvaluable(p.expr); err != nil {
			return nil, fmt.Errorf("can not parse %s: %w", p.expr, err)
		}
	}

	return paths, err
}

func JoinGateway() *forkGateway { return &forkGateway{} }

func NewJoinGateway(paths ...*GatewayPath) (*joinGateway, error) {
	gw := &joinGateway{
		paths: make([]Iterator, 0),
		index: make(map[Node]bool),
		scope: make(map[Node]Variables),
	}

	return gw, gw.AddPaths(paths...)
}

func (gw joinGateway) NodeRef() string { return gw.nodeRef }
func (gw joinGateway) Next() Node      { return gw.next }
func (gw *joinGateway) SetNext(n Node) { gw.next = n }
func (gw *joinGateway) Paths() Nodes {
	var pp = make(Nodes, 0, len(gw.paths))
	for _, p := range gw.paths {
		pp = append(pp, p)
	}

	return pp
}

func (gw *joinGateway) AddPaths(paths ...*GatewayPath) error {
	for _, p := range paths {
		i, is := p.to.(Iterator)
		if !is {
			return fmt.Errorf("expecting iterator node")
		}

		if !gw.index[i] {
			gw.index[i] = true
			gw.paths = append(gw.paths, i)
			i.SetNext(gw)
		}
	}

	return nil
}

func (gw *joinGateway) Join(p Node, scope Variables) (Node, Variables, error) {
	gw.l.Lock()
	defer gw.l.Unlock()

	// Allow scope overriding (in case when parent is executed again)
	//
	// This covers scenario where we route workflow back to one
	// of the nodes that is then joined
	gw.scope[p] = scope

	if len(gw.scope) < len(gw.paths) {
		// Not all collected
		return nil, nil, nil
	}

	// All collected, merge scope from all paths in the defined order
	var out = Variables{}
	for _, p := range gw.paths {
		out = out.Merge(gw.scope[p])
	}

	return gw.next, out, nil
}

func ForkGateway() *forkGateway { return &forkGateway{} }

func NewForkGateway(paths ...*GatewayPath) (*forkGateway, error) {
	fg := ForkGateway()
	return fg, fg.AddPaths(paths...)
}

func (gw forkGateway) NodeRef() string                                    { return gw.nodeRef }
func (gw forkGateway) Paths() Nodes                                       { return gw.paths }
func (gw forkGateway) Test(_ context.Context, _ Variables) (Nodes, error) { return gw.paths, nil }

func (gw *forkGateway) AddPaths(paths ...*GatewayPath) error {
	for _, p := range paths {
		if !gw.index[p.to] {
			gw.index[p.to] = true
			gw.paths = append(gw.paths, p.to)
		}
	}

	return nil
}

func InclGateway() *inclGateway { return &inclGateway{} }

// multiple matches
func NewInclGateway(paths ...*GatewayPath) (*inclGateway, error) {
	var err error
	paths, err = initGatewayPaths(paths...)
	return &inclGateway{paths: paths}, err
}

func (gw inclGateway) NodeRef() string { return gw.nodeRef }
func (gw inclGateway) Paths() Nodes {
	var paths Nodes
	for _, p := range gw.paths {
		paths = append(paths, p.to)
	}
	return paths
}

func (gw *inclGateway) AddPaths(paths ...*GatewayPath) (err error) {
	paths, err = initGatewayPaths(paths...)
	if err != nil {
		return err
	}

	gw.paths = append(gw.paths, paths...)
	return nil
}

// Test returns nodes from all paths that have a matching condition
func (gw inclGateway) Test(ctx context.Context, scope Variables) (to Nodes, err error) {
	for _, p := range gw.paths {
		if result, err := p.eval.EvalBool(ctx, scope); err != nil {
			return nil, err
		} else if result {
			to = append(to, p.to)
		}
	}

	if len(to) == 0 {
		return nil, fmt.Errorf("did not match any of conditions")
	}

	return

}

func ExclGateway() *exclGateway { return &exclGateway{} }

// single match
func NewExclGateway(paths ...*GatewayPath) (gw *exclGateway, err error) {
	gw = &exclGateway{}
	return gw, gw.AddPaths(paths...)
}

func (gw exclGateway) NodeRef() string { return gw.nodeRef }
func (gw exclGateway) Paths() Nodes {
	var paths Nodes
	for _, p := range gw.paths {
		paths = append(paths, p.to)
	}
	return paths
}

func (gw *exclGateway) AddPaths(paths ...*GatewayPath) (err error) {
	paths, err = initGatewayPaths(paths...)
	if err != nil {
		return err
	}

	gw.paths = append(gw.paths, paths...)
	return nil
}

// Test returns first path with matching condition
func (gw exclGateway) Test(ctx context.Context, scope Variables) (to Nodes, err error) {
	for i, p := range gw.paths {
		if len(p.expr) == 0 && i == len(gw.paths)-1 {
			// empty & last; treat it as else
			return Nodes{p.to}, nil
		}

		if result, err := p.eval.EvalBool(ctx, scope); err != nil {
			return nil, err
		} else if result {
			return Nodes{p.to}, nil
		}
	}

	return nil, fmt.Errorf("did not match any of conditions")
}
