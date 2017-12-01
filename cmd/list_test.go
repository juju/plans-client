// Copyright 2017 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package cmd_test

import (
	"time"

	"github.com/juju/cmd/cmdtesting"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/macaroon-bakery.v1/httpbakery"

	"github.com/juju/plans-client/api"
	"github.com/juju/plans-client/api/wireformat"
	"github.com/juju/plans-client/cmd"
	plantesting "github.com/juju/plans-client/testing"
)

type listPlansSuite struct {
	testing.CleanupSuite
	mockAPI *plantesting.MockPlanClient
	stub    *testing.Stub
}

var _ = gc.Suite(&listPlansSuite{})

func (s *listPlansSuite) SetUpTest(c *gc.C) {
	s.stub = &testing.Stub{}

	s.mockAPI = plantesting.NewMockPlanClient()

	s.PatchValue(cmd.NewClient, func(string, *httpbakery.Client) (api.PlanClient, error) {
		return s.mockAPI, nil
	})
	s.PatchValue(cmd.ReadFile, func(string) ([]byte, error) {
		return []byte(plantesting.TestPlan), nil
	})
}

func (s *listPlansSuite) TestCommand(c *gc.C) {
	tests := []struct {
		about   string
		args    []string
		err     string
		stdout  string
		plans   []wireformat.Plan
		apiCall []interface{}
	}{{
		about: "unrecognized args causes error",
		args:  []string{"testisv", "foobar"},
		err:   `unknown command line arguments: foobar`,
	}, {
		about: "everything works",
		args:  []string{"canonical", "--url", "localhost:0"},
		stdout: `PLAN                 	          CREATED ON	EFFECTIVE TIME	     DEFINITION
canonical/test-plan/1	2017-12-01T00:00:00Z	              	test definition
`,
		plans: []wireformat.Plan{{
			Id:              "canonical/test-plan/1",
			URL:             "canonical/test-plan",
			Definition:      "test definition",
			CreatedOn:       time.Date(2017, 12, 1, 0, 0, 0, 0, time.UTC).UTC().Format(time.RFC3339),
			PlanDescription: "test plan description",
			PlanPrice:       "test plan price",
			Released:        true,
		}},
		apiCall: []interface{}{"canonical"},
	}, {
		about: "everything works - json",
		args:  []string{"canonical", "--url", "localhost:0", "--format", "json"},
		stdout: `[{"id":"canonical/test-plan/1","url":"canonical/test-plan","plan":"test definition","created-on":"2017-12-01T00:00:00Z","description":"test plan description","price":"test plan price","released":true}]
`,
		plans: []wireformat.Plan{{
			Id:              "canonical/test-plan/1",
			URL:             "canonical/test-plan",
			Definition:      "test definition",
			CreatedOn:       time.Date(2017, 12, 1, 0, 0, 0, 0, time.UTC).UTC().Format(time.RFC3339),
			PlanDescription: "test plan description",
			PlanPrice:       "test plan price",
			Released:        true,
		}},
		apiCall: []interface{}{"canonical"},
	}, {
		about: "everything works - yaml",
		args:  []string{"canonical", "--url", "localhost:0", "--format", "yaml"},
		stdout: `- id: canonical/test-plan/1
  url: canonical/test-plan
  plan: test definition
  created-on: 2017-12-01T00:00:00Z
  description: test plan description
  price: test plan price
  released: true
`,
		plans: []wireformat.Plan{{
			Id:              "canonical/test-plan/1",
			URL:             "canonical/test-plan",
			Definition:      "test definition",
			CreatedOn:       time.Date(2017, 12, 1, 0, 0, 0, 0, time.UTC).UTC().Format(time.RFC3339),
			PlanDescription: "test plan description",
			PlanPrice:       "test plan price",
			Released:        true,
		}},
		apiCall: []interface{}{"canonical"},
	},
	}

	for i, t := range tests {
		c.Logf("Running test %d %s", i, t.about)
		s.mockAPI.Plans = t.plans
		s.mockAPI.ResetCalls()
		ctx, err := cmdtesting.RunCommand(c, cmd.NewListPlansCommand(), t.args...)
		if t.err != "" {
			c.Assert(err, gc.ErrorMatches, t.err)
			c.Assert(s.mockAPI.Calls(), gc.HasLen, 0)
		} else {
			c.Assert(err, jc.ErrorIsNil)
			c.Assert(s.mockAPI.Calls(), gc.HasLen, 1)
			s.mockAPI.CheckCall(c, 0, "GetPlans", t.apiCall...)
		}
		if ctx != nil {
			c.Assert(cmdtesting.Stdout(ctx), gc.Equals, t.stdout)
		}
	}
}
