// Copyright 2014 Canonical Ltd.  All rights reserved.

// wireformat package contains wireformat structs intended for
// plan management API.
package wireformat

import (
	"encoding/json"
	"regexp"
	"time"

	"github.com/juju/errors"
	"github.com/juju/names"
	"github.com/juju/utils"
)

// Regular expression for validating plan owners and plan names.
var planURLComponentRe = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

type CharmPlanDetail struct {
	CharmURL       string     `json:"charm"`
	Attached       Event      `json:"attached"`
	EffectiveSince *time.Time `json:"effective-since,omitempty"`
	Default        bool       `json:"default"`
	Events         []Event    `json:"events"`
}

// PlanDetails defines the wireformat for a plan with details abouts
// historical lifecycle.
type PlanDetails struct {
	Plan     Plan              `json:"plan"`
	Created  Event             `json:"created-event"`
	Released *Event            `json:"released-event,omitempty"`
	Charms   []CharmPlanDetail `json:"charms,omitempty"`
}

// Event defines the wireformat for a backend.event
type Event struct {
	User string    `json:"user"` // user who triggered the event
	Type string    `json:"type"` // type of the event
	Time time.Time `json:"time"` // timestamp
}

// PlanActive structure adds an active field to the plan wireformat.
type PlanActive struct {
	*Plan
	Active bool `json:"active"`
}

// Plan structure is used as a wire format to store information on ISV-created
// rating plan and charm URLs for which the plan is valid.
type Plan struct {
	Id              string      `json:"id"`         // Full id of the plan format
	URL             string      `json:"url"`        // Name of the rating plan
	Definition      string      `json:"plan"`       // The rating plan source
	CreatedOn       string      `json:"created-on"` // When the plan was created - RFC3339 encoded timestamp
	PlanDescription string      `json:"description"`
	PlanPrice       string      `json:"price"`
	Released        bool        `json:"released"`
	EffectiveTime   *time.Time  `json:"effective-time,omitempty"`
	Model           interface{} `json:"model,omitempty"` // The rating plan model
}

// UUIDResponse defines a response that just contains a uuid.
type UUIDResponse struct {
	UUID string `json:"uuid"`
}

// Validate validates the plan and returns any errors if the contents are invalid.
func (p Plan) Validate() error {
	if p.URL == "" {
		return errors.New("empty plan url")
	}
	if _, err := ParsePlanURL(p.URL); err != nil {
		return errors.Trace(err)
	}

	if p.Definition == "" {
		return errors.New("missing plan definition")
	}
	return nil
}

// AuthorizationRequest defines the struct used to request a plan authorization.
type AuthorizationRequest struct {
	EnvironmentUUID string `json:"env-uuid"`
	CharmURL        string `json:"charm-url"`
	ServiceName     string `json:"service-name"`
	PlanURL         string `json:"plan-url"`
}

// TODO(api-compat): update tags above and remove this type when clients are ready.
type authorizationRequestV1 AuthorizationRequest

