// Copyright 2016 Canonical Ltd.  All rights reserved.

package cmd

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gosuri/uitable"
	"github.com/juju/cmd"
	"github.com/juju/errors"
	"github.com/juju/gnuflag"

	"github.com/CanonicalLtd/plans-client/api/wireformat"
)

const showPlanDoc = `
show-plan displays detailed information about the plan
Example
show-plan canonical/landscape-default
	returns details of the canonical/landscape-default plan.
`

const showPlanPurpose = "show plan details"

var _ cmd.Command = (*ShowCommand)(nil)

// ShowCommand returns plan details
type ShowCommand struct {
	baseCommand

	out            cmd.Output
	PlanURL        string
	ShowContent    bool
	OnlyDefinition bool
}

// NewShowCommand creates a new ShowCommand.
func NewShowCommand() cmd.Command {
	return WrapPlugin(&ShowCommand{})
}

// SetFlags implements Command.SetFlags.
func (c *ShowCommand) SetFlags(f *gnuflag.FlagSet) {
	c.baseCommand.ServiceURL = defaultServiceURL()
	c.baseCommand.SetFlags(f)
	c.out.AddFlags(f, "tabular", map[string]cmd.Formatter{
		"json":    cmd.FormatJson,
		"yaml":    cmd.FormatYaml,
		"tabular": formatTabular,
	})
	f.BoolVar(&c.ShowContent, "content", false, "show plan definition")
	f.BoolVar(&c.OnlyDefinition, "definition", false, "show only the plan definition")
}

// Description returns a one-line description of the command.
func (c *ShowCommand) Description() string {
	return showPlanPurpose
}

// Info implements Command.Info.
func (c *ShowCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "show-plan",
		Args:    "<plan url>",
		Purpose: showPlanPurpose,
		Doc:     showPlanDoc,
	}
}

// Init implements Command.Init.
func (c *ShowCommand) Init(args []string) error {
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
func (c *ShowCommand) Run(ctx *cmd.Context) error {
	defer c.Close()
	client, err := c.NewClient()
	if err != nil {
		return errors.Annotate(err, "failed to create an http client")
	}
	apiClient, err := newClient(c.ServiceURL, client)
	if err != nil {
		return errors.Annotate(err, "failed to create a plan API client")
	}

	plan, err := apiClient.GetPlanDetails(c.PlanURL)
	if err != nil {
		return errors.Annotatef(err, "failed to retrieve plan %v details", c.PlanURL)
	}

	if c.OnlyDefinition {
		fmt.Fprintf(ctx.Stdout, "%v\n%v", plan.Plan.Id, plan.Plan.Definition)
		return nil
	}

	err = c.out.Write(ctx, fromWire(c.ShowContent, plan))
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func fromWire(showContent bool, plan *wireformat.PlanDetails) *planDetails {
	p := planDetails{
		ID:      plan.Plan.Id,
		Created: eventFromWire(plan.Created),
		Charms:  make([]charmDetails, len(plan.Charms)),
	}
	if showContent {
		p.Definition = plan.Plan.Definition
		p.PlanDescription = plan.Plan.PlanDescription
		p.PlanPrice = plan.Plan.PlanPrice
	}
	if plan.Released != nil {
		e := eventFromWire(*plan.Released)
		p.Released = &e
	}
	for i, ch := range plan.Charms {
		p.Charms[i] = charmDetails{
			CharmURL:       ch.CharmURL,
			Attached:       eventFromWire(ch.Attached),
			EffectiveSince: ch.EffectiveSince,
			Default:        ch.Default,
			Events:         make([]eventDetails, len(ch.Events)),
		}
		for j, e := range ch.Events {
			p.Charms[i].Events[j] = eventFromWire(e)
		}
	}
	return &p
}

type planDetails struct {
	ID              string         `json:"id" yaml:"id"`
	Created         eventDetails   `json:"created" yaml:"created"`
	Released        *eventDetails  `json:"released,omitempty" yaml:"released,omitempty"`
	Definition      string         `json:"definition,omitempty" yaml:"definition,omitempty"`
	PlanDescription string         `json:"description,omitempty" yaml:"description"`
	PlanPrice       string         `json:"price,omitempty" yaml:"price,omitempty"`
	Charms          []charmDetails `json:"charms,omitempty" yaml:"charms,omitempty"`
}

type charmDetails struct {
	CharmURL       string         `json:"charm" yaml:"charm"`
	Attached       eventDetails   `json:"attached" yaml:"attached"`
	EffectiveSince *time.Time     `json:"effective-since,omitempty" yaml:"effective-since,omitempty"`
	Default        bool           `json:"default" yaml:"default"`
	Events         []eventDetails `json:"events,omitempty" yaml:"events,omitempty"`
}

type eventDetails struct {
	User string    `json:"user" yaml:"user"`
	Type string    `json:"type" yaml:"type"`
	Time time.Time `json:"time" yaml:"time"`
}

func eventFromWire(event wireformat.Event) eventDetails {
	return eventDetails{
		User: event.User,
		Type: event.Type,
		Time: event.Time,
	}
}

func formatTabular(w io.Writer, value interface{}) error {
	plan, ok := value.(*planDetails)
	if !ok {
		return errors.Errorf("expected value of type %T, got %T", plan, value)
	}

	table := uitable.New()
	table.MaxColWidth = 50
	table.Wrap = true
	for _, col := range []int{1, 2, 3, 4, 5} {
		table.RightAlign(col)
	}
	table.AddRow("PLAN")
	table.AddRow(plan.ID)
	table.AddRow("", "CREATED BY", "TIME")
	table.AddRow("", plan.Created.User, plan.Created.Time)
	if plan.Released != nil {
		table.AddRow("", "RELEASED BY", "TIME")
		table.AddRow("", plan.Released.User, plan.Released.Time)
	}
	if plan.PlanDescription != "" {
		table.AddRow("", "DESCRIPTION", plan.PlanDescription)
	}
	if plan.PlanPrice != "" {
		table.AddRow("", "PRICE", plan.PlanPrice)
	}
	if plan.Definition != "" {
		table.AddRow("", "DEFINITION", plan.Definition)
	}
	if len(plan.Charms) > 0 {
		table.AddRow("CHARMS")
		for _, charm := range plan.Charms {
			if charm.EffectiveSince != nil {
				table.AddRow("CHARM", "ATTACHED BY", "TIME", "DEFAULT", "EFFECTIVE SINCE")
				table.AddRow(charm.CharmURL, charm.Attached.User, charm.Attached.Time, charm.Default, charm.EffectiveSince)
			} else {
				table.AddRow("CHARM", "ATTACHED BY", "TIME", "DEFAULT", "")
				table.AddRow(charm.CharmURL, charm.Attached.User, charm.Attached.Time, charm.Default, "")
			}
			if len(charm.Events) > 0 {
				table.AddRow("", "EVENTS")
				table.AddRow("", "", "BY", "TYPE", "TIME")
				for _, event := range charm.Events {
					table.AddRow("", "", event.User, event.Type, event.Time)
				}
			}

		}
	}

	_, err := w.Write(table.Bytes())
	if err != nil {
		return errors.Annotatef(err, "failed to print table")
	}
	return nil
}
