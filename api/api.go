// Copyright 2017 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

// Package api defines the plan management API.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/juju/errors"
	"gopkg.in/macaroon-bakery.v2/httpbakery"
	"gopkg.in/macaroon.v1"

	"github.com/juju/plans-client/api/wireformat"
)

// PlanClient defines the interface available to clients of the plan api.
type PlanClient interface {
	// Save uploads a new plan to the plans service.
	Save(ctx context.Context, planURL, definition string) (*wireformat.Plan, error)
	// AddCharm associates a charm with the specified plan.
	AddCharm(ctx context.Context, planURL string, charmURL string, isDefault bool) error
	// Get returns a slice of Plans that match the stated criteria, namely
	// the plan URL, owner of the plan or an associated charm url.
	Get(ctx context.Context, planURL string) ([]wireformat.Plan, error)
	// GetPlans returns a slice of plans owned by user or group.
	GetPlans(ctx context.Context, owner string) ([]wireformat.Plan, error)
	// GetDefaultPlan returns the default plan associated with the charm.
	GetDefaultPlan(ctx context.Context, charmURL string) (*wireformat.Plan, error)
	// GetPlansForCharm returns the plans associated with the charm.
	GetPlansForCharm(ctx context.Context, charmURL string) ([]wireformat.Plan, error)
	// Suspend suspends the plan for specified charms.
	Suspend(ctx context.Context, planURL string, all bool, charmURLs ...string) error
	// Resume resumes the plan for specified charms.
	Resume(ctx context.Context, planURL string, all bool, charmURLs ...string) error
	// Release releases the specified plan.
	Release(ctx context.Context, planID string) (*wireformat.Plan, error)
	// GetPlanDetails returns detailed information about a plan.
	GetPlanDetails(ctx context.Context, planURL string) (*wireformat.PlanDetails, error)
	// GetPlanRevisions returns all revision of a plan.
	GetPlanRevisions(ctx context.Context, planURL string) ([]wireformat.Plan, error)
	// Authorize returns the authorization macaroon for the specified environment, charm url and service name.
	Authorize(ctx context.Context, environmentUUID, charmURL, serviceName, plan string) (*macaroon.Macaroon, error)
	// AuthorizeReseller returns the reseller authorization macaroon for the specified application.
	AuthorizeReseller(ctx context.Context, plan, charm, application, applicationOwner, applicationUser string) (*macaroon.Macaroon, error)
	// GetAuthorizations returns a slice of Authorizations that match the
	// criteria specified in the query.
	GetAuthorizations(ctx context.Context, query wireformat.AuthorizationQuery) ([]wireformat.Authorization, error)
	// GetResellerAuthorizations retuns a slice of reseller Authorizations.
	GetResellerAuthorizations(ctx context.Context, query wireformat.ResellerAuthorizationQuery) ([]wireformat.ResellerAuthorization, error)
}

// headerName is the name of the header the handler will look for in incoming requests.
const headerName = "X-Request-ID"

func requestWithId(ctx context.Context, req *http.Request) *http.Request {
	id, ok := ctx.Value(headerName).(string)
	if ok {
		req.Header.Set(headerName, id)
	}
	return req
}

func idHeader(response *http.Response) string {
	return response.Header.Get(headerName)
}

type httpClient interface {
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
func (c *client) Release(ctx context.Context, planID string) (*wireformat.Plan, error) {
	pID, err := wireformat.ParsePlanID(planID)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if pID.Revision == 0 {
		return nil, errors.New("must specify the plan revision")
	}
	u, err := url.Parse(fmt.Sprintf("%s/v3/p/%s/%s/%d/release", c.plansService, pID.Owner, pID.Name, pID.Revision))
	if err != nil {
		return nil, errors.Trace(err)
	}
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, errors.Trace(err)
	}
	req = requestWithId(ctx, req)

	response, err := c.client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "refused discharge") {
			return nil, errors.Annotate(err, `release-plan is currently disabled for public use. Please ask in #juju-partners on freenode or email juju@lists.ubuntu.com`)
		}
		return nil, errors.Annotate(err, "failed to release the plan")
	}
	defer discardClose(response)

	err = unmarshalError("release plan", response)
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
func (c *client) Suspend(ctx context.Context, planURL string, all bool, charmURLs ...string) error {
	return c.suspendResume(ctx, "suspend", planURL, all, charmURLs...)
}

// Resume resumes the plan for specified charms
func (c *client) Resume(ctx context.Context, planURL string, all bool, charmURLs ...string) error {
	return c.suspendResume(ctx, "resume", planURL, all, charmURLs...)
}

func (c *client) suspendResume(ctx context.Context, operation, planURL string, all bool, charmURLs ...string) error {
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
	u, err := url.Parse(fmt.Sprintf("%s/v3/p/%s/%s/%s", c.plansService, pURL.Owner, pURL.Name, operation))
	if err != nil {
		return errors.Trace(err)
	}
	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(data))
	if err != nil {
		return errors.Trace(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req = requestWithId(ctx, req)

	response, err := c.client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "refused discharge") {
			return errors.Errorf(`unauthorized to %s plan: please run "charm whoami" to verify you are member of the %q group`, operation, pURL.Owner)
		}

		return errors.Annotatef(err, "failed to %v the plan", operation)
	}
	defer discardClose(response)

	err = unmarshalError(fmt.Sprintf("%s plan", operation), response)
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

