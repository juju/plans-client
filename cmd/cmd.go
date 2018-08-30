// Copyright 2017 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package cmd

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/juju/cmd"
	"github.com/juju/errors"
	"github.com/juju/gnuflag"
	"github.com/juju/idmclient/ussologin"
	"github.com/juju/juju/juju/osenv"
	"github.com/juju/juju/jujuclient"
	"github.com/juju/persistent-cookiejar"
	"github.com/juju/utils"
	"gopkg.in/juju/environschema.v1/form"
	"gopkg.in/macaroon-bakery.v1/httpbakery"
)

var (
	defaultURL = "https://api.jujucharms.com/omnibus"
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

	// NoBrowser specifies that web-browser-based auth should
	// not be used when authenticating.
	NoBrowser bool
}

// NewClient returns a new http bakery client for Omnibus commands.
func (s *baseCommand) NewClient(ctx *cmd.Context) (*httpbakery.Client, func(), error) {
	jujuXDGDataHome := osenv.JujuXDGDataHomeDir()
	if jujuXDGDataHome == "" {
		return nil, func() {}, errors.Errorf("cannot determine juju data home, required environment variables are not set")
	}
	osenv.SetJujuXDGDataHome(jujuXDGDataHome)
	client := httpbakery.NewClient()
	filler := &form.IOFiller{
		In:  os.Stdin,
		Out: os.Stdout,
	}
	if s.NoBrowser {
		client.VisitWebPage = ussologin.VisitWebPage(
			"juju",
			&http.Client{},
			filler,
			jujuclient.NewTokenStore(),
		)
	} else {
		client.VisitWebPage = httpbakery.OpenWebBrowser
	}
	if jar, err := cookiejar.New(&cookiejar.Options{
		Filename: cookieFile(),
	}); err == nil {
		client.Jar = jar
		return client, func() {
			err := jar.Save()
			if err != nil {
				ctx.Warningf("failed to save cookie jar: %v", err)
			}
		}, nil
	} else {
		ctx.Warningf("failed to create cookie jar")
		return client, func() {}, nil
	}
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
	f.BoolVar(&c.NoBrowser, "B", false, "Do not use web browser for authentication")
	f.BoolVar(&c.NoBrowser, "no-browser-login", false, "")
	if c.ServiceURL == "" {
		c.ServiceURL = defaultServiceURL()
	}
	f.StringVar(&c.ServiceURL, "url", c.ServiceURL, "host and port of the plans services")
}

// cookieFile returns the path to the cookie used to store authorization
// macaroons. The returned value can be overridden by setting the
// JUJU_COOKIEFILE environment variable.
func cookieFile() string {
	if file := os.Getenv("JUJU_COOKIEFILE"); file != "" {
		return file
	}
	return path.Join(utils.Home(), ".go-cookies")
}
