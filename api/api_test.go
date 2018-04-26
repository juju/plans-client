// Copyright 2017 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	stdtesting "testing"
	"time"

	"github.com/juju/errors"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/macaroon.v1"

	"github.com/juju/plans-client/api"
	"github.com/juju/plans-client/api/wireformat"
)

func Test(t *stdtesting.T) {
	gc.TestingT(t)
}

var testPlan = `
# Copyright 2014 Canonical Ltd.  All rights reserved.
    metrics:
      active-users:
        unit:
          transform: max
          period: hour
          gaps: zero
        price: 0.01
`

type clientIntegrationSuite struct {
	httpClient *mockHttpClient
	planClient api.PlanClient
}

var _ = gc.Suite(&clientIntegrationSuite{})

func (s *clientIntegrationSuite) SetUpTest(c *gc.C) {
	s.httpClient = &mockHttpClient{}

	client, err := api.NewPlanClient("https://api.staging.jujucharms.com/omnibus", api.HTTPClient(s.httpClient))
	c.Assert(err, jc.ErrorIsNil)
	s.planClient = client
}

func (s *clientIntegrationSuite) TestSave(c *gc.C) {
	s.httpClient.status = http.StatusOK
	s.httpClient.body = wireformat.Plan{
		Id:  "testisv/default/1",
		URL: "testisv/default",
	}

	plan, err := s.planClient.Save("testisv/default", testPlan)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(plan.Id, gc.Equals, "testisv/default/1")

	s.httpClient.assertRequest(c, "POST", "https://api.staging.jujucharms.com/omnibus/v4/p", wireformat.Plan{
		URL:        "testisv/default",
		Definition: testPlan,
	})
}

func (s *clientIntegrationSuite) TestSaveFail(c *gc.C) {
	s.httpClient.status = http.StatusBadRequest
	s.httpClient.body = struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{
		Code:    "bad request",
		Message: "silly error",
	}

	_, err := s.planClient.Save("testisv/default", testPlan)
	c.Assert(err, gc.ErrorMatches, `failed to save plan: silly error`)
}

func (s *clientIntegrationSuite) TestSaveUnauthorized(c *gc.C) {
	s.httpClient.SetErrors(errors.New("refused discharge: unauthorized"))

	_, err := s.planClient.Save("testisv/default", testPlan)
	c.Assert(err, gc.ErrorMatches, `unauthorized to save the plan: please run "charm whoami" to verify you are member of the "testisv" group`)
}

func (s *clientIntegrationSuite) TestRelease(c *gc.C) {
	p := wireformat.Plan{
		URL:        "testisv/default",
		Definition: testPlan,
	}
	s.httpClient.status = http.StatusOK
	s.httpClient.body = p

	plan, err := s.planClient.Release("testisv/default/1")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(plan, gc.DeepEquals, &p)
	s.httpClient.assertRequest(c, "POST", "https://api.staging.jujucharms.com/omnibus/v4/p/testisv/default/1/release", nil)
}

func (s *clientIntegrationSuite) TestReleaseInvalidPlanURL(c *gc.C) {
	_, err := s.planClient.Release("invalid/format/testisv/0/default")
	c.Assert(err, gc.ErrorMatches, `plan id "invalid/format/testisv/0/default" not valid`)
}

func (s *clientIntegrationSuite) TestReleaseFail(c *gc.C) {
	s.httpClient.status = http.StatusBadRequest
	s.httpClient.body = struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{
		Code:    "bad request",
		Message: "silly error",
	}

	_, err := s.planClient.Release("testisv/default/1")
	c.Assert(err, gc.ErrorMatches, `failed to release plan: silly error`)
}

func (s *clientIntegrationSuite) TestReleaseUnauthorized(c *gc.C) {
	s.httpClient.SetErrors(errors.New("refused discharge: unauthorized"))

	_, err := s.planClient.Release("testisv/default/1")
	c.Assert(err, gc.ErrorMatches, `release-plan is currently disabled for public use. Please ask in #juju-partners on freenode or email juju@lists.ubuntu.com: refused discharge: unauthorized`)
}

