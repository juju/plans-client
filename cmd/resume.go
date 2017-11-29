// Copyright 2017 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package cmd

import (
	"github.com/juju/cmd"
)

const resumePlanDoc = `
resume-plan is used to resume plan for a set of charms
Example
resume-plan foocorp/free cs:~foocorp/app-0 cs:~foocorp/app-1
 	enables deploys of the two specified charms using the foocorp/free plan.
`
const resumePlanPurpose = "resumes plan for specified charms"

// NewResumeCommand creates a new resumeCommand.
func NewResumeCommand() cmd.Command {
	return &suspendResumeCommand{
		op:      resumeOp,
		name:    "resume-plan",
		purpose: resumePlanPurpose,
		doc:     resumePlanDoc,
	}
}
