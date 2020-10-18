package yaml

import (
	"bytes"
	"context"
	"github.com/cortezaproject/corteza-server/pkg/workflow"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSteps(t *testing.T) {
	var (
		tc = []struct {
			name string
			yaml string
			test func(*testing.T, workflow.Node, workflow.Variables, error)
		}{
			{
				"one setter in map",
				`steps: { setter: { foo: "bar" }, (end) }`,
				func(t *testing.T, start workflow.Node, scope workflow.Variables, err error) {
					req := require.New(t)
					req.NoError(err)
					req.NotNil(scope)
					req.NotNil(start)
					workflow.Plot(start)
					scope, err = workflow.Workflow(context.Background(), start, scope)
					req.NoError(err)
					req.NotNil(scope)
					req.Contains(scope, "foo")
					req.Equal("bar", scope["foo"])
				},
			},
			{
				"one setter in sequence",
				`steps: [ { foo: "bar", (ref): setter }, (end) ]`,
				func(t *testing.T, start workflow.Node, scope workflow.Variables, err error) {
					req := require.New(t)
					req.NoError(err)
					req.NotNil(scope)
					req.NotNil(start)
					scope, err = workflow.Workflow(context.Background(), start, scope)
					req.NoError(err)
					req.NotNil(scope)
					req.NotNil(scope)
					req.Contains(scope, "foo")
					req.Equal("bar", scope["foo"])
				},
			},
			{
				"simple gateway",
				`
steps:
  gw:
    (gateway:exclusive):
    - (when): this
      (then): (end)
    - (else): (end)
`,

				func(t *testing.T, start workflow.Node, scope workflow.Variables, err error) {
					req := require.New(t)
					req.NoError(err)
					req.NotNil(scope)
					req.NotNil(start)
				},
			},
			{
				"gateway with embedded then steps",
				`
steps:
  gw:
    (gateway:exclusive):
    - (when): this
      (then): { setter: { foo: "bar" }, (end) }
    - (else): 
        (error): "Poo"
`,

				func(t *testing.T, start workflow.Node, scope workflow.Variables, err error) {
					req := require.New(t)
					req.NoError(err)
					req.NotNil(scope)
					req.NotNil(start)
				},
			},
			{
				"two-stepper",
				`
steps:
  setter1: { foo: 1 }
  setter2: { foo: 2 }
  (end):
`,

				func(t *testing.T, start workflow.Node, scope workflow.Variables, err error) {
					req := require.New(t)
					req.NoError(err)
					req.NotNil(scope)
					req.NotNil(start)
				},
			},
			{
				"step-container-step",
				`
steps:
  setter1: { foo: 1 }
  container: { (sub): [ { (ref): setter2, foo: 2 } ] }
  (end):
`,

				func(t *testing.T, start workflow.Node, scope workflow.Variables, err error) {
					req := require.New(t)
					req.NoError(err)
					req.NotNil(scope)
					req.NotNil(start)
				},
			},
			{
				"nested steps under gw",
				`
steps:
  check:
    (gateway:exclusive):
    - if: foo
      next: { setter1: { foo: 1 }, (end) }
    - next: { setter2: { foo: 2 }, setter3: { foo: 3 }, (end) }
`,

				func(t *testing.T, start workflow.Node, scope workflow.Variables, err error) {
					req := require.New(t)
					req.NoError(err)
					req.NotNil(scope)
					req.NotNil(start)
				},
			},
			{
				"loop",
				`
steps:
  setter1: { foo: foo + 1 }
  while:
    (gateway:exclusive):
    - (when): foo < 5
      (then): setter1
    - (then): (end)
`,

				func(t *testing.T, start workflow.Node, scope workflow.Variables, err error) {
					req := require.New(t)
					req.NoError(err)
					req.NotNil(scope)
					req.NotNil(start)
				},
			},
			{
				"fork-join",
				`
steps:
  fork:
    (gateway:fork):
    - { setter1: { s1: 1, }, (next): join }
    - { setter2: { s2: 2, }, (next): join }
    - { setter3: { s3: 3, }, (next): join }
  join:
    (gateway:join):
    - setter1
    - setter2
    - setter3
  (end):
`,

				func(t *testing.T, start workflow.Node, scope workflow.Variables, err error) {
					req := require.New(t)
					req.NoError(err)
					req.NotNil(scope)
					req.NotNil(start)
				},
			},
		}
	)

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			start, scope, err := Load(bytes.NewBufferString(c.yaml))
			c.test(t, start, scope, err)
		})
	}

}
