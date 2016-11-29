// Copyright 2016 Canonical Ltd.

package testing_test

import (
	"testing"

	gc "gopkg.in/check.v1"

	t "github.com/CanonicalLtd/plans-client/testing"
)

func Test(t *testing.T) {
	gc.TestingT(t)
}

type TSuite struct{}

var _ = gc.Suite(&TSuite{})

func (s *TSuite) TestEquals(c *gc.C) {
	t1 := `line1
line2
line3`
	r, err := t.MultilineEquals.Check([]interface{}{t1, t1}, t.MultilineEquals.Params)
	c.Assert(r, gc.Equals, true)
	c.Assert(err, gc.Equals, "")
}

func (s *TSuite) TestNotEquals(c *gc.C) {
	t1 := `line1
line2
line3`

	t2 := `line1
line2
`
	r, err := t.MultilineEquals.Check([]interface{}{t1, t2}, t.MultilineEquals.Params)
	c.Assert(r, gc.Equals, false)
	c.Assert(err, gc.Equals, `unequal line count: "obtained" has more lines than "expected"`)
}
