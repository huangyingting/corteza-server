package yaml

import (
	"fmt"
	"github.com/cortezaproject/corteza-server/pkg/workflow"
	"gopkg.in/yaml.v3"
)

type (
	expression struct {
		Target string
		Source string
	}

	expressions []*expression
)

func (ee *expressions) UnmarshalYAML(n *yaml.Node) error {
	if !isKind(n, yaml.MappingNode) {
		return nodeErr(n, "unexpected node kind, only mapping types supported for expressions")
	}

	var (
		aux = expressions{}
		exp *expression
		err = iterator(n, func(targetNode, exprNode *yaml.Node) (err error) {
			// skip all keywords
			if keyword(targetNode.Value).valid() {
				return nil
			}

			exp = &expression{Target: targetNode.Value}
			if err = exprNode.Decode(exp); err != nil {
				return err
			}

			aux = append(aux, exp)
			return nil
		})
	)

	if err != nil {
		return fmt.Errorf("could not unmarshal expressions: %w", err)
	}

	*ee = aux
	return nil
}

func (ee expressions) Cast() []*workflow.Expr {
	var (
		oo = workflow.Expressions()
	)

	for _, e := range ee {
		oo.Push(workflow.NewExpr(e.Target, e.Source))
	}

	return oo
}

// SetSource sets value from yaml node and quotes it if needed
//
// It tries to understand what kind of input it got from the quote presence & style:
//  - single quotes indicate expression
//  - double quotes indicate string
//  - unquoted values are decoded and if possible, converted to expression
func (e *expression) UnmarshalYAML(n *yaml.Node) error {
	if !isKind(n, yaml.ScalarNode) {
		return nodeErr(n, "unexpected node kind, only scalar types supported for expression")
	}

	if n.Style == yaml.SingleQuotedStyle {
		// Already in the format we need it
		e.Source = n.Value
		return nil
	}

	if n.Style == yaml.DoubleQuotedStyle {
		// Re-quote it
		e.Source = fmt.Sprintf("%q", n.Value)
		return nil
	}

	// Try to parse the rest
	var (
		auxInt   int64
		auxFloat float64
		auxBool  bool
	)

	switch true {
	case nil == n.Decode(&auxFloat):
		e.Source = fmt.Sprintf("%f", auxFloat)
	case nil == n.Decode(&auxInt):
		e.Source = fmt.Sprintf("%d", auxInt)
	case nil == n.Decode(&auxBool):
		e.Source = fmt.Sprintf("%t", auxBool)
	default:
		e.Source = fmt.Sprintf("\"%v\"", n.Value)
	}

	return nil
}
