package yaml

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestStep(t *testing.T) {
	var (
		tc = []struct {
			name string
			yaml string
			test func(*require.Assertions, *step, error)
		}{
			{
				"end event",
				`(end)`,
				func(req *require.Assertions, step *step, err error) {
					req.NoError(err)
					req.NotNil(step)
					req.NotNil(step.fn)
				},
			},
			{
				"end event with garbage",
				`(end): garbage`,
				func(req *require.Assertions, step *step, err error) {
					req.NoError(err)
					req.NotNil(step)
					req.NotNil(step.fn)
				},
			},
			{
				"end event as error with message",
				`(error): error message`,
				func(req *require.Assertions, step *step, err error) {
					req.NoError(err)
					req.NotNil(step)
					req.NotNil(step.fn)
				},
			},
			{
				"assignment",
				`{ a: 7, c: 5 }`,
				func(req *require.Assertions, step *step, err error) {
					req.NoError(err)
					req.NotNil(step)
					req.NotNil(step.fn)
				},
			},
			{
				"exclusive gateway",
				`(gateway:exclusive): [ { (when): true, (then): (end) }, { (else): (end) } ]`,
				func(req *require.Assertions, step *step, err error) {
					req.NoError(err)
					req.NotNil(step)
					req.NotNil(step.fn)
				},
			},
			{
				"exclusive gateway with embedded steps",
				`(gateway:exclusive): [ { (when): true, (then): [(end)] }, { (else): (end) } ]`,
				func(req *require.Assertions, step *step, err error) {
					req.NoError(err)
					req.NotNil(step)
					req.NotNil(step.fn)
				},
			},
			{
				"fork gateway",
				`(gateway:fork): [ (end), (end), (end) ]`,
				func(req *require.Assertions, step *step, err error) {
					req.NoError(err)
					req.NotNil(step)
					req.NotNil(step.fn)
				},
			},
			{
				"join gateway",
				`(gateway:join): [ path1, path2, path3 ]`,
				func(req *require.Assertions, step *step, err error) {
					req.NoError(err)
					req.NotNil(step)
					req.NotNil(step.fn)
				},
			},
		}
	)

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			s := &step{}
			c.test(require.New(t), s, yaml.NewDecoder(bytes.NewBufferString(c.yaml)).Decode(s))
		})
	}

}
