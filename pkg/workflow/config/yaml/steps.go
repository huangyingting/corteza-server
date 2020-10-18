package yaml

import (
	"fmt"
	"github.com/cortezaproject/corteza-server/pkg/workflow"
	"github.com/davecgh/go-spew/spew"
	"go.uber.org/atomic"
	"gopkg.in/yaml.v3"
	"strings"
)

type (
	steps []*step
)

var (
	stepCounter    atomic.Uint64
	ErrDepsMissing = fmt.Errorf("waiting for udeps to be resolved")
)

func (ss steps) resolve() (workflow.Node, error) {
	return ss.rresolve(0)
}

func (ss steps) rresolve(level int) (workflow.Node, error) {
	if len(ss) == 0 {
		return nil, nil
	}

	var (
		err  error
		next workflow.Node
		//fss  = ss.flatten()

		// total number of all unresolved dependencies
		udeps = len(ss)

		// max resolve iterations we should do
		max = len(ss)

		echo = func(format string, aa ...interface{}) {
			fmt.Printf(strings.Repeat("\t", level)+format+"\n", aa...)
		}
	)

	for _, s := range ss {
		// dec. count of unresolved deps for all steps that are already resolved
		if s.wfn != nil {
			udeps--
		}
	}

	//echo("flattend steps")

	echo(strings.Repeat("%%", 80))
	defer echo(strings.Repeat("%%", 80))

	for m := 0; m < max; m++ {
		echo("iteration %d with %d udeps, %d max", m+1, udeps, max)
		//spew.Dump(fss)
		// going back from the last step and resolve it
		for i := len(ss) - 1; i >= 0; i-- {
			echo(strings.Repeat("=", 80))
			echo("proc %d udeps %d, refs <%s> subs: %d", i, udeps, ss[i].ref, len(ss[i].steps))

			next = nil
			if next, err = ss[i].steps.rresolve(level + 1); err != nil {
				return nil, err
			}

			if ss[i].wfn != nil {
				//echo("  => workflow node already set")
				continue
			}

			if ss[i].fn == nil {
				//echo("  => resolve fn missing")
				continue
			}

			// find next node and send it to resolver
			if i < max-1 && next == nil {
				next = ss[i+1].wfn
			}

			//spew.Dump(i, next)
			ss[i].wfn, err = ss[i].fn(next, ss)
			if err == nil {
				//echo("  => resolved")
				// resolved!!
				udeps--
				continue
			}

			if err == ErrDepsMissing {
				//echo("  => dep missing")
				// offload to next iteration
				continue
			}
		}

		if udeps == 0 {
			break
		}
	}

	if udeps > 0 {
		return nil, fmt.Errorf("could not resolve all deps (%d)", udeps)
	}

	for _, s := range ss {
		if s.wfn != nil {
			return s.wfn, nil
		}
	}

	return nil, nil
}

//func (ss steps) flatten() steps {
//	var out = ss
//
//	// @todo check for unique refs.
//
//	for _, s := range ss {
//		out = append(out, s.steps.flatten()...)
//	}
//
//	for _, s := range ss {
//		s.steps = nil
//	}
//
//	return out
//}

func (ss steps) nodeLookupByRef(ref string) workflow.Node {
	for _, s := range ss {
		if s.ref == ref {
			return s.wfn
		}
	}

	return nil
}

func (ss *steps) UnmarshalYAML(n *yaml.Node) error {
	var (
		s   *step
		buf = steps{}
	)

	_ = spew.Dump

	err := iterator(n, func(k *yaml.Node, v *yaml.Node) (err error) {
		s = &step{ykNode: k, yvNode: v}

		if !handleEndEvent(k, v, s) {
			if k != nil {
				// When mapping node is used, use key node as ref
				s.ref = k.Value
			}

			if err = v.Decode(s); err != nil {
				return
			}
		}

		if len(buf) > 0 {
			// reference current node to the previous
			// this will be resolved later when resolve() is called
			buf[len(buf)-1].next = s.ref
		}

		buf = append(buf, s)

		return
	})

	if err != nil {
		return fmt.Errorf("could not unmarshal steps: %w", err)
	}

	*ss = buf
	return nil
}