func (s *clientIntegrationSuite) TestSuspend(c *gc.C) {
	s.httpClient.status = http.StatusOK

	err := s.planClient.Suspend("testisv/default", false, "testisv/plan1", "testisv/plan2")
	c.Assert(err, jc.ErrorIsNil)
	s.httpClient.assertRequest(c, "POST", "https://api.staging.jujucharms.com/omnibus/v3/p/testisv/default/suspend", struct {
		All    bool     `json:"all"`
		Charms []string `json:"charms"`
	}{
		All:    false,
		Charms: []string{"testisv/plan1", "testisv/plan2"},
	})
}

func (s *clientIntegrationSuite) TestSuspendAll(c *gc.C) {
	s.httpClient.status = http.StatusOK

	err := s.planClient.Suspend("testisv/default", true)
	c.Assert(err, jc.ErrorIsNil)
	s.httpClient.assertRequest(c, "POST", "https://api.staging.jujucharms.com/omnibus/v3/p/testisv/default/suspend", struct {
		All    bool     `json:"all"`
		Charms []string `json:"charms"`
	}{
		All: true,
	})
}

func (s *clientIntegrationSuite) TestSuspendInvalidPlanURL(c *gc.C) {
	err := s.planClient.Suspend("invalid/format/testisv/default", false, "cs:~testers/charm1-0")
	c.Assert(err, gc.ErrorMatches, `plan url "invalid/format/testisv/default" not valid`)
}

func (s *clientIntegrationSuite) TestSuspendFail(c *gc.C) {
	s.httpClient.status = http.StatusBadRequest
	s.httpClient.body = struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{
		Code:    "bad request",
		Message: "silly error",
	}

	err := s.planClient.Suspend("testisv/default", false, "cs:~testers/charm1-0")
	c.Assert(err, gc.ErrorMatches, `failed to suspend plan: silly error`)
}

func (s *clientIntegrationSuite) TestResume(c *gc.C) {
	s.httpClient.status = http.StatusOK

	err := s.planClient.Resume("testisv/default", false, "testisv/plan1", "testisv/plan2")
	c.Assert(err, jc.ErrorIsNil)
	s.httpClient.assertRequest(c, "POST", "https://api.staging.jujucharms.com/omnibus/v3/p/testisv/default/resume", struct {
		All    bool     `json:"all"`
		Charms []string `json:"charms"`
	}{
		All:    false,
		Charms: []string{"testisv/plan1", "testisv/plan2"},
	})
}

func (s *clientIntegrationSuite) TestResumeAll(c *gc.C) {
	s.httpClient.status = http.StatusOK

	err := s.planClient.Resume("testisv/default", true)
	c.Assert(err, jc.ErrorIsNil)
	s.httpClient.assertRequest(c, "POST", "https://api.staging.jujucharms.com/omnibus/v3/p/testisv/default/resume", struct {
		All    bool     `json:"all"`
		Charms []string `json:"charms"`
	}{
		All: true,
	})
}

func (s *clientIntegrationSuite) TestResumeInvalidPlanURL(c *gc.C) {
	err := s.planClient.Resume("invalid/format/testisv/default", false, "cs:~testers/charm1-0")
	c.Assert(err, gc.ErrorMatches, `plan url "invalid/format/testisv/default" not valid`)
}

func (s *clientIntegrationSuite) TestResumeFail(c *gc.C) {
	s.httpClient.status = http.StatusBadRequest
	s.httpClient.body = struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{
		Code:    "bad request",
		Message: "silly error",
	}

	err := s.planClient.Resume("testisv/default", false, "cs:~testers/charm1-0")
	c.Assert(err, gc.ErrorMatches, `failed to resume plan: silly error`)
}

