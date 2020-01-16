// Copyright 2017 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package cmd

import (
	"context"
	"strings"

	"github.com/juju/cmd"
	"github.com/juju/errors"
	"github.com/juju/gnuflag"
	"github.com/juju/juju/cmd/output"
	"gopkg.in/yaml.v2"

	"github.com/juju/plans-client/api/wireformat"
)

const attachPlanDoc = `
attach-plan is used to enable a specific plan for a charm
Example
attach-plan some-charm canonical/landscape-default
	enables deploys of the some-charm using the canonical/landscape-default plan.
`

const attachPlanPurpose = "associates the charm with the plan"

var _ cmd.Command = (*AttachCommand)(nil)

// AttachCommand adds a charm to existing plans
type AttachCommand struct {
	baseCommand

	CharmResolver charmResolver

	out       cmd.Output
	PlanURL   string
	CharmURL  string
	IsDefault bool
}

// NewAttachCommand creates a new AttachCommand.
func NewAttachCommand() cmd.Command {
	return &AttachCommand{
		CharmResolver: NewCharmStoreResolver(),
	}
}

// SetFlags implements Command.SetFlags.
func (c *AttachCommand) SetFlags(f *gnuflag.FlagSet) {
	c.baseCommand.ServiceURL = defaultServiceURL()
	c.baseCommand.SetFlags(f)
	c.out.AddFlags(f, "yaml", output.DefaultFormatters)
	f.BoolVar(&c.IsDefault, "default", false, "set this plan as the default for the charm")
}

// Description returns a one-line description of the command.
func (c *AttachCommand) Description() string {
	return attachPlanPurpose
}

// Info implements Command.Info.
func (c *AttachCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "attach-plan",
		Args:    "<charm url> <plan url>",
		Purpose: attachPlanPurpose,
		Doc:     attachPlanDoc,
	}
}

// Init implements Command.Init.
func (c *AttachCommand) Init(args []string) error {
	if len(args) < 2 {
		return errors.New("missing charm and plan url")
	}
	charmURL, planURL, args := args[0], args[1], args[2:]

	if err := cmd.CheckEmpty(args); err != nil {
		return errors.Errorf("unknown command line arguments: " + strings.Join(args, ","))
	}

	ctx, err := cmd.DefaultContext()
	if err != nil {
		return errors.Trace(err)
	}
	c.PlanURL = planURL
	client, cleanup, err := c.NewClient(ctx)
	if err != nil {
		return errors.Annotate(err, "could not create API client")
	}
	defer cleanup()
	resolved, err := c.CharmResolver.Resolve(client, charmURL)
	if err != nil {
		return errors.Annotate(err, "could not resolve charm url")
	}
	// When adding a charm url to a plan it's best to warn the user if the
	// charm url he specified is not fully resolved already.
	if resolved != charmURL {
		return errors.Errorf("charm url %q is not resolved - did you mean %q?", charmURL, resolved)
	}
	c.CharmURL = charmURL
	pID, err := wireformat.ParsePlanIDWithOptionalRevision(c.PlanURL)
	if err != nil {
		return errors.Annotate(err, "failed to parse plan url")
	}
	if pID.Revision != 0 {
		return errors.Errorf("can't attach plan with specific revision, try %q", pID.PlanURL.String())
	}
	return nil
}

// Run implements Command.Run.
func (c *AttachCommand) Run(ctx *cmd.Context) error {
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

	plans, err := apiClient.Get(context.Background(), c.PlanURL)
	if err != nil {
		return errors.Annotatef(err, "failed to retrieve plan %v", c.PlanURL)
	}
	if len(plans) == 0 {
		return errors.Errorf("cannot find a released plan to attach")
	}
	if len(plans) > 1 {
		return errors.Errorf("expected 1 plan, got %d", len(plans))
	}
	plan := plans[0]
	if !plan.Released {
		return errors.Errorf("cannot attach charm to an unreleased plan")
	}
	planMetricNames, err := planMetrics(plan.Definition)
	if err != nil {
		return errors.Trace(err)
	}
	charmMetricNames, err := c.CharmResolver.Metrics(client, c.CharmURL)
	if err != nil {
		return errors.Trace(err)
	}
	if !sameMetrics(planMetricNames, charmMetricNames) {
		return errors.Errorf("plan %v cannot be used to rate charm %v: no common metrics", c.PlanURL, c.CharmURL)
	}

	err = apiClient.AddCharm(context.Background(), c.PlanURL, c.CharmURL, c.IsDefault)
	if err != nil {
		return errors.Annotate(err, "failed to retrieve plans")
	}

	err = c.out.Write(ctx, "OK")
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

type planDefinition struct {
	Metrics map[string]interface{} `yaml:"metrics"`
}

func planMetrics(definition string) ([]string, error) {
	fail := func(err error) ([]string, error) {
		return []string{}, err
	}
	var plan planDefinition
	err := yaml.Unmarshal([]byte(definition), &plan)
	if err != nil {
		fail(errors.Trace(err))
	}
	metrics := []string{}
	for key, _ := range plan.Metrics {
		metrics = append(metrics, key)
	}
	return metrics, nil
}

func sameMetrics(planMetrics, charmMetrics []string) bool {
	for _, pm := range planMetrics {
		for _, cm := range charmMetrics {
			if pm == cm {
				return true
			}
		}
	}
	return false
}
