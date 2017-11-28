// Copyright 2017 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package testing

import (
	"time"

	"github.com/juju/errors"
	jujutesting "github.com/juju/testing"
	"gopkg.in/macaroon.v1"

	"github.com/juju/plans-client/api"
	"github.com/juju/plans-client/api/wireformat"
)

var TestPlan = `
# Copyright 2014 Canonical Ltd.  All rights reserved.
    description:
        price: 10USD per unit/month
        text: |
           This is a test plan.
    metrics:
      active-users:
        unit:
          transform: max
          period: hour
          gaps: zero
        price: 0.01
`

// MockPlanClient implements a mock of the plan api client.
type MockPlanClient struct {
	*jujutesting.Stub
	PlanDetails   *wireformat.PlanDetails
	PlanRevisions []wireformat.Plan
	Released      bool
}

// NewMockPlanClient returns a new MockPlanClient
func NewMockPlanClient() *MockPlanClient {
	return &MockPlanClient{
		Stub: &jujutesting.Stub{},
	}
}

// Release releases the specified plan.
func (m *MockPlanClient) Release(planURL string) (*wireformat.Plan, error) {
	m.MethodCall(m, "Release", planURL)
	et := time.Date(2016, 1, 1, 1, 0, 0, 0, time.UTC)
	p := &wireformat.Plan{
		Id:            "testisv/default/1",
		URL:           "testisv/default",
		Definition:    TestPlan,
		CreatedOn:     time.Date(2015, 1, 1, 1, 0, 0, 0, time.UTC).Format(time.RFC3339),
		Released:      true,
		EffectiveTime: &et,
	}
	return p, m.NextErr()
}

// Resume resumes the plan for specified charms.
func (m *MockPlanClient) Resume(planURL string, all bool, charmURLs ...string) error {
	m.MethodCall(m, "Resume", planURL, all, charmURLs)
	return m.NextErr()
}

// Suspend suspends the plan for specified charms.
func (m *MockPlanClient) Suspend(planURL string, all bool, charmURLs ...string) error {
	m.MethodCall(m, "Suspend", planURL, all, charmURLs)
	return m.NextErr()
}

// Save stores the plan in the mock.
func (m *MockPlanClient) Save(planURL, definition string) (*wireformat.Plan, error) {
	m.MethodCall(m, "Save", planURL, definition)
	return &wireformat.Plan{
		Id:         "testisv/default/17",
		URL:        "testisv/default",
		Definition: TestPlan,
		CreatedOn:  time.Date(2015, 1, 1, 1, 0, 0, 0, time.UTC).Format(time.RFC3339),
	}, m.NextErr()
}

// AddCharm adds a charm to an existing plan
func (m *MockPlanClient) AddCharm(plan, charmURL string, isDefault bool) error {
	m.MethodCall(m, "AddCharm", plan, charmURL, isDefault)
	return m.NextErr()
}

func (m *MockPlanClient) GetDefaultPlan(charmURL string) (*wireformat.Plan, error) {
	m.MethodCall(m, "GetDefaultPlan", charmURL)
	p := &wireformat.Plan{
		URL:        "testisv/default",
		Definition: TestPlan,
		CreatedOn:  time.Date(2015, 1, 1, 1, 0, 0, 0, time.UTC).Format(time.RFC3339),
	}
	return p, m.NextErr()
}

func (m *MockPlanClient) GetPlansForCharm(charmURL string) ([]wireformat.Plan, error) {
	m.MethodCall(m, "GetPlansForCharm", charmURL)
	p := []wireformat.Plan{{
		URL:        "testisv/default",
		Definition: TestPlan,
		CreatedOn:  time.Date(2015, 1, 1, 1, 0, 0, 0, time.UTC).Format(time.RFC3339),
	}}
	return p, m.NextErr()
}