func (s *clientIntegrationSuite) TestResumeUnauthorized(c *gc.C) {
	s.httpClient.SetErrors(errors.New("refused discharge: unauthorized"))

	err := s.planClient.Resume("testisv/default", true)
	c.Assert(err, gc.ErrorMatches, `unauthorized to resume plan: please run "charm whoami" to verify you are member of the "testisv" group`)
}

func (s *clientIntegrationSuite) TestAddCharm(c *gc.C) {
	s.httpClient.status = http.StatusOK

	err := s.planClient.AddCharm("testisv/default", "cs:~testers/charm1-0", true)
	c.Assert(err, jc.ErrorIsNil)
	s.httpClient.assertRequest(c, "POST", "https://api.staging.jujucharms.com/omnibus/v3/charm", struct {
		Plan    string `json:"plan-url"`
		Charm   string `json:"charm-url"`
		Default bool   `json:"default"`
	}{
		Plan:    "testisv/default",
		Charm:   "cs:~testers/charm1-0",
		Default: true,
	})
}

func (s *clientIntegrationSuite) TestAddCharmUnauthorized(c *gc.C) {
	s.httpClient.SetErrors(errors.New("refused discharge: unauthorized"))

	err := s.planClient.AddCharm("testisv/default", "cs:~testers/charm1-0", true)
	c.Assert(err, gc.ErrorMatches, `unauthorized to add charm: please run "charm whoami" to verify you are member of the "testisv" group`)
}

func (s *clientIntegrationSuite) TestAddCharmFail(c *gc.C) {
	s.httpClient.status = http.StatusBadRequest
	s.httpClient.body = struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{
		Code:    "bad request",
		Message: "silly error",
	}

	err := s.planClient.AddCharm("testisv/default", "cs:~testers/charm1-0", false)
	c.Assert(err, gc.ErrorMatches, `failed to add charm: silly error`)
}

func (s *clientIntegrationSuite) TestGet(c *gc.C) {
	plans := []wireformat.Plan{{
		URL:        "testisv/default",
		Definition: testPlan,
	}}
	s.httpClient.status = http.StatusOK
	s.httpClient.body = plans

	response, err := s.planClient.Get("testisv/default")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(response, gc.DeepEquals, plans)
	s.httpClient.assertRequest(c, "GET", "https://api.staging.jujucharms.com/omnibus/v4/p/testisv/default", nil)
}

func (s *clientIntegrationSuite) TestGetFail(c *gc.C) {
	s.httpClient.status = http.StatusBadRequest
	s.httpClient.body = struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{
		Code:    "bad request",
		Message: "silly error",
	}

	_, err := s.planClient.Get("testisv/default")
	c.Assert(err, gc.ErrorMatches, `failed to retrieve plans: silly error`)
}

func (s *clientIntegrationSuite) TestGetPlans(c *gc.C) {
	p1 := wireformat.Plan{
		Id:         "testisv/default/2",
		URL:        "testisv/default",
		Definition: testPlan,
	}
	p2 := wireformat.Plan{
		Id:         "testisv/default/1",
		URL:        "testisv/default",
		Definition: testPlan,
	}
	p3 := wireformat.Plan{
		Id:         "testisv/another/1",
		URL:        "testisv/another",
		Definition: testPlan,
	}
	plans := []wireformat.Plan{p1, p2, p3}
	s.httpClient.status = http.StatusOK
	s.httpClient.body = plans

	response, err := s.planClient.GetPlans("testisv")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(response, gc.DeepEquals, []wireformat.Plan{p1, p2, p3})
	s.httpClient.assertRequest(c, "GET", "https://api.staging.jujucharms.com/omnibus/v4/p/testisv", nil)
}

func (s *clientIntegrationSuite) TestGetPlansFail(c *gc.C) {
	s.httpClient.status = http.StatusBadRequest
	s.httpClient.body = struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{
		Code:    "bad request",
		Message: "silly error",
	}

	_, err := s.planClient.GetPlans("testisv")
	c.Assert(err, gc.ErrorMatches, `failed to retrieve plans: silly error`)
}

