package yaml

import (
	"bytes"
	"github.com/cortezaproject/corteza-server/pkg/workflow"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExpressions(t *testing.T) {
	var (
		tc = []struct {
			name string
			yaml string
			test func(*require.Assertions, workflow.Variables, error)
		}{
			{
				"strings",
				`
scope:
  s1: unquoted string
  s2: "double-quoted string"
  s3: '"single-double-quoted string"'
  s4: "esc \" foo"
`,
				func(req *require.Assertions, scope workflow.Variables, err error) {
					req.NoError(err)
					req.NotNil(scope)
					req.Equal("unquoted string", scope["s1"])
					req.Equal("esc \" foo", scope["s4"])
				},
			},
			{
				"integers",
				`
scope:
  i1: 42
  i2: '42'
  i3: "42"
`,
				func(req *require.Assertions, scope workflow.Variables, err error) {
					req.NoError(err)
					req.NotNil(scope)
					req.Equal(float64(42), scope["i1"])
					req.Equal("42", scope["i3"])
				},
			},
			{
				"floats",
				`
scope:
 f1: 4.2
 f2: '4.2'
 f3: "4.2"
`,
				func(req *require.Assertions, scope workflow.Variables, err error) {
					req.NoError(err)
					req.NotNil(scope)
					req.Equal(4.2, scope["f1"])
					req.Equal("4.2", scope["f3"])
				},
			},
			{
				"booleans",
				`
scope:
 b1: true
 b2: 'true'
 b3: "true"
`,
				func(req *require.Assertions, scope workflow.Variables, err error) {
					req.NoError(err)
					req.NotNil(scope)
					req.Equal(true, scope["b1"])
					req.Equal("true", scope["b3"])
				},
			},
		}
	)

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			_, scope, err := Load(bytes.NewBufferString(c.yaml))
			c.test(require.New(t), scope, err)
		})
	}

}
