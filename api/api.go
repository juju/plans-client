// Copyright 2016 Canonical Ltd.  All rights reserved.

// Package api defines the plan management API.
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/CanonicalLtd/omniutils"
	"github.com/juju/errors"
	"gopkg.in/macaroon-bakery.v1/httpbakery"
	"gopkg.in/macaroon.v1"

	"github.com/CanonicalLtd/plans-client/api/wireformat"
)

// PlanClient defines the interface available to clients of the plan api.
type PlanClient interface {
	// Save uploads a new plan to the plans service.
	Save(planURL, definition string) (*wireformat.Plan, error)
	// AddCharm associates a charm with the specified plan.
	AddCharm(planURL string, charmURL string, isDefault bool) error
	// Get returns a slice of Plans that match the stated criteria, namely
	// the plan URL, owner of the plan or an associated charm url.
	Get(planURL string) ([]wireformat.Plan, error)
	// GetDefaultPlan returns the default plan associated with the charm.
	GetDefaultPlan(charmURL string) (*wireformat.Plan, error)
	// GetPlansForCharm returns the plans associated with the charm.
	GetPlansForCharm(charmURL string) ([]wireformat.Plan, error)
	// Suspend suspends the plan for specified charms.
	Suspend(planURL string, all bool, charmURLs ...string) error
	// Resume resumes the plan for specified charms.
	Resume(planURL string, all bool, charmURLs ...string) error
	// Release releases the specified plan.
	Release(planURL string) (*wireformat.Plan, error)
	// GetPlanDetails returns detailed information about a plan.
	GetPlanDetails(planURL string) (*wireformat.PlanDetails, error)
	// GetPlanRevisions returns all revision of a plan.
	GetPlanRevisions(planURL string) ([]wireformat.Plan, error)
	// Authorize returns the authorization macaroon for the specified environment, charm url and service name.
	Authorize(environmentUUID, charmURL, serviceName, plan string) (*macaroon.Macaroon, error)
	// AuthorizeReseller returns the reseller authorization macaroon for the specified application.
	AuthorizeReseller(plan, charm, application, applicationOwner, applicationUser string) (*macaroon.Macaroon, error)
	// GetAuthorizations returns a slice of Authorizations that match the
	// criteria specified in the query.
	GetAuthorizations(query wireformat.AuthorizationQuery) ([]wireformat.Authorization, error)
	// GetResellerAuthorizations retuns a slice of reseller Authorizations.
	GetResellerAuthorizations(query wireformat.ResellerAuthorizationQuery) ([]wireformat.ResellerAuthorization, error)
}

type httpClient interface {
	DoWithBody(req *http.Request, body io.ReadSeeker) (*http.Response, error)
	Do(req *http.Request) (*http.Response, error)
}

var _ PlanClient = (*client)(nil)

// client is the implementation of the PlanClient interface.
type client struct {
	plansService string
	client       httpClient
}

// ClientOption defines a function which configures a Client.
type ClientOption func(h *client) error

// HTTPClient returns a function that sets the http client used by the API
// (e.g. if we want to use TLS).
func HTTPClient(c httpClient) func(h *client) error {
	return func(h *client) error {
		h.client = c
		return nil
	}
}

// NewPlanClient returns a new client for plan management.
func NewPlanClient(url string, options ...ClientOption) (*client, error) {
	c := &client{
		plansService: url,
		client:       httpbakery.NewClient(),
	}

	for _, option := range options {
		err := option(c)
		if err != nil {
			return nil, errors.Trace(err)
		}
	}

	return c, nil
}

// Release releases the specified plan.
func (c *client) Release(planURL string) (*wireformat.Plan, error) {
	pID, err := wireformat.ParsePlanID(planURL)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if pID.Revision == 0 {
		return nil, errors.New("must specify the plan revision")
	}
	u, err := url.Parse(fmt.Sprintf("%s/p/%s/%s/%d/release", c.plansService, pID.Owner, pID.Name, pID.Revision))
	if err != nil {
		return nil, errors.Trace(err)
	}
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, errors.Trace(err)
	}

	response, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Annotate(err, "failed to release the plan")
	}
	defer omniutils.DiscardClose(response)

	err = omniutils.UnmarshalError("release plan", response)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var plan wireformat.Plan
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&plan)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &plan, nil
}