// UnmarshalJSON implements a transitional json.Unmarshaler to allow
// forward-compatible processing of fields renamed in Juju 2.0.
func (ar *AuthorizationRequest) UnmarshalJSON(data []byte) error {
	v := struct {
		authorizationRequestV1
		ModelUUID       string `json:"model-uuid"`
		ApplicationName string `json:"application"`
	}{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*ar = AuthorizationRequest(v.authorizationRequestV1)
	if ar.EnvironmentUUID == "" {
		ar.EnvironmentUUID = v.ModelUUID
	}
	if ar.ServiceName == "" {
		ar.ServiceName = v.ApplicationName
	}
	return nil
}

// Validate checks the AuthorizationRequest for errors.
func (s AuthorizationRequest) Validate() error {
	if !utils.IsValidUUIDString(s.EnvironmentUUID) {
		return errors.Errorf("invalid environment UUID: %q", s.EnvironmentUUID)
	}
	if s.ServiceName == "" {
		return errors.New("undefined service name")
	}
	if !names.IsValidService(s.ServiceName) {
		return errors.Errorf("invalid service name: %q", s.ServiceName)
	}
	if s.CharmURL == "" {
		return errors.New("undefined charm url")
	}
	if !names.IsValidCharm(s.CharmURL) {
		return errors.Errorf("invalid charm url: %q", s.CharmURL)
	}
	if s.PlanURL == "" {
		return errors.Errorf("undefined plan url")
	}
	return nil
}

// TODO(api-compat): update tags above and remove this type when clients are ready.
type authorizationV1 Authorization

// ResellerAuthorizationRequest defines the struct resellers use to obtain
// authorization credentials.
type ResellerAuthorizationRequest struct {
	Application string `json:"application"`
	// The reseller of the application.
	ApplicationOwner string `json:"application-owner"`
	// User consuming resources provided by the application.
	ApplicationUser string `json:"application-user"`
	Plan            string `json:"plan"`
	CharmURL        string `json:"charm-url"`
}

// Validate checks the ResellerAuthorizationRequest for errors.
func (r ResellerAuthorizationRequest) Validate() error {
	if r.Application == "" {
		return errors.New("application not specified")
	}
	if r.ApplicationOwner == "" {
		return errors.New("application owner not specified")
	}
	if r.ApplicationUser == "" {
		return errors.New("application user not specified")
	}
	if r.Plan == "" {
		return errors.New("plan not specified")
	}
	if r.CharmURL == "" {
		return errors.New("charm url not specified")
	}
	return nil
}

// Authorization defines the struct containing information on an issued request for a plan authorization macaroon.
type Authorization struct {
	AuthorizationID string    `json:"authorization-id"` // TODO(cmars): rename to AuthUUID & auth-uuid
	User            string    `json:"user"`
	PlanURL         string    `json:"plan"`
	EnvironmentUUID string    `json:"env-uuid"`
	CharmURL        string    `json:"charm-url"`
	ServiceName     string    `json:"service-name"`
	CreatedOn       time.Time `json:"created-on"`
	CredentialsID   string    `json:"credentials-id"`
	PlanDefinition  string    `json:"plan-definition,omitempty"`
	PlanID          string    `json:"plan-id,omitempty"`
}

// UnmarshalJSON implements a transitional json.Unmarshaler to allow
// forward-compatible processing of fields renamed in Juju 2.0.
func (a *Authorization) UnmarshalJSON(data []byte) error {
	v := struct {
		authorizationV1
		ModelUUID       string `json:"model-uuid"`
		ApplicationName string `json:"application"`
	}{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*a = Authorization(v.authorizationV1)
	if a.EnvironmentUUID == "" {
		a.EnvironmentUUID = v.ModelUUID
	}
	if a.ServiceName == "" {
		a.ServiceName = v.ApplicationName
	}
	return nil
}

// ResellerAuthorization defines the struct containing information on an issued
// reseller plan authorization.
type ResellerAuthorization struct {
	AuthUUID         string    `json:"auth-uuid"`
	Plan             string    `json:"plan"`
	CharmURL         string    `json:"charm-url"`
	Application      string    `json:"application"`
	ApplicationOwner string    `json:"owner"`
	ApplicationUser  string    `json:"user"`
	Credentials      []byte    `json:"credentials"`
	CreatedOn        time.Time `json:"created-on"`
	PlanDefinition   string    `json:"plan-definition,omitempty"`
	PlanID           string    `json:"plan-id,omitempty"`
}

// AuthorizationQuery defines the struct used to query
// authorization records.
type AuthorizationQuery struct {
	AuthorizationID string `json:"authorization-id"` // TODO(cmars): rename to AuthUUID & auth-uuid
	User            string `json:"user"`
	PlanURL         string `json:"plan"`
	EnvironmentUUID string `json:"env-uuid"`
	CharmURL        string `json:"charm-url"`
	ServiceName     string `json:"service-name"`
	IncludePlan     bool   `json:"include-plan"`
	StatementPeriod string `json:"statement-period"`
}

// TODO(api-compat): update tags above and remove this type when clients are ready.
type authorizationQueryV1 AuthorizationQuery

// UnmarshalJSON implements a transitional json.Unmarshaler to allow
// forward-compatible processing of fields renamed in Juju 2.0.
func (a *AuthorizationQuery) UnmarshalJSON(data []byte) error {
	v := struct {
		authorizationQueryV1
		ModelUUID       string `json:"model-uuid"`
		ApplicationName string `json:"application"`
	}{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*a = AuthorizationQuery(v.authorizationQueryV1)
	if a.EnvironmentUUID == "" {
		a.EnvironmentUUID = v.ModelUUID
	}
	if a.ServiceName == "" {
		a.ServiceName = v.ApplicationName
	}
	return nil
}

// ServicePlanResponse defines the response for getServicePlan.
type ServicePlanResponse struct {
	CurrentPlan    string          `json:"current-plan"`
	AvailablePlans map[string]Plan `json:"available-plans"`
}

// ResellerAuthorizationQuery defines the struct used to query
// reseller authorization records.
type ResellerAuthorizationQuery struct {
	AuthUUID        string `json:"auth-uuid"`
	Application     string `json:"application"`
	Reseller        string `json:"reseller"`
	User            string `json:"user"`
	IncludePlan     bool   `json:"include-plan"`
	StatementPeriod string `json:"statement-period"`
}

// Validate validates the ResellerAuthorizationQuery.
func (q ResellerAuthorizationQuery) Validate() error {
	if q.Reseller == "" {
		if q.AuthUUID != "" {
			return nil
		}
		return errors.BadRequestf("must specify the reseller name")
	}
	return nil
}