func (s *clientIntegrationSuite) TestGetDefaultPlan(c *gc.C) {
	plan := wireformat.Plan{
		URL:        "testisv/default",
		Definition: testPlan,
	}
	s.httpClient.status = http.StatusOK
	s.httpClient.body = plan

	reponse, err := s.planClient.GetDefaultPlan("cs:~testers/charm1-0")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(reponse, gc.DeepEquals, &plan)
	s.httpClient.assertRequest(c, "GET", "https://api.staging.jujucharms.com/omnibus/v4/charm/default?charm-url="+url.QueryEscape("cs:~testers/charm1-0"), nil)
}

func (s *clientIntegrationSuite) TestGetDefaultPlanFail(c *gc.C) {
	s.httpClient.status = http.StatusBadRequest
	s.httpClient.body = struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{
		Code:    "bad request",
		Message: "silly error",
	}

	_, err := s.planClient.GetDefaultPlan("cs:~testers/charm1-0")
	c.Assert(err, gc.ErrorMatches, `failed to retrieve default plan: silly error`)
}

func (s *clientIntegrationSuite) TestGetPlansForCharm(c *gc.C) {
	plans := []wireformat.Plan{{
		URL:        "testisv/default",
		Definition: testPlan,
	}}
	s.httpClient.status = http.StatusOK
	s.httpClient.body = plans

	reponse, err := s.planClient.GetPlansForCharm("cs:~testers/charm1-0")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(reponse, gc.DeepEquals, plans)
	s.httpClient.assertRequest(c, "GET", "https://api.staging.jujucharms.com/omnibus/v4/charm?charm-url="+url.QueryEscape("cs:~testers/charm1-0"), nil)
}

func (s *clientIntegrationSuite) TestPlansForCharmFail(c *gc.C) {
	s.httpClient.status = http.StatusBadRequest
	s.httpClient.body = struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{
		Code:    "bad request",
		Message: "silly error",
	}

	_, err := s.planClient.GetPlansForCharm("cs:~testers/charm1-0")
	c.Assert(err, gc.ErrorMatches, `failed to retrieve associated plans: silly error`)
}

func (s *clientIntegrationSuite) TestGetPlanDetails(c *gc.C) {
	s.httpClient.status = http.StatusOK
	p := wireformat.PlanDetails{
		Plan: wireformat.Plan{
			URL:        "testisv/default",
			Definition: testPlan,
			CreatedOn:  time.Date(2015, 1, 1, 1, 0, 0, 0, time.UTC).Format(time.RFC3339),
		},
		Created: wireformat.Event{
			User: "jane.jaas",
			Type: "create",
			Time: time.Date(2015, 1, 1, 1, 0, 0, 0, time.UTC),
		},
		Released: &wireformat.Event{
			User: "jane.jaas",
			Type: "release",
			Time: time.Date(2015, 1, 1, 1, 0, 0, 0, time.UTC),
		},
		Charms: []wireformat.CharmPlanDetail{{
			CharmURL: "cs:~testisv/charm1-0",
			Attached: wireformat.Event{
				User: "jane.jaas",
				Type: "create",
				Time: time.Date(2015, 1, 1, 1, 0, 0, 0, time.UTC),
			},
			Default: false,
		}},
	}
	s.httpClient.body = p

	details, err := s.planClient.GetPlanDetails("testisv/default")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(details, gc.DeepEquals, &p)

	s.httpClient.assertRequest(c, "GET", "https://api.staging.jujucharms.com/omnibus/v4/p/testisv/default/details", nil)
}

