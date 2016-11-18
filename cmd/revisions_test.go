// Copyright 2017 Canonical Ltd.  All rights reserved.

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

type showRevisionsSuite struct {
	testing.CleanupSuite
	mockAPI *plantesting.MockPlanClient
	stub    *testing.Stub
}

var _ = gc.Suite(&showRevisionsSuite{})

func (s *showRevisionsSuite) SetUpTest(c *gc.C) {
	s.stub = &testing.Stub{}

	s.mockAPI = plantesting.NewMockPlanClient()

	s.PatchValue(cmd.NewClient, func(string, *httpbakery.Client) (api.PlanClient, error) {
		return s.mockAPI, nil
	})
}

func (s *showRevisionsSuite) TestCommand(c *gc.C) {
	t1 := time.Date(2015, 1, 1, 1, 0, 0, 0, time.UTC)
	t2 := time.Date(2015, 3, 1, 1, 0, 0, 0, time.UTC)
	plans := []wireformat.Plan{{
		Id:              "testisv/default/1",
		URL:             "testisv/default",
		Definition:      plantesting.TestPlan,
		CreatedOn:       time.Date(2015, 1, 1, 1, 0, 0, 0, time.UTC).Format(time.RFC3339),
		PlanDescription: "a test plan",
		PlanPrice:       "a test plan price description",
		EffectiveTime:   &t1,
	}, {
		Id:              "testisv/default/2",
		URL:             "testisv/default",
		Definition:      plantesting.TestPlan,
		CreatedOn:       time.Date(2015, 2, 1, 1, 0, 0, 0, time.UTC).Format(time.RFC3339),
		PlanDescription: "a test plan",
		PlanPrice:       "a test plan price description",
		EffectiveTime:   &t2,
	}}
	s.mockAPI.PlanRevisions = plans

	tests := []struct {
		about            string
		args             []string
		charmMetrics     []string
		resolvedCharmURL string
		err              string
		assertStdout     func(*gc.C, string)
		assertOutput     func(*gc.C, string)
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
		about: "error if we specify plan id",
		args:  []string{"testisv/default/1"},
		err:   `.*plan revision specified where none was expected`,
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
			c.Assert(output, jc.JSONEquals, plans)
		},
		assertCalls: func(stub *testing.Stub) {
			stub.CheckCall(c, 0, "GetPlanRevisions", "testisv/default")
		},
	}, {
		about: "everything works - yaml",
		args:  []string{"testisv/default", "--format", "yaml"},
		assertStdout: func(c *gc.C, output string) {
			c.Assert(output, jc.YAMLEquals, plans)
		},
		assertCalls: func(stub *testing.Stub) {
			stub.CheckCall(c, 0, "GetPlanRevisions", "testisv/default")
		},
	}, {
		about: "everything works - tabular",
		args:  []string{"testisv/default", "--format", "tabular"},
		assertStdout: func(c *gc.C, output string) {
			c.Assert(output, gc.Equals, `PLAN             	          CREATED ON	               EFFECTIVE TIME	                                        DEFINITION
testisv/default/1	2015-01-01T01:00:00Z	2015-01-01 01:00:00 +0000 UTC	                                                  
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
                 	                    	                             	                                                  
testisv/default/2	2015-02-01T01:00:00Z	2015-03-01 01:00:00 +0000 UTC	                                                  
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
			stub.CheckCall(c, 0, "GetPlanRevisions", "testisv/default")
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
		testCommand := &cmd.ShowRevisionsCommand{}

		c.Logf("Running test %d %s", i, t.about)
		ctx, err := cmdtesting.RunCommand(c, testCommand, t.args...)
		if t.err != "" {
			c.Assert(err, gc.ErrorMatches, t.err)
		} else {
			c.Assert(err, jc.ErrorIsNil)
		}
		t.assertCalls(s.mockAPI.Stub)

		if ctx != nil {
			if t.assertStdout != nil {
				t.assertStdout(c, cmdtesting.Stdout(ctx))
			} else {
				t.assertOutput(c, cmdtesting.Stdout(ctx))
			}
		}
	}
}
