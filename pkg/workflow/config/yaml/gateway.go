package yaml

import (
	"github.com/cortezaproject/corteza-server/pkg/workflow"
	"github.com/davecgh/go-spew/spew"
	"gopkg.in/yaml.v3"
)

type (
	gwPaths []gwPath
	gwPath  struct {
		when string
		ref  string
		def  steps
		fn   resolver

		yNode *yaml.Node
	}

	pathAcceptor interface {
		workflow.Node
		AddPaths(...*workflow.GatewayPath) error
	}
)

const (
	kwGatewayJoin      keyword = "(gateway:join)"
	kwGatewayFork      keyword = "(gateway:fork)"
	kwGatewayInclusive keyword = "(gateway:inclusive)"
	kwGatewayExclusive keyword = "(gateway:exclusive)"

	kwWhen keyword = "(when)"
	kwThen keyword = "(then)"
	kwElse keyword = "(else)"
)

var _ = spew.Dump // @todo remove

func makeJoinGatewayResolver(inPaths *yaml.Node) (steps, resolver, error) {
	var (
		pp  = gwPaths{}
		err = inPaths.Decode(&pp)
	)

	if err != nil {
		return nil, nil, err
	}

	// check conditions; only last path should be without condition (ie: else)
	for _, p := range pp {
		if p.when != "" {
			return nil, nil, nodeErr(p.yNode, "unexpected path condition on fork gateway")
		}
	}

	return pp.steps(), makeGatewayResolver(workflow.JoinGateway(), pp...), nil
}

func makeForkGatewayResolver(outPaths *yaml.Node) (steps, resolver, error) {
	var (
		pp  = gwPaths{}
		err = outPaths.Decode(&pp)
	)

	// Since there's no condition needed for forks, we'll wrap each node in a sequence

	if err != nil {
		return nil, nil, err
	}

	// check conditions; only last path should be without condition (ie: else)
	for _, p := range pp {
		if p.when != "" {
			return nil, nil, nodeErr(p.yNode, "unexpected path condition on fork gateway")
		}
	}

	return pp.steps(), makeGatewayResolver(workflow.ForkGateway(), pp...), nil
}

func makeInclusiveGatewayResolver(outPaths *yaml.Node) (steps, resolver, error) {
	var (
		pp   = gwPaths{}
		err  = outPaths.Decode(&pp)
		last = len(pp) - 1
	)

	if err != nil {
		return nil, nil, err
	}

	// check conditions; only last path should be without condition (ie: else)
	for i, p := range pp {
		if p.when == "" && i != last {
			return nil, nil, nodeErr(p.yNode, "only last path can be without condition")
		}
	}

	return pp.steps(), makeGatewayResolver(workflow.InclGateway(), pp...), nil
}

func makeExclusiveGatewayResolver(outPaths *yaml.Node) (steps, resolver, error) {
	var (
		pp   = gwPaths{}
		err  = outPaths.Decode(&pp)
		last = len(pp) - 1
	)

	if err != nil {
		return nil, nil, err
	}

	// check conditions; only last path should be without condition (ie: else)
	for i, p := range pp {
		if p.when == "" && i != last {
			return nil, nil, nodeErr(p.yNode, "only last path can be without condition")
		}
	}

	return pp.steps(), makeGatewayResolver(workflow.ExclGateway(), pp...), nil
}

func makeGatewayResolver(gw pathAcceptor, pp ...gwPath) resolver {
	return func(n workflow.Node, ss steps) (workflow.Node, error) {
		var (
			err error
		)

		for _, p := range pp {
			if ss.nodeLookupByRef(p.ref) == nil {
				return nil, nodeErr(p.yNode, "gateway resolver: %w", ErrDepsMissing)
			}
		}

		for _, p := range pp {
			pn := ss.nodeLookupByRef(p.ref)
			if p.when == "" {
				err = gw.AddPaths(workflow.GwPath(pn))
			} else {
				err = gw.AddPaths(workflow.GwPathWithCondition(p.when, pn))
			}

			if err != nil {
				return nil, nodeErr(p.yNode, "could not add paths to exclusive gateway: %w", err)
			}
		}

		return gw, nil
	}
}

func (p *gwPath) UnmarshalYAML(n *yaml.Node) error {
	var (
		// proc processes path's (then), (else) directive and case when
		// next step is defined sequence's scalar value
		proc = func(n *yaml.Node) error {
			switch n.Kind {
			case yaml.ScalarNode:
				// string reference to next node or an end event
				var s = &step{}
				if handleEndEvent(n, nil, s) {
					p.def = steps{s}
				} else {
					p.ref = n.Value
				}
			case yaml.SequenceNode:
				// mapping or sequence
				if err := n.Decode(&p.def); err != nil {
					return nodeErr(n, "can not unmarshal path: %w", err)
				}
			}

			return nil
		}
	)

	p.yNode = n

	if isKind(n, yaml.ScalarNode) {
		// when path's next step is defined as scalar value
	}

	if !isKind(n, yaml.MappingNode) {
		// when not a scalar value, expect mapping
		return proc(n)
		//return nodeErr(n, "expecting mapping node for paths")
	}

	var (
		kw, err = Checker(n, rules{
			kwWhen: {secondary: []keyword{kwThen}},
			kwElse: {secondary: []keyword{}},
		})
	)

	if err != nil {
		return nodeErr(n, "gateway path check failed: %w", err)
	}

	switch true {
	case kw.has(kwWhen):
		p.when = kw.vNode(kwWhen).Value
		return proc(kw.vNode(kwThen))
	case kw.has(kwElse):
		return proc(kw.vNode(kwElse))
	default:
		//return proc(n)
		return nodeErr(n, "unknown gateway path configuration")

	}
}

// steps collects steps from all paths and returns them
//
// this is needed because sub-steps can be defined as part of the gateway path def
func (pp gwPaths) steps() steps {
	var ss = steps{}

	for _, p := range pp {
		ss = append(ss, p.def...)
	}

	return ss
}