func (s *clientIntegrationSuite) TestGetPlanDetailsRevision(c *gc.C) {
	s.httpClient.status = http.StatusOK
	p := wireformat.PlanDetails{
		Plan: wireformat.Plan{
			URL:        "testisv/default",
			Definition: testPlan,
			CreatedOn:  time.Date(2015, 1, 1, 1, 0, 0, 0, time.UTC).Format(time.RFC3339),
		},
		Created: wireformat.Event{
			User: "jane.jaas",
			Type: "create",
			Time: time.Date(2015, 1, 1, 1, 0, 0, 0, time.UTC),
		},
		Released: &wireformat.Event{
			User: "jane.jaas",
			Type: "release",
			Time: time.Date(2015, 1, 1, 1, 0, 0, 0, time.UTC),
		},
		Charms: []wireformat.CharmPlanDetail{{
			CharmURL: "cs:~testisv/charm1-0",
			Attached: wireformat.Event{
				User: "jane.jaas",
				Type: "create",
				Time: time.Date(2015, 1, 1, 1, 0, 0, 0, time.UTC),
			},
			Default: false,
		}},
	}
	s.httpClient.body = p

	details, err := s.planClient.GetPlanDetails("testisv/default/7")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(details, gc.DeepEquals, &p)

	s.httpClient.assertRequest(c, "GET", "https://api.staging.jujucharms.com/omnibus/v4/p/testisv/default/details?revision=7", nil)
}

func (s *clientIntegrationSuite) TestGetPlanDetailsFail(c *gc.C) {
	s.httpClient.status = http.StatusBadRequest
	s.httpClient.body = struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{
		Code:    "bad request",
		Message: "silly error",
	}

	_, err := s.planClient.GetPlanDetails("testisv/default")
	c.Assert(err, gc.ErrorMatches, `failed to retrieve plan details: silly error`)
}

func (s *clientIntegrationSuite) TestGetPlanDetailsUnauthorized(c *gc.C) {
	s.httpClient.SetErrors(errors.New("refused discharge: unauthorized"))

	_, err := s.planClient.GetPlanDetails("testisv/default")
	c.Assert(err, gc.ErrorMatches, `unauthorized to retrieve plan details: please run "charm whoami" to verify you are member of the "testisv" group`)
}

func (s *clientIntegrationSuite) TestGetPlanDetailsNotFound(c *gc.C) {
	s.httpClient.status = http.StatusNotFound
	s.httpClient.body = struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{
		Code:    "not found",
		Message: "silly error",
	}

	_, err := s.planClient.GetPlanDetails("testisv/default")
	c.Assert(err, gc.ErrorMatches, `failed to retrieve plan details: silly error`)
}

func (s *clientIntegrationSuite) TestSuspendResumeFailsWithPlanRevision(c *gc.C) {
	err := s.planClient.Suspend("testisv/default/1", false, "cs:~testers/charm1-0")
	c.Assert(err, gc.ErrorMatches, `plan url "testisv/default/1" not valid`)
	err = s.planClient.Suspend("testisv/default/1", false, "cs:~testers/charm1-0")
	c.Assert(err, gc.ErrorMatches, `plan url "testisv/default/1" not valid`)
}

func (s *clientIntegrationSuite) TestSuspendUnauthorized(c *gc.C) {
	s.httpClient.SetErrors(errors.New("refused discharge: unauthorized"))

	err := s.planClient.Suspend("testisv/default", true)
	c.Assert(err, gc.ErrorMatches, `unauthorized to suspend plan: please run "charm whoami" to verify you are member of the "testisv" group`)
}

func (s *clientIntegrationSuite) TestGetPlanRevisions(c *gc.C) {
	plans := []wireformat.Plan{{
		Id:         "testisv/default/1",
		URL:        "testisv/default",
		Definition: testPlan,
	}, {
		Id:         "testisv/default/2",
		URL:        "testisv/default",
		Definition: testPlan,
	}}
	s.httpClient.status = http.StatusOK
	s.httpClient.body = plans

	response, err := s.planClient.GetPlanRevisions("testisv/default")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(response, gc.DeepEquals, plans)
	s.httpClient.assertRequest(c, "GET", "https://api.staging.jujucharms.com/omnibus/v4/p/testisv/default/revisions", nil)
}

