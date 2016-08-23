// Copyright 2016 Canonical Ltd.  All rights reserved.

package cmd_test

import (
	jujucmd "github.com/juju/cmd"
	"github.com/juju/cmd/cmdtesting"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/macaroon-bakery.v1/httpbakery"

	"github.com/CanonicalLtd/plans-client/api"
	"github.com/CanonicalLtd/plans-client/cmd"
	plantesting "github.com/CanonicalLtd/plans-client/testing"
)

type resumeSuite struct {
	testing.CleanupSuite
	mockAPI *plantesting.MockPlanClient
	stub    *testing.Stub
}

var _ = gc.Suite(&resumeSuite{})

func (s *resumeSuite) SetUpTest(c *gc.C) {
	s.stub = &testing.Stub{}

	s.mockAPI = plantesting.NewMockPlanClient()

	s.PatchValue(cmd.NewClient, func(string, *httpbakery.Client) (api.PlanClient, error) {
		return s.mockAPI, nil
	})
}

func (s *resumeSuite) TestAttach(c *gc.C) {
	tests := []struct {
		about       string
		args        []string
		err         string
		assertCalls func(*testing.Stub)
	}{{
		about: "everything works",
		args:  []string{"resume-plan", "testisv/default", "some-charm-url1", "some-charm-url2"},
		assertCalls: func(stub *testing.Stub) {
			stub.CheckCall(c, 0, "Resume", "testisv/default", false, []string{"some-charm-url1", "some-charm-url2"})
		},
	}, {
		about: "everything works - all flag",
		args:  []string{"resume-plan", "testisv/default", "--all"},
		assertCalls: func(stub *testing.Stub) {
			stub.CheckCall(c, 0, "Resume", "testisv/default", true, []string{})
		},
	}, {
		about: "arg conflict - all flag",
		args:  []string{"resume-plan", "testisv/default", "some-charm-url", "--all"},
		err:   `cannot use --all and specify charm urls`,
		assertCalls: func(stub *testing.Stub) {
			stub.CheckNoCalls(c)
		},
	}, {
		about: "missing args",
		args:  []string{"resume-plan"},
		err:   `missing plan or charm url`,
		assertCalls: func(stub *testing.Stub) {
			stub.CheckNoCalls(c)
		},
	},
	}

	for i, t := range tests {
		s.mockAPI.ResetCalls()
		testCommand := jujucmd.NewSuperCommand(
			jujucmd.SuperCommandParams{
				Name:    "test",
				Doc:     "test command",
				Purpose: "testing",
			},
		)
		testCommand.Register(cmd.NewResumeCommand())

		c.Logf("Running test %d %s", i, t.about)
		_, err := cmdtesting.RunCommand(c, testCommand, t.args...)
		if t.err != "" {
			c.Assert(err, gc.ErrorMatches, t.err)
		} else {
			c.Assert(err, jc.ErrorIsNil)
		}
		t.assertCalls(s.mockAPI.Stub)
	}
}
