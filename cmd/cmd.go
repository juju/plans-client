// Copyright 2016 Canonical Ltd.  All rights reserved.

package cmd

import (
	"io/ioutil"
	"os"

	"github.com/juju/cmd"
	"github.com/juju/errors"
	"github.com/juju/persistent-cookiejar"
	"gopkg.in/macaroon-bakery.v1/httpbakery"
	"launchpad.net/gnuflag"
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
	cookiejar  *cookiejar.Jar
}

// NewClient returns a new http bakery client for Omnibus commands.
func (s *baseCommand) NewClient() (*httpbakery.Client, error) {
	if s.cookiejar == nil {
		cookieFile := cookiejar.DefaultCookieFile()
		jar, err := cookiejar.New(&cookiejar.Options{
			Filename: cookieFile,
		})
		if err != nil {
			return nil, errors.Trace(err)
		}
		s.cookiejar = jar
	}
	client := httpbakery.NewClient()
	client.Jar = s.cookiejar
	client.VisitWebPage = httpbakery.OpenWebBrowser
	return client, nil
}

// Close saves the persistent cookie jar used by the specified httpbakery.Client.
func (s *baseCommand) Close() error {
	if s.cookiejar != nil {
		return s.cookiejar.Save()
	}
	return nil
}

// NewbaseCommand creates a new baseCommand with the default service
// url set.
func NewbaseCommand() *baseCommand {
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