func (s *clientIntegrationSuite) TestGetPlanRevisionsFail(c *gc.C) {
	s.httpClient.status = http.StatusBadRequest
	s.httpClient.body = struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{
		Code:    "bad request",
		Message: "silly error",
	}

	_, err := s.planClient.GetPlanRevisions("testisv/default")
	c.Assert(err, gc.ErrorMatches, `failed to retrieve plan revisions: silly error`)
}

func (s *clientIntegrationSuite) TestGetPlanRevisionsUnauthorized(c *gc.C) {
	s.httpClient.SetErrors(errors.New("refused discharge: unauthorized"))

	_, err := s.planClient.GetPlanRevisions("testisv/default")
	c.Assert(err, gc.ErrorMatches, `unauthorized to retrieve plan revisions: please run "charm whoami" to verify you are member of the "testisv" group`)
}

func (s *clientIntegrationSuite) TestAuthorize(c *gc.C) {
	m, err := macaroon.New([]byte{}, "abc", "")
	c.Assert(err, jc.ErrorIsNil)
	s.httpClient.status = http.StatusOK
	s.httpClient.body = m

	client, err := api.NewPlanClient("https://api.staging.jujucharms.com/omnibus", api.HTTPClient(s.httpClient))
	c.Assert(err, jc.ErrorIsNil)
	macaroon, err := client.Authorize("envUUID", "cs:~testers/charm1-0", "test-service", "testisv/default")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(macaroon, jc.DeepEquals, m)
	s.httpClient.assertRequest(c, "POST", "https://api.staging.jujucharms.com/omnibus/v3/plan/authorize", wireformat.AuthorizationRequest{
		EnvironmentUUID: "envUUID",
		CharmURL:        "cs:~testers/charm1-0",
		ServiceName:     "test-service",
		PlanURL:         "testisv/default",
	})
}

func (s *clientIntegrationSuite) TestAuthorizeFail(c *gc.C) {
	s.httpClient.status = http.StatusBadRequest
	s.httpClient.body = struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{
		Code:    "bad request",
		Message: "silly error",
	}

	client, err := api.NewPlanClient("", api.HTTPClient(s.httpClient))
	c.Assert(err, jc.ErrorIsNil)
	_, err = client.Authorize("envUUID", "cs:~testers/charm1-0", "test-service", "testisv/default")
	c.Assert(err, gc.ErrorMatches, `failed to authorize plan: silly error`)
}

func (s *clientIntegrationSuite) TestResellerAuthorize(c *gc.C) {
	m, err := macaroon.New([]byte{}, "abc", "")
	c.Assert(err, jc.ErrorIsNil)
	s.httpClient.status = http.StatusOK
	s.httpClient.body = m

	client, err := api.NewPlanClient("https://api.staging.jujucharms.com/omnibus", api.HTTPClient(s.httpClient))
	c.Assert(err, jc.ErrorIsNil)
	macaroon, err := client.AuthorizeReseller("canonical/jimm", "cs:~sabdf/jimm-0", "jimm", "sabdfl", "test-user")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(macaroon, jc.DeepEquals, m)
	s.httpClient.assertRequest(c, "POST", "https://api.staging.jujucharms.com/omnibus/v3/plan/reseller/authorize", wireformat.ResellerAuthorizationRequest{
		Plan:             "canonical/jimm",
		CharmURL:         "cs:~sabdf/jimm-0",
		Application:      "jimm",
		ApplicationOwner: "sabdfl",
		ApplicationUser:  "test-user",
	})
}

func (s *clientIntegrationSuite) TestResellerAuthorizeFail(c *gc.C) {
	s.httpClient.status = http.StatusBadRequest
	s.httpClient.body = struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{
		Code:    "bad request",
		Message: "silly error",
	}

	client, err := api.NewPlanClient("", api.HTTPClient(s.httpClient))
	c.Assert(err, jc.ErrorIsNil)
	_, err = client.AuthorizeReseller("canonical/jimm", "cs:~sabdf/jimm-0", "jimm", "sabdfl", "test-user")
	c.Assert(err, gc.ErrorMatches, `failed to authorize reseller plan: silly error`)
}

