// Copyright 2015 Canonical Ltd.

package testing

import (
	"time"

	jujutesting "github.com/juju/testing"

	"github.com/CanonicalLtd/plans-client/api"
	"github.com/CanonicalLtd/plans-client/api/wireformat"
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
}

// NewMockPlanClient returns a new MockPlanClient
func NewMockPlanClient() *MockPlanClient {
	return &MockPlanClient{
		&jujutesting.Stub{},
	}
}

// Release releases the specified plan.
func (m *MockPlanClient) Release(planURL string) (*wireformat.Plan, error) {
	m.MethodCall(m, "Release", planURL)
	p := &wireformat.Plan{
		URL:        "testisv/default",
		Definition: TestPlan,
		CreatedOn:  time.Date(2015, 0, 0, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
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
func (m *MockPlanClient) Save(planURL, definition string) error {
	m.MethodCall(m, "Save", planURL, definition)
	return m.NextErr()
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
		CreatedOn:  time.Date(2015, 0, 0, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
	}
	return p, m.NextErr()
}

func (m *MockPlanClient) GetPlansForCharm(charmURL string) ([]wireformat.Plan, error) {
	m.MethodCall(m, "GetPlansForCharm", charmURL)
	p := []wireformat.Plan{{
		URL:        "testisv/default",
		Definition: TestPlan,
		CreatedOn:  time.Date(2015, 0, 0, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
	}}
	return p, m.NextErr()
}

// Get returns all plans stored in the mock, regardless of the query.
func (m *MockPlanClient) Get(planURL string) ([]wireformat.Plan, error) {
	m.MethodCall(m, "Get", planURL)
	p := wireformat.Plan{
		URL:        planURL,
		Definition: TestPlan,
		CreatedOn:  time.Date(2015, 0, 0, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
	}
	return []wireformat.Plan{p}, nil
}

var _ api.PlanClient = (*MockPlanClient)(nil)
