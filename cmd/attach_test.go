// Copyright 2016 Canonical Ltd.  All rights reserved.

package cmd_test

import (
	"net/http"

	"github.com/juju/cmd/cmdtesting"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/macaroon-bakery.v1/httpbakery"

	"github.com/CanonicalLtd/plans-client/api"
	"github.com/CanonicalLtd/plans-client/cmd"
	plantesting "github.com/CanonicalLtd/plans-client/testing"
)

type attachSuite struct {
	testing.CleanupSuite
	mockAPI *plantesting.MockPlanClient
	stub    *testing.Stub
}

var _ = gc.Suite(&attachSuite{})

func (s *attachSuite) SetUpTest(c *gc.C) {
	s.stub = &testing.Stub{}

	s.mockAPI = plantesting.NewMockPlanClient()

	s.PatchValue(cmd.NewClient, func(string, *httpbakery.Client) (api.PlanClient, error) {
		return s.mockAPI, nil
	})
	s.PatchValue(cmd.ReadFile, func(string) ([]byte, error) {
		return []byte(plantesting.TestPlan), nil
	})
}

func (s *attachSuite) TestCommand(c *gc.C) {
	tests := []struct {
		about            string
		args             []string
		charmMetrics     []string
		resolvedCharmURL string
		err              string
		stdout           string
		assertCalls      func(*testing.Stub)
	}{{
		about:  "unrecognized args causes error",
		args:   []string{"some-charm-url", "testisv", "some-arg"},
		err:    `unknown command line arguments: some-arg`,
		stdout: "",
		assertCalls: func(stub *testing.Stub) {
			stub.CheckNoCalls(c)
		},
	}, {
		about: "everything works",
		args:  []string{"some-charm-url", "testisv/default"},
		stdout: `OK
`,
		assertCalls: func(stub *testing.Stub) {
			stub.CheckCall(c, 0, "Get", "testisv/default")
			stub.CheckCall(c, 1, "AddCharm", "testisv/default", "some-charm-url", false)
		},
	}, {
		about:            "unresolved charm url causes error",
		resolvedCharmURL: "series/some-charm-url-4",
		args:             []string{"some-charm-url", "testisv/default", "--default"},
		err:              `charm url "some-charm-url" is not resolved - did you mean "series/some-charm-url-4"\?`,
		stdout:           "",
		assertCalls: func(stub *testing.Stub) {
			stub.CheckNoCalls(c)
		},
	}, {
		about: "everything works - set default plan",
		args:  []string{"some-charm-url", "testisv/default", "--default"},
		stdout: `OK
`,
		assertCalls: func(stub *testing.Stub) {
			stub.CheckCall(c, 0, "Get", "testisv/default")
			stub.CheckCall(c, 1, "AddCharm", "testisv/default", "some-charm-url", true)
		},
	}, {
		about:        "plan not valid for charm",
		args:         []string{"some-charm-url", "testisv/default"},
		charmMetrics: []string{"pings"},
		err:          "plan testisv/default cannot be used to rate charm some-charm-url: no common metrics",
		assertCalls: func(stub *testing.Stub) {
			stub.CheckCall(c, 0, "Get", "testisv/default")
		},
	}, {
		about: "missing args",
		args:  []string{},
		err:   `missing charm and plan url`,
		assertCalls: func(stub *testing.Stub) {
			stub.CheckNoCalls(c)
		},
	},
	}

	for i, t := range tests {
		s.mockAPI.ResetCalls()
		testCommand := &cmd.AttachCommand{
			CharmResolver: &mockCharmResolver{
				Stub:         &testing.Stub{},
				ResolvedURL:  t.resolvedCharmURL,
				CharmMetrics: t.charmMetrics,
			},
		}

		c.Logf("Running test %d %s", i, t.about)
		ctx, err := cmdtesting.RunCommand(c, testCommand, t.args...)
		if t.err != "" {
			c.Assert(err, gc.ErrorMatches, t.err)
		} else {
			c.Assert(err, jc.ErrorIsNil)
		}
		t.assertCalls(s.mockAPI.Stub)

		if ctx != nil {
			c.Assert(cmdtesting.Stdout(ctx), gc.Equals, t.stdout)
		}
	}
}

// mockCharmResolver is a mock implementation of cmd.CharmResolver.
type mockCharmResolver struct {
	*testing.Stub
	ResolvedURL  string
	CharmMetrics []string
}

// Resolve implements cmd.CharmResolver.
func (r *mockCharmResolver) Resolve(_ *http.Client, charmURL string) (string, error) {
	r.AddCall("Resolve", charmURL)
	if r.ResolvedURL != "" {
		return r.ResolvedURL, r.NextErr()
	}
	return charmURL, r.NextErr()
}

func (r *mockCharmResolver) Metrics(_ *http.Client, charmURL string) ([]string, error) {
	r.AddCall("Resolve", charmURL)
	if r.CharmMetrics != nil {
		return r.CharmMetrics, r.NextErr()
	}
	return []string{"active-users"}, r.NextErr()
}
