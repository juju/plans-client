// Copyright 2015 Canonical Ltd.

package wireformat_test

import (
	stdtesting "testing"

	"github.com/juju/utils"
	gc "gopkg.in/check.v1"

	"github.com/CanonicalLtd/plans-client/api/wireformat"
)

const PingPlan = `
    metrics:
      pings:
        unit:
          transform: max
          period: hour
          gaps: zero
        price: 0.01
      pongs:
        unit:
          transform: max
          period: hour
          gaps: zero
        price: 0.2
`

func Test(t *stdtesting.T) { gc.TestingT(t) }

var _ = gc.Suite(&suite{})

type suite struct{}

func (*suite) TestPlanWireValidation(c *gc.C) {
	tests := []struct {
		about  string
		plan   wireformat.Plan
		result string
	}{{
		about: "a valid plan",
		plan: wireformat.Plan{
			URL:        "isv/plana",
			Definition: PingPlan},
		result: "",
	}, {
		about: "no plan name",
		plan: wireformat.Plan{
			URL:        "",
			Definition: PingPlan},
		result: "empty plan url",
	}, {
		about: "invalid name format",
		plan: wireformat.Plan{
			URL:        "isvname",
			Definition: PingPlan},
		result: "plan url \"isvname\" not valid",
	}, {
		about: "missing definition",
		plan: wireformat.Plan{
			URL:        "isv/plan",
			Definition: ""},
		result: "missing plan definition",
	}}
	for i, test := range tests {
		c.Logf("%d : %s", i, test.about)
		err := test.plan.Validate()
		if test.result == "" {
			c.Check(err, gc.IsNil)
		} else {
			c.Check(err, gc.ErrorMatches, test.result)
		}
	}
}

func (*suite) TestAuthorizationRequestValidation(c *gc.C) {
	tests := []struct {
		about   string
		request wireformat.AuthorizationRequest
		result  string
	}{{
		about: "a valid request",
		request: wireformat.AuthorizationRequest{
			EnvironmentUUID: utils.MustNewUUID().String(),
			CharmURL:        "cs:wordpress",
			ServiceName:     "wordpress",
			PlanURL:         "test-isv/default",
		},
		result: "",
	}, {
		about: "invalid env uuid",
		request: wireformat.AuthorizationRequest{
			EnvironmentUUID: "abc",
			CharmURL:        "cs:wordpress",
			ServiceName:     "wordpress",
			PlanURL:         "test-isv/default",
		},
		result: "invalid environment UUID: \"abc\"",
	}, {
		about: "missing charm url",
		request: wireformat.AuthorizationRequest{
			EnvironmentUUID: utils.MustNewUUID().String(),
			ServiceName:     "wordpress",
			PlanURL:         "test-isv/default",
		},
		result: "undefined charm url",
	}, {
		about: "invalid service name",
		request: wireformat.AuthorizationRequest{
			EnvironmentUUID: utils.MustNewUUID().String(),
			CharmURL:        "cs:wordpress",
			ServiceName:     "wordpress/0",
			PlanURL:         "test-isv/default",
		},
		result: "invalid service name: \"wordpress/0\"",
	}, {
		about: "invalid charm url",
		request: wireformat.AuthorizationRequest{
			EnvironmentUUID: utils.MustNewUUID().String(),
			CharmURL:        "http://my-charm.com",
			ServiceName:     "wordpress",
			PlanURL:         "test-isv/default",
		},
		result: "invalid charm url: \"http://my-charm.com\"",
	}, {
		about: "missing plan url",
		request: wireformat.AuthorizationRequest{
			EnvironmentUUID: utils.MustNewUUID().String(),
			CharmURL:        "cs:wordpress",
			ServiceName:     "wordpress",
		},
		result: "undefined plan url",
	}}
	for i, test := range tests {
		c.Logf("%d : %s", i, test.about)
		err := test.request.Validate()
		if test.result == "" {
			c.Check(err, gc.IsNil)
		} else {
			c.Check(err, gc.ErrorMatches, test.result)
		}
	}
}