func (s *clientIntegrationSuite) TestGetResellerAuthorization(c *gc.C) {
	s.httpClient.status = http.StatusOK
	s.httpClient.body = []wireformat.ResellerAuthorization{{
		AuthUUID: "blah-di-blah",
	}}

	client, err := api.NewPlanClient("https://api.staging.jujucharms.com/omnibus", api.HTTPClient(s.httpClient))
	c.Assert(err, jc.ErrorIsNil)
	auths, err := client.GetResellerAuthorizations(wireformat.ResellerAuthorizationQuery{Reseller: "isv"})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(auths, gc.HasLen, 1)
	s.httpClient.assertRequest(c, "GET", "https://api.staging.jujucharms.com/omnibus/v4/plan/resellers/authorization?reseller=isv", nil)
}

func (s *clientIntegrationSuite) TestGetResellerAuthorizationEmptyQuery(c *gc.C) {
	s.httpClient.status = http.StatusOK
	s.httpClient.body = []wireformat.ResellerAuthorization{{
		AuthUUID: "blah-di-blah",
	}}

	client, err := api.NewPlanClient("https://api.staging.jujucharms.com/omnibus", api.HTTPClient(s.httpClient))
	c.Assert(err, jc.ErrorIsNil)
	auths, err := client.GetResellerAuthorizations(wireformat.ResellerAuthorizationQuery{})
	c.Assert(err, gc.ErrorMatches, `empty reseller authorization query`)
	c.Assert(auths, gc.HasLen, 0)
	s.httpClient.assertNoRequest(c)
}

type mockHttpClient struct {
	testing.Stub
	status        int
	body          interface{}
	requestMethod string
	requestURL    string
	requestBody   []byte
}

func (m *mockHttpClient) Do(req *http.Request) (*http.Response, error) {
	var err error
	m.requestURL = req.URL.String()
	m.requestMethod = req.Method
	if req.Body != nil {
		m.requestBody, err = ioutil.ReadAll(req.Body)
	}
	data := []byte{}
	if m.body != nil {
		data, err = json.Marshal(m.body)
		if err != nil {
			return nil, errors.Trace(err)
		}
	}
	return &http.Response{
		Status:     http.StatusText(m.status),
		StatusCode: m.status,
		Proto:      "HTTP/1.0",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Body:       ioutil.NopCloser(bytes.NewReader(data)),
	}, m.NextErr()
}

func (m *mockHttpClient) DoWithBody(req *http.Request, body io.ReadSeeker) (*http.Response, error) {
	var err error
	m.requestURL = req.URL.String()
	m.requestMethod = req.Method
	if body != nil {
		m.requestBody, err = ioutil.ReadAll(body)
	}

	data := []byte{}
	if m.body != nil {
		data, err = json.Marshal(m.body)
		if err != nil {
			return nil, errors.Trace(err)
		}
	}
	return &http.Response{
		Status:     http.StatusText(m.status),
		StatusCode: m.status,
		Proto:      "HTTP/1.0",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Body:       ioutil.NopCloser(bytes.NewReader(data)),
	}, m.NextErr()
}

func (m *mockHttpClient) assertRequest(c *gc.C, method, expectedURL, expectedBody interface{}) {
	c.Assert(m.requestMethod, gc.Equals, method)
	c.Assert(m.requestURL, gc.Equals, expectedURL)
	if expectedBody != nil {
		c.Assert(string(m.requestBody), jc.JSONEquals, expectedBody)
	} else {
		c.Assert(len(m.requestBody), gc.Equals, 0)
	}
}

func (m *mockHttpClient) assertNoRequest(c *gc.C) {
	c.Assert(m.requestMethod, gc.Equals, "")
	c.Assert(m.requestURL, gc.Equals, "")
}