// Suspend suspends the plan for specified charms
func (c *client) Suspend(planURL string, all bool, charmURLs ...string) error {
	return c.suspendResume("suspend", planURL, all, charmURLs...)
}

// Resume resumes the plan for specified charms
func (c *client) Resume(planURL string, all bool, charmURLs ...string) error {
	return c.suspendResume("resume", planURL, all, charmURLs...)
}

func (c *client) suspendResume(operation, planURL string, all bool, charmURLs ...string) error {
	pURL, err := wireformat.ParsePlanURL(planURL)
	if err != nil {
		return errors.Trace(err)
	}
	request := struct {
		All    bool     `json:"all"`
		Charms []string `json:"charms"`
	}{
		All:    all,
		Charms: charmURLs,
	}
	data, err := json.Marshal(request)
	if err != nil {
		return errors.Trace(err)
	}
	u, err := url.Parse(fmt.Sprintf("%s/p/%s/%s/%s", c.plansService, pURL.Owner, pURL.Name, operation))
	if err != nil {
		return errors.Trace(err)
	}
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return errors.Trace(err)
	}
	req.Header.Set("Content-Type", "application/json")

	response, err := c.client.DoWithBody(req, bytes.NewReader(data))
	if err != nil {
		return errors.Annotate(err, "failed to resume the plan")
	}
	defer omniutils.DiscardClose(response)

	err = omniutils.UnmarshalError(fmt.Sprintf("%s plan", operation), response)
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

// Save stores the rating plan definition (definition - plan definition yaml) under a
// specified name (planURL).
func (c *client) Save(planURL string, definition string) (*wireformat.Plan, error) {
	_, err := wireformat.ParsePlanURL(planURL)
	if err != nil {
		return nil, errors.Trace(err)
	}

	u, err := url.Parse(c.plansService + "/p")
	if err != nil {
		return nil, errors.Trace(err)
	}
	plan := wireformat.Plan{URL: planURL, Definition: definition}

	payload := &bytes.Buffer{}
	err = json.NewEncoder(payload).Encode(plan)
	if err != nil {
		return nil, errors.Annotate(err, "failed to marshal the plan structure")
	}

	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, errors.Annotate(err, "failed to create a POST request")
	}
	req.Header.Set("Content-Type", "application/json")

	response, err := c.client.DoWithBody(req, bytes.NewReader(payload.Bytes()))
	if err != nil {
		return nil, errors.Annotate(err, "failed to store the plan")
	}

	err = omniutils.UnmarshalError("save plan", response)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var planResult wireformat.Plan
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&planResult)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &planResult, nil
}

// AddCharm adds the specified charm to all plans matching the criteria.
// If uuid is defined, both, the isvname and planname may be empty ("").
func (c *client) AddCharm(planURL string, charmURL string, isDefault bool) error {
	_, err := wireformat.ParsePlanURL(planURL)
	if err != nil {
		return errors.Trace(err)
	}

	u, err := url.Parse(c.plansService + "/charm")
	if err != nil {
		return errors.Trace(err)
	}

	query := struct {
		Plan    string `json:"plan-url"`
		Charm   string `json:"charm-url"`
		Default bool   `json:"default"`
	}{
		Plan:    planURL,
		Charm:   charmURL,
		Default: isDefault,
	}

	payload := &bytes.Buffer{}
	err = json.NewEncoder(payload).Encode(query)
	if err != nil {
		return errors.Annotate(err, "failed to marshal the plan structure")
	}

	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return errors.Annotate(err, "failed to create a POST request")
	}
	req.Header.Set("Content-Type", "application/json")

	response, err := c.client.DoWithBody(req, bytes.NewReader(payload.Bytes()))
	if err != nil {
		return errors.Annotate(err, "failed to update plan")
	}
	defer omniutils.DiscardClose(response)

	err = omniutils.UnmarshalError("update plan", response)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

// Get performs a query on the plans service and returns all matching plans.
func (c *client) Get(planURL string) ([]wireformat.Plan, error) {
	_, err := wireformat.ParsePlanURL(planURL)
	if err != nil {
		return nil, errors.Trace(err)
	}

	u, err := url.Parse(c.plansService + "/p/" + planURL)
	if err != nil {
		return nil, errors.Trace(err)
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, errors.Annotate(err, "failed to create a GET request")
	}

	response, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Annotate(err, "failed to retrieve matching plans")
	}
	defer omniutils.DiscardClose(response)
	err = omniutils.UnmarshalError("retrieve plans", response)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var plans []wireformat.Plan
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&plans)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to unmarshal the response")
	}
	return plans, nil
}

