package yaml

import (
	"gopkg.in/yaml.v3"
	"strings"
)

type (
	configurator struct {
		kNodes map[keyword]*yaml.Node
		vNodes map[keyword]*yaml.Node
		ee     expressions
	}

	rules map[keyword]struct {
		exprAllowed bool
		secondary   []keyword
	}
)

// Makes new keyword index from node
func Checker(n *yaml.Node, c rules) (*configurator, error) {
	var (
		i = &configurator{
			kNodes: make(map[keyword]*yaml.Node),
			vNodes: make(map[keyword]*yaml.Node),
			ee:     expressions{},
		}

		// does processed node contain only configurator
		// or not if it contains expressions as well
		kwOnly = true
	)

	if !isKind(n, yaml.MappingNode) {
		if isKind(n, yaml.ScalarNode) && strings.HasPrefix(n.Value, "(") && strings.HasSuffix(n.Value, ")") {
			return nil, nodeErr(n, "expecting mapping node, found unexpected keyword %s", n.Value)
		}

		return nil, nodeErr(n, "expecting mapping node")
	}

	_ = iterator(n, func(k *yaml.Node, v *yaml.Node) error {
		kw := keyword(k.Value)
		if kw.valid() {
			i.kNodes[kw] = k
			i.vNodes[kw] = v
		} else {
			kwOnly = false
		}

		return nil
	})

	for kw, combo := range c {
		if !i.has(kw) {
			continue
		}

		if !combo.exprAllowed && !kwOnly {
			return nil, nodeErr(n, "unexpected expression used with %s", kw)
		}

		if extra := i.diff(append([]keyword{kw}, combo.secondary...)...); len(extra) > 0 {
			return nil, nodeErr(n, "unexpected keyword %s used with %s", strings.Join(extra, ", "), kw)
		}
	}

	if !kwOnly {
		if err := i.ee.UnmarshalYAML(n); err != nil {
			return nil, err
		}
	}

	return i, nil
}

// removes keyword and returns true if keyword existed
func (c configurator) has(keyword keyword) bool {
	_, exists := c.kNodes[keyword]
	return exists
}

func (c configurator) delete(keyword keyword) {
	delete(c.kNodes, keyword)
	delete(c.vNodes, keyword)
}

func (c configurator) vNode(k keyword) *yaml.Node {
	return c.vNodes[k]
}

func (c configurator) kNode(k keyword) *yaml.Node {
	return c.kNodes[k]
}

func (c configurator) diff(oo ...keyword) (extra []string) {
	var (
		kwm = make(map[keyword]bool)
	)

	for _, o := range oo {
		kwm[o] = true
	}

	// extra keywords
	for o := range c.kNodes {
		if !kwm[o] {
			extra = append(extra, string(o))
		}
	}

	return extra
}
