// Copyright 2017 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package cmd

import (
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
	defaultURL = "https://api.jujucharms.com/omnibus/v3"
	readFile   = ioutil.ReadFile
)

// defaultServiceURL returns the default public URL for plans clients.
func defaultServiceURL() string {
	obURL := os.Getenv("JUJU_PLANS")
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