// GetPlanRevisions returns all revisions of a plan.
func (c *client) GetPlanRevisions(plan string) ([]wireformat.Plan, error) {
	planID, err := wireformat.ParsePlanIDWithOptionalRevision(plan)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if planID.Revision != 0 {
		return nil, errors.New("plan revision specified where none was expected")
	}

	u, err := url.Parse(fmt.Sprintf("%s/p/%s/%s/revisions", c.plansService, planID.Owner, planID.Name))
	if err != nil {
		return nil, errors.Trace(err)
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, errors.Annotate(err, "failed to create a GET request")
	}

	response, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Annotate(err, "failed to retrieve plan revisions")
	}
	defer omniutils.DiscardClose(response)
	err = omniutils.UnmarshalError("retrieve plan revisions", response)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var plans []wireformat.Plan
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&plans)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to unmarshal the response")
	}
	return plans, nil
}

// GetDefaultPlan returns the default plan for the specified charm.
func (c *client) GetDefaultPlan(charmURL string) (*wireformat.Plan, error) {
	u, err := url.Parse(c.plansService + "/charm/default")
	if err != nil {
		return nil, errors.Trace(err)
	}
	query := u.Query()
	query.Set("charm-url", charmURL)
	u.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, errors.Annotate(err, "failed to create GET request")
	}
	response, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Annotate(err, "failed to retrieve default plan")
	}
	defer omniutils.DiscardClose(response)

	err = omniutils.UnmarshalError("retrieve default plan", response)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var plan wireformat.Plan
	dec := json.NewDecoder(response.Body)
	err = dec.Decode(&plan)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to unmarshal response")
	}
	return &plan, nil
}

// GetPlansForCharm returns the default plan for the specified charm.
func (c *client) GetPlansForCharm(charmURL string) ([]wireformat.Plan, error) {
	u, err := url.Parse(c.plansService + "/charm")
	if err != nil {
		return nil, errors.Trace(err)
	}
	query := u.Query()
	query.Set("charm-url", charmURL)
	u.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, errors.Annotate(err, "failed to create GET request")
	}
	response, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Annotate(err, "failed to retrieve default plan")
	}
	defer omniutils.DiscardClose(response)

	err = omniutils.UnmarshalError("retrieve associated plans", response)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var plans []wireformat.Plan
	dec := json.NewDecoder(response.Body)
	err = dec.Decode(&plans)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to unmarshal response")
	}
	return plans, nil
}

// GetPlanDetailes returns detailed information about a plan.
func (c *client) GetPlanDetails(planURL string) (*wireformat.PlanDetails, error) {
	query := url.Values{}
	purl, err := wireformat.ParsePlanIDWithOptionalRevision(planURL)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if purl.Revision != 0 {
		query.Add("revision", fmt.Sprintf("%d", purl.Revision))
	}

	u, err := url.Parse(c.plansService + "/p/" + purl.PlanURL.String() + "/details")
	if err != nil {
		return nil, errors.Trace(err)
	}
	u.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, errors.Annotate(err, "failed to create a GET request")
	}

	response, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Annotate(err, "failed to retrieve matching plans")
	}
	defer omniutils.DiscardClose(response)

	err = omniutils.UnmarshalError("retrieve plans", response)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var plan wireformat.PlanDetails
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&plan)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to unmarshal the response")
	}
	return &plan, nil
}

// Authorize implements the AuthorizationClient.Authorize method.
func (c *client) Authorize(environmentUUID, charmURL, serviceName, planURL string) (*macaroon.Macaroon, error) {
	u, err := url.Parse(c.plansService + "/plan/authorize")
	if err != nil {
		return nil, errors.Trace(err)
	}

	auth := wireformat.AuthorizationRequest{
		EnvironmentUUID: environmentUUID,
		CharmURL:        charmURL,
		ServiceName:     serviceName,
		PlanURL:         planURL,
	}

	buff := &bytes.Buffer{}
	encoder := json.NewEncoder(buff)
	err = encoder.Encode(auth)
	if err != nil {
		return nil, errors.Trace(err)
	}

	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, errors.Trace(err)
	}
	req.Header.Set("Content-Type", "application/json")

	response, err := c.client.DoWithBody(req, bytes.NewReader(buff.Bytes()))
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer omniutils.DiscardClose(response)

	err = omniutils.UnmarshalError("authorize plan", response)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var m *macaroon.Macaroon
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&m)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to unmarshal the response")
	}

	return m, nil
}