// Save stores the rating plan definition (definition - plan definition yaml) under a
// specified name (planURL).
func (c *client) Save(ctx context.Context, planURL string, definition string) (*wireformat.Plan, error) {
	pURL, err := wireformat.ParsePlanURL(planURL)
	if err != nil {
		return nil, errors.Trace(err)
	}

	u, err := url.Parse(c.plansService + "/v3/p")
	if err != nil {
		return nil, errors.Trace(err)
	}
	plan := wireformat.Plan{URL: planURL, Definition: definition}

	payload := &bytes.Buffer{}
	err = json.NewEncoder(payload).Encode(plan)
	if err != nil {
		return nil, errors.Annotate(err, "failed to marshal the plan structure")
	}

	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(payload.Bytes()))
	if err != nil {
		return nil, errors.Annotate(err, "failed to create a POST request")
	}
	req.Header.Set("Content-Type", "application/json")
	req = requestWithId(ctx, req)

	response, err := c.client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "refused discharge") {
			return nil, errors.Errorf(`unauthorized to save the plan: please run "charm whoami" to verify you are member of the %q group`, pURL.Owner)
		}
		return nil, errors.Annotate(err, "failed to save the plan")
	}
	defer discardClose(response)

	err = unmarshalError("save plan", response)
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
func (c *client) AddCharm(ctx context.Context, planURL string, charmURL string, isDefault bool) error {
	pURL, err := wireformat.ParsePlanURL(planURL)
	if err != nil {
		return errors.Trace(err)
	}

	u, err := url.Parse(c.plansService + "/v3/charm")
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

	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(payload.Bytes()))
	if err != nil {
		return errors.Annotate(err, "failed to create a POST request")
	}
	req.Header.Set("Content-Type", "application/json")
	req = requestWithId(ctx, req)

	response, err := c.client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "refused discharge") {
			return errors.Errorf(`unauthorized to add charm: please run "charm whoami" to verify you are member of the %q group`, pURL.Owner)
		}
		return errors.Annotate(err, "failed to add charm")
	}
	defer discardClose(response)

	err = unmarshalError("add charm", response)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

// Get performs a query on the plans service and returns all matching plans.
func (c *client) Get(ctx context.Context, planURL string) ([]wireformat.Plan, error) {
	_, err := wireformat.ParsePlanURL(planURL)
	if err != nil {
		return nil, errors.Trace(err)
	}

	u, err := url.Parse(c.plansService + "/v3/p/" + planURL)
	if err != nil {
		return nil, errors.Trace(err)
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, errors.Annotate(err, "failed to create a GET request")
	}
	req = requestWithId(ctx, req)

	response, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Annotate(err, "failed to retrieve matching plans")
	}
	defer discardClose(response)
	err = unmarshalError("retrieve plans", response)
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

// GetPlans returns a plans owned by the user or group.
func (c *client) GetPlans(ctx context.Context, owner string) ([]wireformat.Plan, error) {
	u, err := url.Parse(c.plansService + "/v3/p/" + owner)
	if err != nil {
		return nil, errors.Trace(err)
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, errors.Annotate(err, "failed to create a GET request")
	}
	req = requestWithId(ctx, req)

	response, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Annotate(err, "failed to retrieve plans")
	}
	defer discardClose(response)
	err = unmarshalError("retrieve plans", response)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var plans []wireformat.Plan
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&plans)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to unmarshal the response")
	}
	sort.Slice(plans, func(i, j int) bool {
		return plans[i].Id > plans[j].Id
	})
	return plans, nil
}

