package workflow

import (
	"context"
	"fmt"
	"github.com/PaesslerAG/gval"
	"strings"
)

type (
	Expr struct {
		// target for the result of the evaluated expression
		target string

		// expression
		source string

		// expression, ready to be executed
		eval gval.Evaluable
	}

	expressions []*Expr

	setActivity struct {
		next Node

		// List of expressions that will be evaluated
		// with given scope
		ee expressions
	}
)

var (
	_ Executor = &setActivity{}
)

func NewExpr(dst, source string) *Expr {
	return &Expr{target: dst, source: source}
}

func NewSetActivity(next Node, ee ...*Expr) (*setActivity, error) {
	var (
		set = &setActivity{
			next: next,
			ee:   ee,
		}
	)

	if err := set.ee.Init(); err != nil {
		return nil, err
	}

	return set, nil
}

func (s setActivity) NodeRef() string { return "setter" }
func (s setActivity) Next() Node      { return s.next }
func (s *setActivity) SetNext(n Node) { s.next = n }

func (s setActivity) Exec(ctx context.Context, params Variables) (Variables, error) {
	return s.ee.Run(ctx, params)
}

func Expressions(ee ...*Expr) expressions {
	return ee
}

func (ee *expressions) Push(new ...*Expr) {
	*ee = append(*ee, new...)
}

func (ee expressions) Init() error {
	var (
		err  error
		lang = gval.Full()
	)

	for _, e := range ee {
		if e.eval, err = lang.NewEvaluable(e.source); err != nil {
			return fmt.Errorf("can no parse %s for %s: %w", e.source, e.target, err)
		}
	}

	return nil
}

func (ee expressions) Run(ctx context.Context, params ...Variables) (Variables, error) {
	var (
		err error

		// Create scope from params
		//
		// We'll use it for evaluation of preconfigured expressions on the setter
		// and as a container for each value that comes out of that evaluation
		scope = Variables{}.Merge(params...)
	)

	for _, e := range ee {
		if strings.Contains(e.target, ".") {
			// handle property setting
			return nil, fmt.Errorf("no support for prop setting ATM")
		}

		if scope[e.target], err = e.eval(ctx, scope); err != nil {
			return nil, fmt.Errorf("could not evaluate %s for %s: %w", e.source, e.target, err)
		}
	}

	return scope, nil
}
