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
	defaultServiceURL = "http://localhost:9080/v1"
	readFile          = ioutil.ReadFile
)

// DefaultServiceURL returns the default public URL for plans clients.
func DefaultServiceURL() string {
	obURL := os.Getenv("OB_URL")
	if obURL != "" {
		return obURL
	}
	return defaultServiceURL
}

type clientCommandBase struct {
	cmd.CommandBase

	cookiejar *cookiejar.Jar
}

// NewClient returns a new http bakery client for Omnibus commands.
func (s *clientCommandBase) NewClient() (*httpbakery.Client, error) {
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
func (s *clientCommandBase) Close() error {
	if s.cookiejar != nil {
		return s.cookiejar.Save()
	}
	return nil
}

// HttpClientCommand implements a command that is capable of instantiating
// a proper http client to communicate with a service.
type HttpClientCommand struct {
	clientCommandBase
	ServiceURL string
}

// NewHttpClientCommand creates a new HttpClientCommand with the default service
// url set.
func NewHttpClientCommand() HttpClientCommand {
	return HttpClientCommand{
		ServiceURL: DefaultServiceURL(),
	}
}

// FlaggedHttpClientCommand represents an HttpClientCommand that
// exposes flags to alter the service URL.
type FlaggedHttpClientCommand struct {
	HttpClientCommand
}

// SetFlag implements the Command interface.
func (c *FlaggedHttpClientCommand) SetFlags(f *gnuflag.FlagSet) {
	if c.ServiceURL == "" {
		c.ServiceURL = DefaultServiceURL()
	}
	f.StringVar(&c.ServiceURL, "url", c.ServiceURL, "host and port of the plans services")
}
