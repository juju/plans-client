// Copyright 2017 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package cmd

import (
	"io"
	"strings"

	"github.com/gosuri/uitable"
	"github.com/juju/cmd"
	"github.com/juju/errors"
	"github.com/juju/gnuflag"

	"github.com/juju/plans-client/api/wireformat"
)

const showPlanRevisionsDoc = `
show-plan-revisions displays all plan revisions
Example
show-plan-revisions canonical/landscape-default
	returns all revisions of the canonical/landscape-default plan.
`

const showPlanRevisionsPurpose = "show all revision of a plan"

var _ cmd.Command = (*ShowRevisionsCommand)(nil)

// ShowRevisionsCommand returns plan details
type ShowRevisionsCommand struct {
	baseCommand

	out     cmd.Output
	PlanURL string
}

// NewShowRevisionsCommand creates a new ShowRevisionsCommand.
func NewShowRevisionsCommand() cmd.Command {
	return &ShowRevisionsCommand{}
}

// SetFlags implements Command.SetFlags.
func (c *ShowRevisionsCommand) SetFlags(f *gnuflag.FlagSet) {
	c.baseCommand.ServiceURL = defaultServiceURL()
	c.baseCommand.SetFlags(f)
	c.out.AddFlags(f, "tabular", map[string]cmd.Formatter{
		"json":    cmd.FormatJson,
		"yaml":    cmd.FormatYaml,
		"tabular": formatPlansTabular,
	})
}

// Description returns a one-line description of the command.
func (c *ShowRevisionsCommand) Description() string {
	return showPlanRevisionsPurpose
}

// Info implements Command.Info.
func (c *ShowRevisionsCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "show-plan-revisions",
		Args:    "<plan url>",
		Purpose: showPlanRevisionsPurpose,
		Doc:     showPlanRevisionsDoc,
	}
}

// Init implements Command.Init.
func (c *ShowRevisionsCommand) Init(args []string) error {
	if len(args) < 1 {
		return errors.New("missing plan url")
	}
	planURL, args := args[0], args[1:]
	c.PlanURL = planURL

	if err := cmd.CheckEmpty(args); err != nil {
		return errors.Errorf("unknown command line arguments: " + strings.Join(args, ","))
	}

	return nil
}

// Run implements Command.Run.
func (c *ShowRevisionsCommand) Run(ctx *cmd.Context) error {
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

	plans, err := apiClient.GetPlanRevisions(c.PlanURL)
	if err != nil {
		return errors.Annotatef(err, "failed to retrieve plan %v revisions", c.PlanURL)
	}

	err = c.out.Write(ctx, plans)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func formatPlansTabular(w io.Writer, value interface{}) error {
	plans, ok := value.([]wireformat.Plan)
	if !ok {
		return errors.Errorf("expected value of type %T, got %T", plans, value)
	}

	table := uitable.New()
	table.MaxColWidth = 50
	table.Wrap = true
	for _, col := range []int{1, 2, 3, 4} {
		table.RightAlign(col)
	}
	table.AddRow("PLAN", "CREATED ON", "EFFECTIVE TIME", "DEFINITION")
	for _, plan := range plans {
		if plan.EffectiveTime != nil {
			table.AddRow(plan.Id, plan.CreatedOn, plan.EffectiveTime, plan.Definition)
		} else {
			table.AddRow(plan.Id, plan.CreatedOn, "", plan.Definition)
		}
	}

	_, err := w.Write(table.Bytes())
	if err != nil {
		return errors.Annotatef(err, "failed to print table")
	}
	return nil
}
