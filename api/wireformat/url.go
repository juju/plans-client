// Copyright 2017 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package wireformat

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/juju/errors"
	"gopkg.in/juju/names.v2"
)

// Regular expression for validating plan names.
var validPlanName = regexp.MustCompile(`^[a-z](-?[a-z0-9]+)+$`)

// PlanURL holds the components of a plan url.
type PlanURL struct {
	Owner string
	Name  string
}

// PlanID identifies a plan revision.
type PlanID struct {
	PlanURL
	Revision int
}

// ParsePlanURL converts a plan url in canonical string format into
// a PlanURL.
// Canonical string format is:  owner/name
func ParsePlanURL(url string) (*PlanURL, error) {
	parts := strings.Split(url, "/")
	var planUrl PlanURL
	if len(parts) == 2 {
		planUrl = PlanURL{
			Owner: parts[0],
			Name:  parts[1],
		}
	} else {
		return nil, errors.NotValidf("plan url %q", url)
	}
	if err := planUrl.Validate(); err != nil {
		return nil, errors.Trace(err)
	}
	return &planUrl, nil
}

// ParsePlanID parses the string representation of a plan ID.
// Canonical string format is:  owner/name/revision
func ParsePlanID(id string) (*PlanID, error) {
	parts := strings.Split(id, "/")
	var planId PlanID
	if len(parts) == 3 {
		rev, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, errors.Annotatef(err, "invalid revision format")
		}
		url, err := ParsePlanURL(fmt.Sprintf("%s/%s", parts[0], parts[1]))
		if err != nil {
			return nil, errors.NotValidf("plan id %q", id)
		}
		planId = PlanID{
			PlanURL:  *url,
			Revision: rev,
		}
	} else {
		return nil, errors.NotValidf("plan id %q", id)
	}
	if err := planId.Validate(); err != nil {
		return nil, errors.Trace(err)
	}
	return &planId, nil
}

// ParsePlanIDWithOptionalRevision parses the string representation of a plan ID.
// If revision is not specified, it will be set to 0.
// Canonical string format is:  owner/name/revision
func ParsePlanIDWithOptionalRevision(id string) (*PlanID, error) {
	parts := strings.Split(id, "/")
	var planId PlanID
	if len(parts) == 3 {
		rev, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, errors.Annotatef(err, "invalid revision format")
		}
		url, err := ParsePlanURL(fmt.Sprintf("%s/%s", parts[0], parts[1]))
		if err != nil {
			return nil, errors.NotValidf("plan id %q", id)
		}
		planId = PlanID{
			PlanURL:  *url,
			Revision: rev,
		}
		if err := planId.Validate(); err != nil {
			return nil, errors.Trace(err)
		}
	} else if len(parts) == 2 {
		url, err := ParsePlanURL(fmt.Sprintf("%s/%s", parts[0], parts[1]))
		if err != nil {
			return nil, errors.NotValidf("plan id %q", id)
		}
		planId = PlanID{
			PlanURL:  *url,
			Revision: 0,
		}
	} else {
		return nil, errors.NotValidf("plan id %q", id)
	}
	return &planId, nil
}

// String returns the plan url in canonical string format.
func (p PlanURL) String() string {
	return fmt.Sprintf("%s/%s", p.Owner, p.Name)
}

// Validate returns an error if one of the components of the PlanURL
// is not in acceptable format.
func (p PlanURL) Validate() error {
	if !names.IsValidUser(p.Owner) {
		return errors.NotValidf("plan owner %q", p.Owner)
	}
	if !validPlanName.MatchString(p.Name) {
		return errors.NotValidf("plan name %q", p.Name)
	}
	return nil
}

// Revision generates a PlanID based on the PlanURL and provided revision
// number.
func (p PlanURL) Revision(rev int) PlanID {
	return PlanID{p, rev}
}

// String returns the plan id in canonical string format.
func (p PlanID) String() string {
	return fmt.Sprintf("%s/%s/%d", p.Owner, p.Name, p.Revision)
}

// Validate returns an error if one of the components of the PlanID
// is not in acceptable format.
func (p PlanID) Validate() error {
	if p.Revision <= 0 {
		return errors.Errorf("revision must be greater than 0")
	}
	if err := p.PlanURL.Validate(); err != nil {
		return errors.Annotate(err, "invalid plan id")
	}
	return nil
}
