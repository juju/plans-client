// Copyright 2016 Canonical Ltd.  All rights reserved.

package cmd

const resumePlanDoc = `
resume-plan is used to resume plan for a set of charms
Example
resume-plan foocorp/free cs:~foocorp/app-0 cs:~foocorp/app-1
 	enables deploys of the two specified charms using the foocorp/free plan.
`

// NewResumeCommand creates a new resumeCommand.
func NewResumeCommand() *suspendResumeCommand {
	return &suspendResumeCommand{
		op:      resumeOp,
		name:    "resume-plan",
		purpose: "resumes plan for specified charms",
		doc:     resumePlanDoc,
	}
}
