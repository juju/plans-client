// Copyright 2017 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package cmd_test

import (
	"github.com/juju/cmd/cmdtesting"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/macaroon-bakery.v2/httpbakery"

	"github.com/juju/plans-client/api"
	"github.com/juju/plans-client/cmd"
	plantesting "github.com/juju/plans-client/testing"
)

type pushSuite struct {
	testing.CleanupSuite
	mockAPI *plantesting.MockPlanClient
	stub    *testing.Stub
}

var _ = gc.Suite(&pushSuite{})

func (s *pushSuite) SetUpTest(c *gc.C) {
	s.stub = &testing.Stub{}

	s.mockAPI = plantesting.NewMockPlanClient()

	s.PatchValue(cmd.NewClient, func(string, *httpbakery.Client) (api.PlanClient, error) {
		return s.mockAPI, nil
	})
	s.PatchValue(cmd.ReadFile, func(string) ([]byte, error) {
		return []byte(plantesting.TestPlan), nil
	})
}

func (s *pushSuite) TestCommand(c *gc.C) {
	tests := []struct {
		about   string
		args    []string
		err     string
		stdout  string
		apiCall []interface{}
	}{{
		about: "unrecognized args causes error",
		args:  []string{"example.yaml", "testisv/default", "foobar"},
		err:   `unknown command line arguments: foobar`,
	}, {
		about:   "everything works",
		args:    []string{"example.yaml", "testisv/default", "--url", "localhost:0"},
		stdout:  "testisv/default/17\n",
		apiCall: []interface{}{"testisv/default", plantesting.TestPlan},
	},
	}

	for i, t := range tests {
		c.Logf("Running test %d %s", i, t.about)
		ctx, err := cmdtesting.RunCommand(c, cmd.NewPushCommand(), t.args...)
		if t.err != "" {
			c.Assert(err, gc.ErrorMatches, t.err)
			c.Assert(s.mockAPI.Calls(), gc.HasLen, 0)
		} else {
			c.Assert(err, jc.ErrorIsNil)
			c.Assert(s.mockAPI.Calls(), gc.HasLen, 1)
			s.mockAPI.CheckCall(c, 0, "Save", t.apiCall...)
		}
		if ctx != nil {
			c.Assert(cmdtesting.Stdout(ctx), gc.Equals, t.stdout)
		}
	}
}
