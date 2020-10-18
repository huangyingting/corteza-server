package workflow

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

type (
	testEndEvent struct{ scope Variables }
)

var (
	_ Finalizer = &testEndEvent{}
)

func NewTestEndEvent() *testEndEvent                                  { return &testEndEvent{} }
func (t *testEndEvent) Finalize(_ context.Context, s Variables) error { t.scope = s; return nil }
func (testEndEvent) NodeRef() string                                  { return "end-event" }

func MustMakeNode(n Node, err error) Node {
	if err != nil {
		panic(err)
	}
	return n
}

func TestRun_Set(t *testing.T) {
	var (
		req = require.New(t)
		ctx = context.Background()
		ee  = NewTestEndEvent()
		sa  = MustMakeNode(NewSetActivity(ee, NewExpr("foo", `"bar"`)))
	)

	{
		es, err := Workflow(ctx, sa, Variables{})
		req.NoError(err)
		req.Equal(es["foo"], "bar")
	}
}

func TestRun_Condition(t *testing.T) {
	var (
		req = require.New(t)
		ctx = context.Background()

		ee  = NewTestEndEvent()
		bar = MustMakeNode(NewSetActivity(ee, NewExpr("foo", `"bar"`)))
		baz = MustMakeNode(NewSetActivity(ee, NewExpr("foo", `"baz"`)))

		cnd = MustMakeNode(NewExclGateway(
			GwPathWithCondition("setFooToBar", bar),
			GwPath(baz),
		))
	)

	{
		es, err := Workflow(ctx, cnd, Variables{"setFooToBar": true})
		req.NoError(err)
		req.Equal(es["foo"], "bar")
	}
	{
		es, err := Workflow(ctx, cnd, Variables{"setFooToBar": false})
		req.NoError(err)
		req.Equal(es["foo"], "baz")
	}
}

func TestRun_Loop(t *testing.T) {
	var (
		req = require.New(t)
		ctx = context.Background()

		ee  = NewTestEndEvent()
		inc = MustMakeNode(NewSetActivity(nil, NewExpr("counter", `counter + 1`)))

		cnd = MustMakeNode(NewExclGateway(
			GwPathWithCondition("counter < 5", inc),
			GwPath(ee),
		))
	)

	inc.(*setActivity).SetNext(cnd)

	{
		es, err := Workflow(ctx, cnd, Variables{"counter": 0})
		req.NoError(err)
		req.Equal(float64(5), es["counter"])
	}
}

func TestRun_Join(t *testing.T) {
	var (
		req = require.New(t)
		ctx = context.Background()

		ee  = NewTestEndEvent()
		foo = MustMakeNode(NewSetActivity(nil, NewExpr("foo", `"set"`), NewExpr("count", "count + 1"), NewExpr("order", "1")))
		bar = MustMakeNode(NewSetActivity(nil, NewExpr("bar", `"set"`), NewExpr("count", "count + 1"), NewExpr("order", "2")))
		baz = MustMakeNode(NewSetActivity(nil, NewExpr("baz", `"set"`), NewExpr("count", "count + 1"), NewExpr("order", "3")))

		fork = MustMakeNode(NewForkGateway(GwPath(foo), GwPath(bar), GwPath(baz)))
		join = MustMakeNode(NewJoinGateway(GwPath(foo), GwPath(bar), GwPath(baz)))
	)

	join.(*joinGateway).SetNext(ee)

	{
		es, err := Workflow(ctx, fork, Variables{
			"setFooToBar": true,
			"count":       0,
			"order":       0,
		})

		req.NoError(err)

		// test if all paths are executed
		// all there should be set
		req.Equal(es["foo"], "set")
		req.Equal(es["bar"], "set")
		req.Equal(es["baz"], "set")

		// tests if paths are isolated;
		// expecting 1, all setters start with scope where count==0
		req.Equal(es["count"], float64(1))
		// tests if scope from paths merged in proper order
		// expecting 3 - value from the node last added to the join
		req.Equal(es["order"], float64(3))
	}
}
