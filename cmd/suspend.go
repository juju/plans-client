// Copyright 2017 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package cmd

import (
	"github.com/juju/cmd"
	"github.com/juju/errors"
	"github.com/juju/gnuflag"
)

const suspendPlanDoc = `
suspend-plan is used to suspend plan for a set of charms
Example
suspend-plan foocorp/free cs:~foocorp/app-0 cs:~foocorp/app-1
	disables deploys of the two specified charms using the foocorp/free plan.
`
const suspendPlanPurpose = "suspends plan for specified charms"

// NewSuspendCommand creates a new command that can
// be used to suspend plans.
func NewSuspendCommand() cmd.Command {
	return &suspendResumeCommand{
		op:      suspendOp,
		name:    "suspend-plan",
		purpose: suspendPlanPurpose,
		doc:     suspendPlanDoc,
	}
}

type operation string

const (
	resumeOp  = operation("resume")
	suspendOp = operation("suspend")
)

// suspendResumeCommand suspends plan for a set of charms.
type suspendResumeCommand struct {
	baseCommand

	op      operation
	name    string
	purpose string
	doc     string

	PlanURL   string
	CharmURLs []string
	All       bool
}

// SetFlags implements Command.SetFlags.
func (c *suspendResumeCommand) SetFlags(f *gnuflag.FlagSet) {
	c.baseCommand.ServiceURL = defaultServiceURL()
	c.baseCommand.SetFlags(f)
	f.BoolVar(&c.All, "all", false, "suspend plan for all charms")
}

// Description returns a one-line description of the command.
func (c *suspendResumeCommand) Description() string {
	return suspendPlanPurpose
}

// Info implements Command.Info.
func (c *suspendResumeCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    c.name,
		Args:    "<plan url> [<charm url>[...<charm url N>]]",
		Purpose: c.purpose,
		Doc:     c.doc,
	}
}

// Init implements Command.Init.
func (c *suspendResumeCommand) Init(args []string) error {
	if !c.All && len(args) < 2 {
		return errors.New("missing plan or charm url")
	} else if c.All && len(args) > 1 {
		return errors.New("cannot use --all and specify charm urls")
	}

	c.PlanURL, c.CharmURLs = args[0], args[1:]
	return nil
}

// Run implements Command.Run.
func (c *suspendResumeCommand) Run(ctx *cmd.Context) error {
	defer c.Close()
	client, cleanup, err := c.NewClient(ctx)
	if err != nil {
		return errors.Annotate(err, "failed to create an http client")
	}
	defer cleanup()
	apiClient, err := newClient(c.ServiceURL, client)
	if err != nil {
		return errors.Annotate(err, "failed to create a plan API client")
	}
	switch c.op {
	case suspendOp:
		return errors.Trace(apiClient.Suspend(c.PlanURL, c.All, c.CharmURLs...))
	case resumeOp:
		return errors.Trace(apiClient.Resume(c.PlanURL, c.All, c.CharmURLs...))
	default:
		return errors.New("unknown operation")
	}
}
