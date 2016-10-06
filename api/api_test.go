// Copyright 2014 Canonical Ltd.  All rights reserved.

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
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/CanonicalLtd/plans-client/api"
	"github.com/CanonicalLtd/plans-client/api/wireformat"
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

	client, err := api.NewPlanClient("", api.HTTPClient(s.httpClient))
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

	s.httpClient.assertRequest(c, "POST", "/p", wireformat.Plan{
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
	c.Assert(err, gc.ErrorMatches, `failed to store the plan: silly error \[bad request\]`)
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
	s.httpClient.assertRequest(c, "POST", "/p/testisv/default/1/release", nil)
}

func (s *clientIntegrationSuite) TestReleaseInvalidPlanURL(c *gc.C) {
	_, err := s.planClient.Release("invalid/format/testisv/0/default")
	c.Assert(err, gc.ErrorMatches, "invalid plan url format")
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
	c.Assert(err, gc.ErrorMatches, `failed to release the plan: silly error \[bad request\]`)
}

func (s *clientIntegrationSuite) TestSuspend(c *gc.C) {
	s.httpClient.status = http.StatusOK

	err := s.planClient.Suspend("testisv/default", false, "testisv/plan1", "testisv/plan2")
	c.Assert(err, jc.ErrorIsNil)
	s.httpClient.assertRequest(c, "POST", "/p/testisv/default/suspend", struct {
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
	s.httpClient.assertRequest(c, "POST", "/p/testisv/default/suspend", struct {
		All    bool     `json:"all"`
		Charms []string `json:"charms"`
	}{
		All: true,
	})
}

func (s *clientIntegrationSuite) TestSuspendInvalidPlanURL(c *gc.C) {
	err := s.planClient.Suspend("invalid/format/testisv/default", false, "cs:~testers/charm1-0")
	c.Assert(err, gc.ErrorMatches, "invalid plan url format")
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
	c.Assert(err, gc.ErrorMatches, `failed to suspend the plan: silly error \[bad request\]`)
}

func (s *clientIntegrationSuite) TestResume(c *gc.C) {
	s.httpClient.status = http.StatusOK

	err := s.planClient.Resume("testisv/default", false, "testisv/plan1", "testisv/plan2")
	c.Assert(err, jc.ErrorIsNil)
	s.httpClient.assertRequest(c, "POST", "/p/testisv/default/resume", struct {
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
	s.httpClient.assertRequest(c, "POST", "/p/testisv/default/resume", struct {
		All    bool     `json:"all"`
		Charms []string `json:"charms"`
	}{
		All: true,
	})
}

func (s *clientIntegrationSuite) TestResumeInvalidPlanURL(c *gc.C) {
	err := s.planClient.Resume("invalid/format/testisv/default", false, "cs:~testers/charm1-0")
	c.Assert(err, gc.ErrorMatches, "invalid plan url format")
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
	c.Assert(err, gc.ErrorMatches, `failed to resume the plan: silly error \[bad request\]`)
}

func (s *clientIntegrationSuite) TestAddCharm(c *gc.C) {
	s.httpClient.status = http.StatusOK

	err := s.planClient.AddCharm("testisv/default", "cs:~testers/charm1-0", true)
	c.Assert(err, jc.ErrorIsNil)
	s.httpClient.assertRequest(c, "POST", "/charm", struct {
		Plan    string `json:"plan-url"`
		Charm   string `json:"charm-url"`
		Default bool   `json:"default"`
	}{
		Plan:    "testisv/default",
		Charm:   "cs:~testers/charm1-0",
		Default: true,
	})
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
	c.Assert(err, gc.ErrorMatches, `failed to update the plan: silly error \[bad request\]`)
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
	s.httpClient.assertRequest(c, "GET", "/p/testisv/default", nil)
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
	c.Assert(err, gc.ErrorMatches, `failed to retrieve matching plans: silly error \[bad request\]`)
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
	s.httpClient.assertRequest(c, "GET", "/charm/default?charm-url="+url.QueryEscape("cs:~testers/charm1-0"), nil)
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
	c.Assert(err, gc.ErrorMatches, `failed to retrieve default plan: silly error \[bad request\]`)
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
	s.httpClient.assertRequest(c, "GET", "/charm?charm-url="+url.QueryEscape("cs:~testers/charm1-0"), nil)
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
	c.Assert(err, gc.ErrorMatches, `failed to retrieve associated plans: silly error \[bad request\]`)
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

	s.httpClient.assertRequest(c, "GET", "/p/testisv/default/details", nil)
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

	s.httpClient.assertRequest(c, "GET", "/p/testisv/default/details?revision=7", nil)
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
	c.Assert(err, gc.ErrorMatches, `failed to retrieve plan details: silly error \[bad request\]`)
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
	c.Assert(err, gc.ErrorMatches, "testisv/default not found")
	c.Assert(errors.IsNotFound(err), jc.IsTrue)
}

type mockHttpClient struct {
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
	}, nil
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
	}, nil
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
