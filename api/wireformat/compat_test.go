// Copyright 2016 Canonical Ltd. All rights reserved.

package wireformat_test

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/CanonicalLtd/plans-client/api/wireformat"
)

// TODO(api-compat): remove when clients are ready.
type wireCompatSuite struct{}

var _ = gc.Suite(&wireCompatSuite{})

func (s *wireCompatSuite) TestAuthorizationRequest(c *gc.C) {
	oldJSON := []byte(`{
	"env-uuid": "env-is-model",
	"charm-url": "some-charm",
	"service-name": "service-is-application",
	"plan-url": "some-plan",
	"budget": "some-budget",
	"limit": "some-limit"
}`)
	newJSON := []byte(`{
	"model-uuid": "env-is-model",
	"charm-url": "some-charm",
	"application": "service-is-application",
	"plan-url": "some-plan",
	"budget": "some-budget",
	"limit": "some-limit"
}`)
	var oldWire, newWire wireformat.AuthorizationRequest
	c.Assert(oldWire.UnmarshalJSON(oldJSON), jc.ErrorIsNil)
	c.Assert(newWire.UnmarshalJSON(newJSON), jc.ErrorIsNil)
	c.Assert(oldWire, gc.DeepEquals, newWire)
	c.Assert(oldWire.EnvironmentUUID, gc.Equals, "env-is-model")
	c.Assert(newWire.EnvironmentUUID, gc.Equals, "env-is-model")
	c.Assert(oldWire.ServiceName, gc.Equals, "service-is-application")
	c.Assert(newWire.ServiceName, gc.Equals, "service-is-application")
}

func (s *wireCompatSuite) TestAuthorization(c *gc.C) {
	oldJSON := []byte(`{
	"authorization-id": "some-authorization",
	"user": "some-user",
	"plan": "some-plan",
	"env-uuid": "env-is-model",
	"charm-url": "some-charm",
	"service-name": "service-is-application",
	"created-on": "2016-08-06T12:34:56Z",
	"credentials-id": "some-creds"
}`)
	newJSON := []byte(`{
	"authorization-id": "some-authorization",
	"user": "some-user",
	"plan": "some-plan",
	"model-uuid": "env-is-model",
	"charm-url": "some-charm",
	"application": "service-is-application",
	"created-on": "2016-08-06T12:34:56Z",
	"credentials-id": "some-creds"
}`)
	var oldWire, newWire wireformat.Authorization
	c.Assert(oldWire.UnmarshalJSON(oldJSON), jc.ErrorIsNil)
	c.Assert(newWire.UnmarshalJSON(newJSON), jc.ErrorIsNil)
	c.Assert(oldWire, gc.DeepEquals, newWire)
	c.Assert(oldWire.EnvironmentUUID, gc.DeepEquals, "env-is-model")
	c.Assert(newWire.EnvironmentUUID, gc.DeepEquals, "env-is-model")
	c.Assert(oldWire.ServiceName, gc.DeepEquals, "service-is-application")
	c.Assert(newWire.ServiceName, gc.DeepEquals, "service-is-application")
}

func (s *wireCompatSuite) TestAuthorizationQuery(c *gc.C) {
	oldJSON := []byte(`{
	"authorization-id": "some-auth",
	"user": "some-user",
	"plan": "some-plan",
	"env-uuid": "env-is-model",
	"charm-url": "some-charm",
	"service-name": "service-is-application"
}`)
	newJSON := []byte(`{
	"authorization-id": "some-auth",
	"user": "some-user",
	"plan": "some-plan",
	"model-uuid": "env-is-model",
	"charm-url": "some-charm",
	"application": "service-is-application"
}`)
	var oldWire, newWire wireformat.AuthorizationQuery
	c.Assert(oldWire.UnmarshalJSON(oldJSON), jc.ErrorIsNil)
	c.Assert(newWire.UnmarshalJSON(newJSON), jc.ErrorIsNil)
	c.Assert(oldWire, gc.DeepEquals, newWire)
	c.Assert(oldWire.EnvironmentUUID, gc.DeepEquals, "env-is-model")
	c.Assert(newWire.EnvironmentUUID, gc.DeepEquals, "env-is-model")
	c.Assert(oldWire.ServiceName, gc.DeepEquals, "service-is-application")
	c.Assert(newWire.ServiceName, gc.DeepEquals, "service-is-application")
}
