package yaml

import (
	"fmt"
	"github.com/cortezaproject/corteza-server/pkg/workflow"
	"gopkg.in/yaml.v3"
)

type (
	resolver func(workflow.Node, steps) (workflow.Node, error)
	step     struct {
		wfn   workflow.Node
		ref   string
		next  string
		fn    resolver
		steps steps

		// points to key node from which this step was created
		ykNode *yaml.Node

		// points to value node from which this step was created
		yvNode *yaml.Node
	}
)

const (
	// Flow configurator
	kwEnd   keyword = "(end)"
	kwError keyword = "(error)"
	kwRef   keyword = "(ref)"
	kwSub   keyword = "(sub)"
	kwCatch keyword = "(catch)"
	kwNext  keyword = "(next)"

	kwExec keyword = "(exec)"
	kwArgs keyword = "(args)"
)

func (s *step) UnmarshalYAML(n *yaml.Node) error {
	// step source
	s.yvNode = n

	//println("handling end event step unmarshaler")
	if handleEndEvent(n, nil, s) {
		return nil
	}

	var (
		cfg, err = Checker(n, rules{
			kwEnd:              {false, []keyword{}},
			kwError:            {false, []keyword{}},
			kwExec:             {true, []keyword{kwRef, kwArgs, kwNext}},
			kwGatewayJoin:      {false, []keyword{kwRef, kwNext}},
			kwGatewayFork:      {false, []keyword{kwRef}},
			kwGatewayInclusive: {false, []keyword{kwRef}},
			kwGatewayExclusive: {false, []keyword{kwRef, kwNext}},
			kwSub:              {false, []keyword{kwRef}},
		})
	)

	if err != nil {
		return err
	}

	switch true {
	case cfg.has(kwEnd):
		println("handling end event from step unmarshaler (end)")
		handleEndEvent(cfg.kNode(kwEnd), cfg.vNode(kwEnd), s)
		return nil

	case cfg.has(kwError):
		println("handling end event from step unmarshaler (error)")
		handleEndEvent(cfg.kNode(kwError), cfg.vNode(kwError), s)
		return nil
	}

	if ref := cfg.vNode(kwRef); ref != nil {
		if !isKind(ref, yaml.ScalarNode) {
			return nodeErr(ref, "reference node (ref) value should be scalar (non map, non sequence)")
		}

		if s.ref != "" {
			return nodeErr(ref, "reference ID (ref) already set with mapping node")
		}

		s.ref = ref.Value

		// remove keyword to simplify required/optional check in the switch below
		cfg.delete(kwRef)
	}

	//if s.ref == "" {
	//	return nodeErr(n, "missing workflow step reference (ref)")
	//}

	switch true {
	case cfg.has(kwExec):
		var execFn = cfg.vNode(kwExec).Value
		// @todo figure out if exec's fn is supported
		_ = execFn

		var rawArgs = expressions{}
		if err = cfg.vNode(kwArgs).Decode(&rawArgs); err != nil {
			return err
		}

		var args = workflow.Expressions(rawArgs.Cast()...)
		if err = args.Init(); err != nil {
			return err
		}

		// @todo check collected args against params for exec
		// @todo add resolver

	case cfg.has(kwSub):
		if err = cfg.vNode(kwSub).Decode(&s.steps); err != nil {
			return err
		}

		// Next step is the first one in the sub
		s.next = s.steps[0].ref

	case cfg.has(kwGatewayJoin):
		if s.steps, s.fn, err = makeJoinGatewayResolver(cfg.vNode(kwGatewayJoin)); err != nil {
			return err
		}

	case cfg.has(kwGatewayFork):
		if s.steps, s.fn, err = makeForkGatewayResolver(cfg.vNode(kwGatewayFork)); err != nil {
			return err
		}

	case cfg.has(kwGatewayInclusive):
		if s.steps, s.fn, err = makeInclusiveGatewayResolver(cfg.vNode(kwGatewayInclusive)); err != nil {
			return err
		}

	case cfg.has(kwGatewayExclusive):
		if s.steps, s.fn, err = makeExclusiveGatewayResolver(cfg.vNode(kwGatewayExclusive)); err != nil {
			return err
		}
	default:
		s.fn, err = makeSetResolver(n)
	}

	if next := cfg.vNode(kwNext); next != nil {
		if !isKind(next, yaml.ScalarNode) {
			return nodeErr(n, "next node (next) value should be scalar (non map, non sequence)")
		}

		s.next = next.Value
	}

	return nil
}

// handleEndEvent handles end node and returns true if handled
func handleEndEvent(k, v *yaml.Node, s *step) bool {
	_ = fmt.Sprintf

	if k != nil && !isKind(k, yaml.ScalarNode) {
		//println("not nil & not scalar")
		return false
	}

	//fmt.Printf("handling end event\n    [K] => %+v\n    [V] => %+v\n", k, v)

	switch true {

	case kwEnd.is(k):
		s.fn = func(node workflow.Node, s steps) (workflow.Node, error) {
			return workflow.EndEvent(), nil
		}
		return true

	case kwError.is(k):
		var message string
		if isKind(v, yaml.ScalarNode) {
			message = v.Value
		}

		s.fn = func(node workflow.Node, s steps) (workflow.Node, error) {
			return workflow.ErrorEvent(message), nil
		}
		return true

	}

	return false
}

func makeSetResolver(set *yaml.Node) (resolver, error) {
	var (
		expr = &expressions{}
		err  = expr.UnmarshalYAML(set)
	)

	if err != nil {
		return nil, err
	}

	return func(n workflow.Node, ss steps) (workflow.Node, error) {
		if n == nil {
			return nil, ErrDepsMissing
		}

		return workflow.NewSetActivity(n, expr.Cast()...)
	}, nil
}