// Get returns all plans stored in the mock, regardless of the query.
func (m *MockPlanClient) Get(planURL string) ([]wireformat.Plan, error) {
	m.MethodCall(m, "Get", planURL)
	p := wireformat.Plan{
		URL:        planURL,
		Definition: TestPlan,
		CreatedOn:  time.Date(2015, 1, 1, 1, 0, 0, 0, time.UTC).Format(time.RFC3339),
		Released:   m.Released,
	}
	return []wireformat.Plan{p}, nil
}

// GetPlanRevisions returns all revisions of a plan.
func (m *MockPlanClient) GetPlanRevisions(plan string) ([]wireformat.Plan, error) {
	pID, err := wireformat.ParsePlanIDWithOptionalRevision(plan)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if pID.Revision != 0 {
		return nil, errors.New("plan revision specified where none was expected")
	}
	m.MethodCall(m, "GetPlanRevisions", plan)
	return m.PlanRevisions, nil
}

// GetPlanDetails returns detailed information about a plan.
func (m *MockPlanClient) GetPlanDetails(planURL string) (*wireformat.PlanDetails, error) {
	m.MethodCall(m, "GetPlanDetails", planURL)
	if m.PlanDetails != nil {
		return m.PlanDetails, m.NextErr()
	} else {
		t := time.Date(2015, 1, 1, 1, 0, 0, 0, time.UTC)
		return &wireformat.PlanDetails{
			Plan: wireformat.Plan{
				Id:              planURL + "/1", //TODO (mattyw) Not ideal.
				URL:             planURL,
				Definition:      TestPlan,
				CreatedOn:       time.Date(2015, 1, 1, 1, 0, 0, 0, time.UTC).Format(time.RFC3339),
				PlanDescription: "a test plan",
				PlanPrice:       "a test plan price description",
				EffectiveTime:   &t,
			},
			Created: wireformat.Event{
				User: "jane.jaas",
				Type: "create",
				Time: time.Date(2015, 1, 1, 1, 0, 0, 0, time.UTC),
			},
			Released: &wireformat.Event{
				User: "john.jaas",
				Type: "release",
				Time: time.Date(2016, 1, 1, 1, 0, 0, 0, time.UTC),
			},
			Charms: []wireformat.CharmPlanDetail{{
				CharmURL: "cs:~testisv/charm1-0",
				Attached: wireformat.Event{
					User: "jane.jaas",
					Type: "create",
					Time: time.Date(2015, 1, 1, 1, 0, 0, 0, time.UTC),
				},
				Default: false,
			}, {
				CharmURL: "cs:~testisv/charm2-1",
				Attached: wireformat.Event{
					User: "joe.jaas",
					Type: "create",
					Time: time.Date(2015, 1, 1, 1, 0, 0, 0, time.UTC),
				},
				EffectiveSince: &t,
				Default:        true,
				Events: []wireformat.Event{{
					User: "eve.jaas",
					Type: "suspend",
					Time: time.Date(2015, 1, 1, 1, 2, 3, 0, time.UTC),
				}},
			}},
		}, m.NextErr()
	}
}

// Authorize returns the authorization macaroon for the specified environment, charm url and service name.
func (m *MockPlanClient) Authorize(environmentUUID, charmURL, serviceName, plan string) (*macaroon.Macaroon, error) {
	panic("not implemented")
}

// AuthorizeReseller returns the reseller authorization macaroon for the specified application.
func (m *MockPlanClient) AuthorizeReseller(plan, charm, application, applicationOwner, applicationUser string) (*macaroon.Macaroon, error) {
	panic("not implemented")
}

// GetAuthorizations returns a slice of Authorizations that match the
// criteria specified in the query.
func (m *MockPlanClient) GetAuthorizations(query wireformat.AuthorizationQuery) ([]wireformat.Authorization, error) {
	panic("not implemented")
}

// GetResellerAuthorizations retuns a slice of reseller Authorizations.
func (m *MockPlanClient) GetResellerAuthorizations(query wireformat.ResellerAuthorizationQuery) ([]wireformat.ResellerAuthorization, error) {
	panic("not implemented")
}

var _ api.PlanClient = (*MockPlanClient)(nil)
