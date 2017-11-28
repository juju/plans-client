// Copyright 2017 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/juju/cmd"
	"github.com/juju/gnuflag"
	"github.com/juju/idmclient/ussologin"
	"github.com/juju/juju/juju/osenv"
	"github.com/juju/juju/jujuclient"
	"gopkg.in/juju/environschema.v1/form"
	"gopkg.in/macaroon-bakery.v1/httpbakery"
)

var (
	defaultURL = "https://api.jujucharms.com/omnibus/v2"
	readFile   = ioutil.ReadFile
)

// defaultServiceURL returns the default public URL for plans clients.
func defaultServiceURL() string {
	obURL := os.Getenv("OB_URL")
	if obURL != "" {
		return obURL
	}
	return defaultURL
}

type baseCommand struct {
	cmd.CommandBase

	ServiceURL string
}

// NewClient returns a new http bakery client for Omnibus commands.
func (s *baseCommand) NewClient() (*httpbakery.Client, error) {
	jujuXDGDataHome := osenv.JujuXDGDataHomeDir()
	if jujuXDGDataHome == "" {
		panic("cannot determine juju data home, required model variables are not set")
	}
	osenv.SetJujuXDGDataHome(jujuXDGDataHome)
	client := httpbakery.NewClient()
	filler := &form.IOFiller{
		In:  os.Stdin,
		Out: os.Stdout,
	}
	client.VisitWebPage = ussologin.VisitWebPage(
		"juju",
		&http.Client{},
		filler,
		jujuclient.NewTokenStore(),
	)
	return client, nil
}

// Close saves the persistent cookie jar used by the specified httpbakery.Client.
func (s *baseCommand) Close() error {
	return nil
}

// newBaseCommand creates a new baseCommand with the default service
// url set.
func newBaseCommand() *baseCommand {
	return &baseCommand{
		ServiceURL: defaultServiceURL(),
	}
}

// SetFlag implements the Command interface.
func (c *baseCommand) SetFlags(f *gnuflag.FlagSet) {
	if c.ServiceURL == "" {
		c.ServiceURL = defaultServiceURL()
	}
	f.StringVar(&c.ServiceURL, "url", c.ServiceURL, "host and port of the plans services")
}

type commandWithDescription interface {
	cmd.Command
	Description() string
}

// WrapPlugin returns a wrapped plugin command.
func WrapPlugin(cmd commandWithDescription) cmd.Command {
	return &pluginWrapper{commandWithDescription: cmd}
}

type pluginWrapper struct {
	commandWithDescription
	Description bool
}

// SetFlags implements Command.SetFlags.
func (c *pluginWrapper) SetFlags(f *gnuflag.FlagSet) {
	c.commandWithDescription.SetFlags(f)
	f.BoolVar(&c.Description, "description", false, "returns command description")
}

// Init implements Command.Init.
func (c *pluginWrapper) Init(args []string) error {
	if c.Description {
		return nil
	}
	return c.commandWithDescription.Init(args)
}

// Run implements Command.Run.
func (c *pluginWrapper) Run(ctx *cmd.Context) error {
	if c.Description {
		fmt.Fprint(ctx.Stdout, c.commandWithDescription.Description())
		return nil
	}
	return c.commandWithDescription.Run(ctx)
}
