// Copyright 2016 Canonical Ltd.  All rights reserved.

package cmd_test

import (
	"time"

	"github.com/juju/cmd/cmdtesting"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/macaroon-bakery.v1/httpbakery"

	"github.com/CanonicalLtd/plans-client/api"
	"github.com/CanonicalLtd/plans-client/api/wireformat"
	"github.com/CanonicalLtd/plans-client/cmd"
	plantesting "github.com/CanonicalLtd/plans-client/testing"
)

type showSuite struct {
	testing.CleanupSuite
	mockAPI *plantesting.MockPlanClient
	stub    *testing.Stub
}

var _ = gc.Suite(&showSuite{})

func (s *showSuite) SetUpTest(c *gc.C) {
	s.stub = &testing.Stub{}

	s.mockAPI = plantesting.NewMockPlanClient()

	s.PatchValue(cmd.NewClient, func(string, *httpbakery.Client) (api.PlanClient, error) {
		return s.mockAPI, nil
	})
	s.PatchValue(cmd.ReadFile, func(string) ([]byte, error) {
		return []byte(plantesting.TestPlan), nil
	})
}

func (s *showSuite) TestCommand(c *gc.C) {
	t := time.Date(2015, 0, 0, 0, 0, 0, 0, time.UTC)
	p := &wireformat.PlanDetails{
		Plan: wireformat.Plan{
			URL:             "testisv/default",
			Definition:      plantesting.TestPlan,
			CreatedOn:       time.Date(2015, 0, 0, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			PlanDescription: "a test plan",
			PlanPrice:       "a test plan price description",
		},
		Created: wireformat.Event{
			User: "jane.jaas",
			Type: "create",
			Time: time.Date(2015, 0, 0, 0, 0, 0, 0, time.UTC),
		},
		Released: &wireformat.Event{
			User: "jane.jaas",
			Type: "release",
			Time: time.Date(2015, 0, 0, 0, 0, 0, 0, time.UTC),
		},
		Charms: []wireformat.CharmPlanDetail{{
			CharmURL: "cs:~testisv/charm1-0",
			Attached: wireformat.Event{
				User: "jane.jaas",
				Type: "create",
				Time: time.Date(2015, 0, 0, 0, 0, 0, 0, time.UTC),
			},
			Default: false,
		}, {
			CharmURL: "cs:~testisv/charm2-1",
			Attached: wireformat.Event{
				User: "joe.jaas",
				Type: "create",
				Time: time.Date(2015, 0, 0, 0, 0, 0, 0, time.UTC),
			},
			EffectiveSince: &t,
			Default:        true,
			Events: []wireformat.Event{{
				User: "eve.jaas",
				Type: "suspend",
				Time: time.Date(2015, 0, 0, 1, 2, 3, 0, time.UTC),
			}},
		}},
	}

	tests := []struct {
		about            string
		args             []string
		charmMetrics     []string
		resolvedCharmURL string
		err              string
		assertStdout     func(*gc.C, string)
		assertCalls      func(*testing.Stub)
	}{{
		about: "unrecognized args causes error",
		args:  []string{"testisv/default", "some-arg"},
		err:   `unknown command line arguments: some-arg`,
		assertStdout: func(c *gc.C, output string) {
			c.Assert(output, gc.Equals, "")
		},
		assertCalls: func(stub *testing.Stub) {
			stub.CheckNoCalls(c)
		},
	}, {
		about: "everything works - json",
		args:  []string{"testisv/default", "--format", "json"},
		assertStdout: func(c *gc.C, output string) {
			c.Assert(output, jc.JSONEquals, cmd.FromWire(false, p))
		},
		assertCalls: func(stub *testing.Stub) {
			stub.CheckCall(c, 0, "GetPlanDetails", "testisv/default")
		},
	}, {
		about: "everything works - yaml",
		args:  []string{"testisv/default", "--format", "yaml"},
		assertStdout: func(c *gc.C, output string) {
			c.Assert(output, jc.YAMLEquals, cmd.FromWire(false, p))
		},
		assertCalls: func(stub *testing.Stub) {
			stub.CheckCall(c, 0, "GetPlanDetails", "testisv/default")
		},
	}, {
		about: "everything works - json - content",
		args:  []string{"testisv/default", "--format", "json", "--content"},
		assertStdout: func(c *gc.C, output string) {
			c.Assert(output, jc.JSONEquals, cmd.FromWire(true, p))
		},
		assertCalls: func(stub *testing.Stub) {
			stub.CheckCall(c, 0, "GetPlanDetails", "testisv/default")
		},
	}, {
		about: "everything works - yaml - content",
		args:  []string{"testisv/default", "--format", "yaml", "--content"},
		assertStdout: func(c *gc.C, output string) {
			c.Assert(output, jc.YAMLEquals, cmd.FromWire(true, p))
		},
		assertCalls: func(stub *testing.Stub) {
			stub.CheckCall(c, 0, "GetPlanDetails", "testisv/default")
		},
	}, {
		about: "everything works - tabular",
		args:  []string{"testisv/default"},
		assertStdout: func(c *gc.C, output string) {
			c.Assert(output, gc.Equals, `PLAN                
testisv/default     
                    	 CREATED BY	                         TIME
                    	  jane.jaas	2014-11-30 00:00:00 +0000 UTC
                    	RELEASED BY	                         TIME
                    	  jane.jaas	2014-11-30 00:00:00 +0000 UTC
CHARMS              
CHARM               	ATTACHED BY	                         TIME	DEFAULT	                             
cs:~testisv/charm1-0	  jane.jaas	2014-11-30 00:00:00 +0000 UTC	  false	                             
CHARM               	ATTACHED BY	                         TIME	DEFAULT	              EFFECTIVE SINCE
cs:~testisv/charm2-1	   joe.jaas	2014-11-30 00:00:00 +0000 UTC	   true	2014-11-30 00:00:00 +0000 UTC
                    	     EVENTS
                    	           	                           BY	   TYPE	                         TIME
                    	           	                     eve.jaas	suspend	2014-11-30 01:02:03 +0000 UTC
`)
		},
		assertCalls: func(stub *testing.Stub) {
			stub.CheckCall(c, 0, "GetPlanDetails", "testisv/default")
		},
	}, {
		about: "everything works - tabular - content",
		args:  []string{"testisv/default", "--content"},
		assertStdout: func(c *gc.C, output string) {
			c.Assert(output, gc.Equals, `PLAN                
testisv/default     
                    	 CREATED BY	                                              TIME
                    	  jane.jaas	                     2014-11-30 00:00:00 +0000 UTC
                    	RELEASED BY	                                              TIME
                    	  jane.jaas	                     2014-11-30 00:00:00 +0000 UTC
                    	DESCRIPTION	                                       a test plan
                    	      PRICE	                     a test plan price description
                    	 DEFINITION	                                                  
                    	           	# Copyright 2014 Canonical Ltd.  All rights       
                    	           	reserved.                                         
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
                    	           	                                                  
CHARMS              
CHARM               	ATTACHED BY	                                              TIME	DEFAULT	                             
cs:~testisv/charm1-0	  jane.jaas	                     2014-11-30 00:00:00 +0000 UTC	  false	                             
CHARM               	ATTACHED BY	                                              TIME	DEFAULT	              EFFECTIVE SINCE
cs:~testisv/charm2-1	   joe.jaas	                     2014-11-30 00:00:00 +0000 UTC	   true	2014-11-30 00:00:00 +0000 UTC
                    	     EVENTS
                    	           	                                                BY	   TYPE	                         TIME
                    	           	                                          eve.jaas	suspend	2014-11-30 01:02:03 +0000 UTC
`)
		},
		assertCalls: func(stub *testing.Stub) {
			stub.CheckCall(c, 0, "GetPlanDetails", "testisv/default")
		},
	}, {
		about: "missing args",
		args:  []string{},
		err:   `missing plan url`,
		assertCalls: func(stub *testing.Stub) {
			stub.CheckNoCalls(c)
		},
	},
	}

	for i, t := range tests {
		s.mockAPI.ResetCalls()
		testCommand := &cmd.ShowCommand{}

		c.Logf("Running test %d %s", i, t.about)
		ctx, err := cmdtesting.RunCommand(c, testCommand, t.args...)
		if t.err != "" {
			c.Assert(err, gc.ErrorMatches, t.err)
		} else {
			c.Assert(err, jc.ErrorIsNil)
		}
		t.assertCalls(s.mockAPI.Stub)

		if ctx != nil {
			t.assertStdout(c, cmdtesting.Stdout(ctx))
		}
	}
}

func (s *showSuite) TestCommandWithUnreleasedPlan(c *gc.C) {
	p := &wireformat.PlanDetails{
		Plan: wireformat.Plan{
			URL:             "testisv/default",
			Definition:      plantesting.TestPlan,
			CreatedOn:       time.Date(2015, 0, 0, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			PlanDescription: "a test plan",
			PlanPrice:       "a test plan price description",
		},
		Created: wireformat.Event{
			User: "jane.jaas",
			Type: "create",
			Time: time.Date(2015, 0, 0, 0, 0, 0, 0, time.UTC),
		},
	}

	tests := []struct {
		about            string
		args             []string
		charmMetrics     []string
		resolvedCharmURL string
		err              string
		assertStdout     func(*gc.C, string)
		assertCalls      func(*testing.Stub)
	}{{
		about: "everything works - json",
		args:  []string{"testisv/default", "--format", "json"},
		assertStdout: func(c *gc.C, output string) {
			c.Assert(output, jc.JSONEquals, cmd.FromWire(false, p))
		},
		assertCalls: func(stub *testing.Stub) {
			stub.CheckCall(c, 0, "GetPlanDetails", "testisv/default")
		},
	}, {
		about: "everything works - yaml",
		args:  []string{"testisv/default", "--format", "yaml"},
		assertStdout: func(c *gc.C, output string) {
			c.Assert(output, jc.YAMLEquals, cmd.FromWire(false, p))
		},
		assertCalls: func(stub *testing.Stub) {
			stub.CheckCall(c, 0, "GetPlanDetails", "testisv/default")
		},
	}, {
		about: "everything works - json - content",
		args:  []string{"testisv/default", "--format", "json", "--content"},
		assertStdout: func(c *gc.C, output string) {
			c.Assert(output, jc.JSONEquals, cmd.FromWire(true, p))
		},
		assertCalls: func(stub *testing.Stub) {
			stub.CheckCall(c, 0, "GetPlanDetails", "testisv/default")
		},
	}, {
		about: "everything works - yaml - content",
		args:  []string{"testisv/default", "--format", "yaml", "--content"},
		assertStdout: func(c *gc.C, output string) {
			c.Assert(output, jc.YAMLEquals, cmd.FromWire(true, p))
		},
		assertCalls: func(stub *testing.Stub) {
			stub.CheckCall(c, 0, "GetPlanDetails", "testisv/default")
		},
	}, {
		about: "everything works - tabular",
		args:  []string{"testisv/default"},
		assertStdout: func(c *gc.C, output string) {
			c.Assert(output, gc.Equals, `PLAN           
testisv/default
               	CREATED BY	                         TIME
               	 jane.jaas	2014-11-30 00:00:00 +0000 UTC
`)
		},
		assertCalls: func(stub *testing.Stub) {
			stub.CheckCall(c, 0, "GetPlanDetails", "testisv/default")
		},
	}, {
		about: "everything works - tabular - content",
		args:  []string{"testisv/default", "--content"},
		assertStdout: func(c *gc.C, output string) {
			c.Assert(output, gc.Equals, `PLAN           
testisv/default
               	 CREATED BY	                                              TIME
               	  jane.jaas	                     2014-11-30 00:00:00 +0000 UTC
               	DESCRIPTION	                                       a test plan
               	      PRICE	                     a test plan price description
               	 DEFINITION	                                                  
               	           	# Copyright 2014 Canonical Ltd.  All rights       
               	           	reserved.                                         
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
               	           	                                                  
`)
		},
		assertCalls: func(stub *testing.Stub) {
			stub.CheckCall(c, 0, "GetPlanDetails", "testisv/default")
		},
	}, {
		about: "missing args",
		args:  []string{},
		err:   `missing plan url`,
		assertCalls: func(stub *testing.Stub) {
			stub.CheckNoCalls(c)
		},
	},
	}

	for i, t := range tests {
		s.mockAPI.ResetCalls()
		s.mockAPI.PlanDetails = p
		testCommand := &cmd.ShowCommand{}

		c.Logf("Running test %d %s", i, t.about)
		ctx, err := cmdtesting.RunCommand(c, testCommand, t.args...)
		if t.err != "" {
			c.Assert(err, gc.ErrorMatches, t.err)
		} else {
			c.Assert(err, jc.ErrorIsNil)
		}
		t.assertCalls(s.mockAPI.Stub)

		if ctx != nil {
			t.assertStdout(c, cmdtesting.Stdout(ctx))
		}
	}
}
