// Copyright 2017 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package cmd

import (
	"fmt"
	"strings"

	"github.com/juju/cmd"
	"github.com/juju/errors"
	"github.com/juju/gnuflag"
	"gopkg.in/macaroon-bakery.v1/httpbakery"

	"github.com/juju/plans-client/api"
)

const pushDoc = `
push-plan is used to upload a new plan
Examples
push-plan plan.yaml canonical/default
	uploads a new plan owned by canonical under the name default with the
	definition contained in the file plan.yaml
`
const pushPlanPurpose = "push new plan"

var (
	newClient = func(url string, client *httpbakery.Client) (api.PlanClient, error) {
		return api.NewPlanClient(url, api.HTTPClient(client))
	}
)

// NewPushCommand returns a new PushCommand.
func NewPushCommand() cmd.Command {
	return WrapPlugin(&PushCommand{})
}

// PushCommand uploads a new plan to the plans service
type PushCommand struct {
	baseCommand
	out      cmd.Output
	Filename string
	PlanURL  string
}

// SetFlags implements Command.SetFlags.
func (c *PushCommand) SetFlags(f *gnuflag.FlagSet) {
	c.baseCommand.ServiceURL = defaultServiceURL()
	c.baseCommand.SetFlags(f)
}

// Description returns a one-line description of the command.
func (c *PushCommand) Description() string {
	return pushPlanPurpose
}

// Info implements Command.Info.
func (c *PushCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "push-plan",
		Args:    "<filename> <plan url>",
		Purpose: pushPlanPurpose,
		Doc:     pushDoc,
	}
}

// Init reads and verifies the cli arguments for the PlanAddCommang
func (c *PushCommand) Init(args []string) error {
	if len(args) < 2 {
		return errors.New("missing arguments")
	}
	fn, pn, args := args[0], args[1], args[2:]

	if err := cmd.CheckEmpty(args); err != nil {
		return errors.Errorf("unknown command line arguments: " + strings.Join(args, ","))
	}

	c.PlanURL = pn
	c.Filename = fn
	return nil
}

// Run implements Command.Run.
// Uploads a new plan to the plan service
func (c *PushCommand) Run(ctx *cmd.Context) error {
	data, err := readFile(c.Filename)
	if err != nil {
		return errors.Annotatef(err, "could not read the rating plan from file %q", c.Filename)
	}

	client, err := c.NewClient()
	if err != nil {
		return errors.Annotate(err, "failed to create an http client")
	}

	apiClient, err := newClient(c.ServiceURL, client)
	if err != nil {
		return errors.Annotate(err, "failed to create a plan API client")
	}
	plan, err := apiClient.Save(c.PlanURL, string(data))
	if err != nil {
		return errors.Annotate(err, "failed to save the plan")
	}

	fmt.Fprintf(ctx.Stdout, "%v\n", plan.Id)
	return nil
}
