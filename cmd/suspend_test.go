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

type suspendSuite struct {
	testing.CleanupSuite
	mockAPI *plantesting.MockPlanClient
	stub    *testing.Stub
}

var _ = gc.Suite(&suspendSuite{})

func (s *suspendSuite) SetUpTest(c *gc.C) {
	s.stub = &testing.Stub{}

	s.mockAPI = plantesting.NewMockPlanClient()

	s.PatchValue(cmd.NewClient, func(string, *httpbakery.Client) (api.PlanClient, error) {
		return s.mockAPI, nil
	})
}

func (s *suspendSuite) TestCommand(c *gc.C) {
	tests := []struct {
		about       string
		args        []string
		err         string
		assertCalls func(*testing.Stub)
	}{{
		about: "everything works",
		args:  []string{"testisv/default", "some-charm-url1", "some-charm-url2"},
		assertCalls: func(stub *testing.Stub) {
			stub.CheckCall(c, 0, "Suspend", "testisv/default", false, []string{"some-charm-url1", "some-charm-url2"})
		},
	}, {
		about: "everything works - all flag",
		args:  []string{"testisv/default", "--all"},
		assertCalls: func(stub *testing.Stub) {
			stub.CheckCall(c, 0, "Suspend", "testisv/default", true, []string{})
		},
	}, {
		about: "arg conflict - all flag",
		args:  []string{"testisv/default", "some-charm-url", "--all"},
		err:   `cannot use --all and specify charm urls`,
		assertCalls: func(stub *testing.Stub) {
			stub.CheckNoCalls(c)
		},
	}, {
		about: "missing args",
		args:  []string{"suspend-plan"},
		err:   `missing plan or charm url`,
		assertCalls: func(stub *testing.Stub) {
			stub.CheckNoCalls(c)
		},
	},
	}

	for i, t := range tests {
		s.mockAPI.ResetCalls()
		c.Logf("Running test %d %s", i, t.about)
		_, err := cmdtesting.RunCommand(c, cmd.NewSuspendCommand(), t.args...)
		if t.err != "" {
			c.Assert(err, gc.ErrorMatches, t.err)
		} else {
			c.Assert(err, jc.ErrorIsNil)
		}
		t.assertCalls(s.mockAPI.Stub)
	}
}
