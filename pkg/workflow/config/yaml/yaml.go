package yaml

import (
	"context"
	"fmt"
	"github.com/cortezaproject/corteza-server/pkg/workflow"
	"gopkg.in/yaml.v3"
	"io"
	"strings"
)

type (
	root struct {
		Scope expressions
		Steps steps
	}

	keyword string
)

func (k keyword) valid() bool {
	return strings.HasPrefix(string(k), "(") && strings.HasSuffix(string(k), ")")
}

func (k keyword) is(n *yaml.Node) bool {
	return n != nil && n.Value == string(k)
}

func Load(r io.Reader) (workflow.Node, workflow.Variables, error) {
	var (
		d   = yaml.NewDecoder(r)
		cfg = &root{}
		err = d.Decode(cfg)
	)

	if err != nil {
		return nil, nil, err
	}

	return cfg.Convert()
}

func (r *root) Convert() (start workflow.Node, scope workflow.Variables, err error) {
	var (
		expr = workflow.Expressions(r.Scope.Cast()...)
	)

	if err = expr.Init(); err != nil {
		return nil, nil, err
	}

	if scope, err = expr.Run(context.Background()); err != nil {
		return nil, nil, err
	}

	if start, err = r.Steps.resolve(); err != nil {
		return nil, nil, err
	}

	return
}

func nodeErr(n *yaml.Node, format string, aa ...interface{}) error {
	format += " (%d:%d)"
	aa = append(aa, n.Line, n.Column)
	return fmt.Errorf(format, aa...)
}

func iterator(n *yaml.Node, fn func(*yaml.Node, *yaml.Node) error) error {
	if isKind(n, yaml.SequenceNode) {
		var placeholder *yaml.Node
		for i := 0; i < len(n.Content); i++ {
			if err := fn(placeholder, n.Content[i]); err != nil {
				return err
			}
		}

		return nil
	}

	if isKind(n, yaml.MappingNode) {
		for i := 0; i < len(n.Content); i += 2 {
			if err := fn(n.Content[i], n.Content[i+1]); err != nil {
				return err
			}
		}

		return nil
	}

	return nodeErr(n, "iterator is expecting mapping or sequence node")
}

func isKind(n *yaml.Node, tt ...yaml.Kind) bool {
	if n != nil {
		for _, t := range tt {
			if t == n.Kind {
				return true
			}
		}
	}

	return false
}
