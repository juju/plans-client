// Copyright 2016 Canonical Ltd.  All rights reserved.

package cmd_test

import (
	"github.com/juju/cmd/cmdtesting"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/macaroon-bakery.v1/httpbakery"

	"github.com/CanonicalLtd/plans-client/api"
	"github.com/CanonicalLtd/plans-client/cmd"
	plantesting "github.com/CanonicalLtd/plans-client/testing"
)

type releaseSuite struct {
	testing.CleanupSuite
	mockAPI *plantesting.MockPlanClient
	stub    *testing.Stub
}

var _ = gc.Suite(&releaseSuite{})

func (s *releaseSuite) SetUpTest(c *gc.C) {
	s.stub = &testing.Stub{}

	s.mockAPI = plantesting.NewMockPlanClient()

	s.PatchValue(cmd.NewClient, func(string, *httpbakery.Client) (api.PlanClient, error) {
		return s.mockAPI, nil
	})
}

func (s *releaseSuite) TestReleaseCommand(c *gc.C) {
	tests := []struct {
		about   string
		args    []string
		err     string
		stdout  string
		apiCall []interface{}
	}{{
		about:  "unrecognized args causes error",
		args:   []string{"testisv/default", "foobar"},
		stdout: "ERROR unknown command line arguments: foobar\n",
		err:    `unknown command line arguments: foobar`,
	}, {
		about: "everything works",
		args:  []string{"testisv/default", "--url", "localhost:0"},
		stdout: `testisv/default/1
effective from 01 Jan 16 01:00 UTC
`,
		apiCall: []interface{}{"testisv/default"},
	},
	}

	for i, t := range tests {
		c.Logf("Running test %d %s", i, t.about)
		ctx, err := cmdtesting.RunCommand(c, cmd.NewReleaseCommand(), t.args...)
		if t.err != "" {
			c.Assert(err, gc.ErrorMatches, t.err)
			c.Assert(s.mockAPI.Calls(), gc.HasLen, 0)
		} else {
			c.Assert(err, jc.ErrorIsNil)
			c.Assert(s.mockAPI.Calls(), gc.HasLen, 1)
			s.mockAPI.CheckCall(c, 0, "Release", t.apiCall...)
		}
		if ctx != nil {
			c.Assert(cmdtesting.Stderr(ctx), gc.Equals, t.stdout)
		}
	}
}