// GetAuthorizations implements the PlanAuthorizationClient.GetAuthorizations interface.
func (c *client) GetAuthorizations(query wireformat.AuthorizationQuery) ([]wireformat.Authorization, error) {
	u, err := url.Parse(c.plansService + "/plan/authorization")
	if err != nil {
		return nil, errors.Trace(err)
	}
	q := u.Query()
	q.Set("authorization-id", query.AuthorizationID)
	q.Set("user", query.User)
	q.Set("plan-url", query.PlanURL)
	q.Set("env-uuid", query.EnvironmentUUID)
	q.Set("charm-url", query.CharmURL)
	q.Set("service-name", query.ServiceName)
	q.Set("include-plan", strconv.FormatBool(query.IncludePlan))
	q.Set("statement-period", query.StatementPeriod)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, errors.Annotate(err, "failed to create GET request")
	}

	response, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Annotate(err, "failed to retrieve authorizations")
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return []wireformat.Authorization{}, nil
	}
	err = omniutils.UnmarshalError("retrieve authorizations", response)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var auths []wireformat.Authorization
	dec := json.NewDecoder(response.Body)
	err = dec.Decode(&auths)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to unmarshal response")
	}
	return auths, nil
}

// AuthorizeReseller returns the reseller authorization macaroon for the specified application.
func (c *client) AuthorizeReseller(plan, charm, application, applicationOwner, applicationUser string) (*macaroon.Macaroon, error) {
	u, err := url.Parse(c.plansService + "/plan/reseller/authorize")
	if err != nil {
		return nil, errors.Trace(err)
	}

	auth := wireformat.ResellerAuthorizationRequest{
		Plan:             plan,
		CharmURL:         charm,
		Application:      application,
		ApplicationOwner: applicationOwner,
		ApplicationUser:  applicationUser,
	}

	buff := &bytes.Buffer{}
	encoder := json.NewEncoder(buff)
	err = encoder.Encode(auth)
	if err != nil {
		return nil, errors.Trace(err)
	}

	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, errors.Trace(err)
	}
	req.Header.Set("Content-Type", "application/json")

	response, err := c.client.DoWithBody(req, bytes.NewReader(buff.Bytes()))
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer omniutils.DiscardClose(response)

	err = omniutils.UnmarshalError("authorize reseller plan", response)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var m *macaroon.Macaroon
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&m)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to unmarshal the response")
	}

	return m, nil
}

// GetResellerAuthorizations implements the PlanAuthorizationClient.GetResellerAuthorizations interface.
func (c *client) GetResellerAuthorizations(query wireformat.ResellerAuthorizationQuery) ([]wireformat.ResellerAuthorization, error) {
	u, err := url.Parse(fmt.Sprintf("%s/plan/resellers/authorization", c.plansService))
	if err != nil {
		return nil, errors.Trace(err)
	}
	q := u.Query()
	if query.AuthUUID != "" {
		q.Set("auth-uuid", query.AuthUUID)
	}
	if query.User != "" {
		q.Set("user", query.User)
	}
	if query.Application != "" {
		q.Set("application", query.Application)
	}
	if query.Reseller != "" {
		q.Set("reseller", query.Reseller)
	}
	if query.IncludePlan {
		q.Set("include-plan", strconv.FormatBool(query.IncludePlan))
		q.Set("statement-period", query.StatementPeriod)
	}
	if len(q) == 0 {
		return nil, errors.BadRequestf("empty reseller authorization query")
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, errors.Annotate(err, "failed to create GET request")
	}

	response, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Annotate(err, "failed to retrieve authorizations")
	}
	defer response.Body.Close()

	err = omniutils.UnmarshalError("retrieve reseller authorizations", response)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var auths []wireformat.ResellerAuthorization
	dec := json.NewDecoder(response.Body)
	err = dec.Decode(&auths)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to unmarshal response")
	}
	return auths, nil
}
