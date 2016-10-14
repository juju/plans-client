// Copyright 2016 Canonical Ltd.  All rights reserved.

package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/juju/cmd"
	"github.com/juju/errors"
	"launchpad.net/gnuflag"
)

const releaseDoc = `
release-plan is used to release the specified plan
Example
release-plan canonical/foobar
	release the canonical/foobar plan
`
const releasePlanPurpose = "release the plan"

// ReleaseCommand adds a charm to existing plans
type ReleaseCommand struct {
	baseCommand
	Plan string
}

// NewReleaseCommand creates a new ReleaseCommand.
func NewReleaseCommand() cmd.Command {
	return WrapPlugin(&ReleaseCommand{})
}

// SetFlags implements Command.SetFlags.
func (c *ReleaseCommand) SetFlags(f *gnuflag.FlagSet) {
	c.baseCommand.ServiceURL = defaultServiceURL()
	c.baseCommand.SetFlags(f)
}

// Info implements Command.Info.
func (c *ReleaseCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "release-plans",
		Args:    "<plan>",
		Purpose: releasePlanPurpose,
		Doc:     releaseDoc,
	}
}

// Description returns a one-line description of the command.
func (c *ReleaseCommand) Description() string {
	return releasePlanPurpose
}

// Init implements Command.Init.
func (c *ReleaseCommand) Init(args []string) error {
	if len(args) < 1 {
		return errors.New("missing plan")
	}
	c.Plan = args[0]

	if err := cmd.CheckEmpty(args[1:]); err != nil {
		return errors.Errorf("unknown command line arguments: " + strings.Join(args[1:], ","))
	}
	return nil
}

// Run implements Command.Run.
func (c *ReleaseCommand) Run(ctx *cmd.Context) error {
	client, err := c.NewClient()
	if err != nil {
		return errors.Annotate(err, "failed to create an http client")
	}
	apiClient, err := newClient(c.ServiceURL, client)
	if err != nil {
		return errors.Annotate(err, "failed to create a plan API client")
	}
	plan, err := apiClient.Release(c.Plan)
	if err != nil {
		return errors.Trace(err)
	}
	fmt.Fprintln(ctx.Stderr, plan.Id)
	if plan.EffectiveTime != nil {
		fmt.Fprintf(ctx.Stderr, "effective from %v\n", plan.EffectiveTime.Format(time.RFC822))
	}

	return nil
}
