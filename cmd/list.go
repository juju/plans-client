// Copyright 2017 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package cmd

import (
	"context"
	"strings"

	"github.com/juju/cmd"
	"github.com/juju/errors"
	"github.com/juju/gnuflag"
)

const listPlansDoc = `
list-plans is to list plans owned by a user or group
Examples
list-plans canonical
	lists all plans owned by canonical
`
const listPlansPurpose = "list plans"

// NewListPlansCommand returns a new ListPlansCommand.
func NewListPlansCommand() cmd.Command {
	return &ListPlansCommand{}
}

// ListPlansCommand lists plans owned by the specified owner.
type ListPlansCommand struct {
	baseCommand
	out   cmd.Output
	Owner string
}

// SetFlags implements Command.SetFlags.
func (c *ListPlansCommand) SetFlags(f *gnuflag.FlagSet) {
	c.baseCommand.ServiceURL = defaultServiceURL()
	c.baseCommand.SetFlags(f)
	c.out.AddFlags(f, "tabular", map[string]cmd.Formatter{
		"json":    cmd.FormatJson,
		"yaml":    cmd.FormatYaml,
		"tabular": formatPlansTabular,
	})
}

// Description returns a one-line description of the command.
func (c *ListPlansCommand) Description() string {
	return listPlansPurpose
}

// Info implements Command.Info.
func (c *ListPlansCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "list-plans",
		Args:    "<owner>",
		Purpose: listPlansPurpose,
		Doc:     listPlansDoc,
	}
}

// Init reads and verifies the cli arguments for the PlanAddCommang
func (c *ListPlansCommand) Init(args []string) error {
	if len(args) < 1 {
		return errors.New("missing arguments")
	}
	owner, args := args[0], args[1:]

	if err := cmd.CheckEmpty(args); err != nil {
		return errors.Errorf("unknown command line arguments: " + strings.Join(args, ","))
	}

	c.Owner = owner
	return nil
}

// Run implements Command.Run.
// Uploads a new plan to the plan service
func (c *ListPlansCommand) Run(ctx *cmd.Context) error {
	client, cleanup, err := c.NewClient(ctx)
	if err != nil {
		return errors.Annotate(err, "failed to create an http client")
	}
	defer cleanup()

	apiClient, err := newClient(c.ServiceURL, client)
	if err != nil {
		return errors.Annotate(err, "failed to create a plan API client")
	}
	plans, err := apiClient.GetPlans(context.Background(), c.Owner)
	if err != nil {
		return errors.Annotate(err, "failed to retrieve plans")
	}

	c.out.Write(ctx, plans)
	return nil
}