// GetPlanRevisions returns all revisions of a plan.
func (c *client) GetPlanRevisions(ctx context.Context, plan string) ([]wireformat.Plan, error) {
	planID, err := wireformat.ParsePlanIDWithOptionalRevision(plan)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if planID.Revision != 0 {
		return nil, errors.New("plan revision specified where none was expected")
	}

	u, err := url.Parse(fmt.Sprintf("%s/v3/p/%s/%s/revisions", c.plansService, planID.Owner, planID.Name))
	if err != nil {
		return nil, errors.Trace(err)
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, errors.Annotate(err, "failed to create a GET request")
	}
	req = requestWithId(ctx, req)

	response, err := c.client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "refused discharge") {
			return nil, errors.Errorf(`unauthorized to retrieve plan revisions: please run "charm whoami" to verify you are member of the %q group`, planID.Owner)
		}
		return nil, errors.Annotate(err, "failed to retrieve plan revisions")
	}
	defer discardClose(response)
	err = unmarshalError("retrieve plan revisions", response)
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
func (c *client) GetDefaultPlan(ctx context.Context, charmURL string) (*wireformat.Plan, error) {
	u, err := url.Parse(c.plansService + "/v3/charm/default")
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
	req = requestWithId(ctx, req)

	response, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Annotate(err, "failed to retrieve default plan")
	}
	defer discardClose(response)

	err = unmarshalError("retrieve default plan", response)
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
func (c *client) GetPlansForCharm(ctx context.Context, charmURL string) ([]wireformat.Plan, error) {
	u, err := url.Parse(c.plansService + "/v3/charm")
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
	req = requestWithId(ctx, req)

	response, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Annotate(err, "failed to retrieve default plan")
	}
	defer discardClose(response)

	err = unmarshalError("retrieve associated plans", response)
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
func (c *client) GetPlanDetails(ctx context.Context, planURL string) (*wireformat.PlanDetails, error) {
	query := url.Values{}
	purl, err := wireformat.ParsePlanIDWithOptionalRevision(planURL)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if purl.Revision != 0 {
		query.Add("revision", fmt.Sprintf("%d", purl.Revision))
	}

	u, err := url.Parse(c.plansService + "/v3/p/" + purl.PlanURL.String() + "/details")
	if err != nil {
		return nil, errors.Trace(err)
	}
	u.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, errors.Annotate(err, "failed to create a GET request")
	}
	req = requestWithId(ctx, req)

	response, err := c.client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "refused discharge") {
			return nil, errors.Errorf(`unauthorized to retrieve plan details: please run "charm whoami" to verify you are member of the %q group`, purl.Owner)
		}
		return nil, errors.Annotate(err, "failed to retrieve plan details")
	}
	defer discardClose(response)

	err = unmarshalError("retrieve plan details", response)
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
func (c *client) Authorize(ctx context.Context, environmentUUID, charmURL, serviceName, planURL string) (*macaroon.Macaroon, error) {
	u, err := url.Parse(c.plansService + "/v3/plan/authorize")
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

	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(buff.Bytes()))
	if err != nil {
		return nil, errors.Trace(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req = requestWithId(ctx, req)

	response, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer discardClose(response)

	err = unmarshalError("authorize plan", response)
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
func (c *client) GetAuthorizations(ctx context.Context, query wireformat.AuthorizationQuery) ([]wireformat.Authorization, error) {
	u, err := url.Parse(c.plansService + "/v3/plan/authorization")
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
	req = requestWithId(ctx, req)

	response, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Annotate(err, "failed to retrieve authorizations")
	}
	defer discardClose(response)

	if response.StatusCode == http.StatusNotFound {
		return []wireformat.Authorization{}, nil
	}
	err = unmarshalError("retrieve authorizations", response)
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
func (c *client) AuthorizeReseller(ctx context.Context, plan, charm, application, applicationOwner, applicationUser string) (*macaroon.Macaroon, error) {
	u, err := url.Parse(c.plansService + "/v3/plan/reseller/authorize")
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

	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(buff.Bytes()))
	if err != nil {
		return nil, errors.Trace(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req = requestWithId(ctx, req)

	response, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer discardClose(response)

	err = unmarshalError("authorize reseller plan", response)
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
func (c *client) GetResellerAuthorizations(ctx context.Context, query wireformat.ResellerAuthorizationQuery) ([]wireformat.ResellerAuthorization, error) {
	u, err := url.Parse(fmt.Sprintf("%s/v3/plan/resellers/authorization", c.plansService))
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
	req = requestWithId(ctx, req)

	response, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Annotate(err, "failed to retrieve authorizations")
	}
	defer discardClose(response)

	err = unmarshalError("retrieve reseller authorizations", response)
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

func discardClose(response *http.Response) {
	if response == nil || response.Body == nil {
		return
	}
	io.Copy(ioutil.Discard, response.Body)
	response.Body.Close()
}

func unmarshalError(action string, response *http.Response) error {
	if response.StatusCode != http.StatusOK {
		id := idHeader(response)
		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return errors.Errorf("failed to %s: received status code %d [ID:%v]", action, response.StatusCode, id)
		}
		var e struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		err = json.Unmarshal(data, &e)
		if err != nil {
			return errors.Errorf("failed to %v: received status code %d and response %q [ID:%v]", action, response.StatusCode, string(data), id)
		}

		msg := fmt.Sprintf("failed to %v [ID:%v]", action, id)
		retErr := fmt.Errorf(e.Message)

		switch response.StatusCode {
		case http.StatusNotFound:
			return errors.NewNotFound(retErr, msg)
		case http.StatusBadRequest:
			return errors.NewBadRequest(retErr, msg)
		case http.StatusNotImplemented:
			return errors.NewNotImplemented(retErr, msg)
		case http.StatusUnauthorized:
			return errors.NewUnauthorized(retErr, msg)
		case http.StatusConflict:
			return errors.NewAlreadyExists(retErr, msg)
		default:
			return errors.Errorf("failed to %v: %v [code: %v, ID:%v]", action, e.Message, e.Code, id)
		}
	}
	return nil
}
