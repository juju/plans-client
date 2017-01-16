// Copyright 2016 Canonical Ltd.

package wireformat_test

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/CanonicalLtd/plans-client/api/wireformat"
)

type URLSuite struct{}

var _ = gc.Suite(&URLSuite{})

func (t *URLSuite) TestPlanURLParsing(c *gc.C) {
	tests := []struct {
		about  string
		url    string
		result *wireformat.PlanURL
		err    string
	}{{
		about:  "url without revision",
		url:    "owner/plan",
		result: &wireformat.PlanURL{Owner: "owner", Name: "plan"},
		err:    "",
	}, {
		about: "only a string",
		url:   "owner",
		err:   `plan url "owner" not valid`,
	}, {
		about: "empty string",
		url:   "",
		err:   `plan url "" not valid`,
	}, {
		about: "extra fields",
		url:   "owner/name/1",
		err:   `plan url "owner/name/1" not valid`,
	}, {
		about: "just a string",
		url:   "bad owner",
		err:   `plan url "bad owner" not valid`,
	}, {
		about: "empty name",
		url:   "owner/",
		err:   `plan name "" not valid`,
	}, {
		about: "lots of spaces",
		url:   "not an/ owner",
		err:   `plan owner "not an" not valid`,
	}, {
		about: "bad name format",
		url:   "owner/not a valid plan",
		err:   `plan name "not a valid plan" not valid`,
	}, {
		about: "nil values",
		url:   "//",
		err:   `plan url "//" not valid`,
	}, {
		about: "all bad",
		url:   "bad owner/bad name",
		err:   `plan owner "bad owner" not valid`,
	}, {
		about: "no owner",
		url:   "//name/6",
		err:   `plan url "//name/6" not valid`,
	}, {
		about: "misplaced slash",
		url:   "bob/name/",
		err:   `plan url "bob/name/" not valid`,
	},
	}
	for i, test := range tests {
		c.Logf("test %d %s", i, test.about)
		p, err := wireformat.ParsePlanURL(test.url)
		if test.err != "" {
			c.Check(p, gc.IsNil)
			c.Check(err, gc.ErrorMatches, test.err)
		} else {
			c.Check(*p, gc.DeepEquals, *test.result)
			c.Check(err, jc.ErrorIsNil)
		}
	}

}

func (t *URLSuite) TestPlanIDParsing(c *gc.C) {
	tests := []struct {
		about  string
		url    string
		result *wireformat.PlanID
		err    string
	}{{
		about: "url without revision",
		url:   "owner/plan",
		err:   `plan id "owner/plan" not valid`,
	}, {
		about:  "url with revision",
		url:    "owner/name/4",
		result: &wireformat.PlanID{PlanURL: wireformat.PlanURL{Owner: "owner", Name: "name"}, Revision: 4},
	}, {
		about: "only a string",
		url:   "owner",
		err:   `plan id "owner" not valid`,
	}, {
		about: "empty string",
		url:   "",
		err:   `plan id "" not valid`,
	}, {
		about: "extra fields",
		url:   "owner/name/1/this",
		err:   `plan id "owner/name/1/this" not valid`,
	}, {
		about: "non numerical revision",
		url:   "owner/name/abc",
		err:   `invalid revision format: strconv.ParseInt: parsing "abc": invalid syntax`,
	}, {
		about: "negative revision",
		url:   "owner/name/-2",
		err:   `revision must be greater than 0`,
	}, {
		about: "just a string",
		url:   "bad owner",
		err:   `plan id "bad owner" not valid`,
	}, {
		about: "empty name",
		url:   "owner//2",
		err:   `plan id "owner//2" not valid`,
	}, {
		about: "lots of spaces",
		url:   "not an/ owner/plan",
		err:   `invalid revision format: strconv.ParseInt: parsing "plan": invalid syntax`,
	}, {
		about: "bad name format",
		url:   "owner/not a valid plan/0",
		err:   `plan id "owner/not a valid plan/0" not valid`,
	}, {
		about: "nil values",
		url:   "//0",
		err:   `plan id "//0" not valid`,
	}, {
		about: "all bad",
		url:   "bad owner/bad name/hah",
		err:   `invalid revision format: strconv.ParseInt: parsing "hah": invalid syntax`,
	}, {
		about: "no owner",
		url:   "//name/6",
		err:   `plan id "//name/6" not valid`,
	}, {
		about: "no revision, extra slash",
		url:   "bob/name/",
		err:   `invalid revision format: strconv.ParseInt: parsing "": invalid syntax`,
	},
	}
	for i, test := range tests {
		c.Logf("test %d %s", i, test.about)
		p, err := wireformat.ParsePlanID(test.url)
		if test.err != "" {
			c.Check(p, gc.IsNil)
			c.Check(err, gc.ErrorMatches, test.err)
		} else {
			c.Check(*p, gc.DeepEquals, *test.result)
			c.Check(err, jc.ErrorIsNil)
		}
	}
}
