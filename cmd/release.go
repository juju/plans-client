// Copyright 2017 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/juju/cmd"
	"github.com/juju/errors"
	"github.com/juju/gnuflag"
)

const releaseDoc = `
release-plan is used to release the specified plan revision
Example
release-plan canonical/foobar/1
	release revision 1 of the canonical/foobar plan
`
const releasePlanPurpose = "release the plan"

// ReleaseCommand adds a charm to existing plans
type ReleaseCommand struct {
	baseCommand
	Plan string
}

// NewReleaseCommand creates a new ReleaseCommand.
func NewReleaseCommand() cmd.Command {
	return &ReleaseCommand{}
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
		Args:    "<plan-revision>",
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
	client, cleanup, err := c.NewClient(ctx)
	if err != nil {
		return errors.Annotate(err, "failed to create an http client")
	}
	defer cleanup()
	apiClient, err := newClient(c.ServiceURL, client)
	if err != nil {
		return errors.Annotate(err, "failed to create a plan API client")
	}
	plan, err := apiClient.Release(context.Background(), c.Plan)
	if err != nil {
		return errors.Trace(err)
	}
	fmt.Fprintln(ctx.Stderr, plan.Id)
	if plan.EffectiveTime != nil {
		fmt.Fprintf(ctx.Stderr, "effective from %v\n", plan.EffectiveTime.Format(time.RFC822))
	}

	return nil
}
