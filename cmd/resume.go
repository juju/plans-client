// Copyright 2016 Canonical Ltd.  All rights reserved.

package cmd

import (
	"github.com/juju/cmd"
	"github.com/juju/errors"
	"launchpad.net/gnuflag"
)

const resumePlanDoc = `
resume-plan is used to resume plan for a set of charms
Example
resume-plan foocorp/free cs:~foocorp/app-0 cs:~foocorp/app-1
 	enables deploys of the two specified charms using the foocorp/free plan.
`

// resumeCommand resumes plan for a set of charms.
type resumeCommand struct {
	baseCommand

	PlanURL   string
	CharmURLs []string
	All       bool
}

// NewResumeCommand creates a new resumeCommand.
func NewResumeCommand() *resumeCommand {
	return &resumeCommand{}
}

// SetFlags implements Command.SetFlags.
func (c *resumeCommand) SetFlags(f *gnuflag.FlagSet) {
	c.baseCommand.ServiceURL = defaultServiceURL()
	c.baseCommand.SetFlags(f)
	f.BoolVar(&c.All, "all", false, "resume plan for all charms")
}

// Info implements Command.Info.
func (c *resumeCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "resume-plan",
		Args:    "<plan url> [<charm url>[...<charm url N>]]",
		Purpose: "resumes plan for specified charms",
		Doc:     resumePlanDoc,
	}
}

// Init implements Command.Init.
func (c *resumeCommand) Init(args []string) error {
	if !c.All && len(args) < 2 {
		return errors.New("missing plan or charm url")
	} else if c.All && len(args) > 1 {
		return errors.New("cannot use --all and specify charm urls")
	}

	c.PlanURL, c.CharmURLs = args[0], args[1:]
	return nil
}

// Run implements Command.Run.
func (c *resumeCommand) Run(ctx *cmd.Context) error {
	defer c.Close()
	client, err := c.NewClient()
	if err != nil {
		return errors.Annotate(err, "failed to create an http client")
	}
	apiClient, err := newClient(c.ServiceURL, client)
	if err != nil {
		return errors.Annotate(err, "failed to create a plan API client")
	}
	return errors.Trace(apiClient.Resume(c.PlanURL, c.All, c.CharmURLs...))
}
