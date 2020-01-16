// Copyright 2017 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package cmd

import (
	"io/ioutil"
	"os"

	"github.com/juju/cmd"
	"github.com/juju/gnuflag"
	"github.com/juju/idmclient/ussologin"
	"github.com/juju/juju/juju/osenv"
	cookiejar "github.com/juju/persistent-cookiejar"
	"golang.org/x/net/publicsuffix"
	"gopkg.in/juju/environschema.v1/form"
	"gopkg.in/macaroon-bakery.v2/httpbakery"
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
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
		Filename:         cookiejar.DefaultCookieFile(),
	})
	if err != nil {
		return nil, nil, err
	}
	bakeryClient := httpbakery.NewClient()
	bakeryClient.Jar = jar

	if s.NoBrowser {
		tokenStore := ussologin.NewFileTokenStore(ussoTokenPath())
		bakeryClient.AddInteractor(ussologin.NewInteractor(ussologin.StoreTokenGetter{
			Store: tokenStore,
			TokenGetter: ussologin.FormTokenGetter{
				Filler: &form.IOFiller{
					In:  os.Stdin,
					Out: os.Stdout,
				},
				Name: "charm",
			},
		}))
	}
	bakeryClient.AddInteractor(httpbakery.WebBrowserInteractor{})

	return bakeryClient, func() {
		err := jar.Save()
		if err != nil {
			ctx.Warningf("failed to save cookie jar: %v", err)
		}
	}, nil
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

func ussoTokenPath() string {
	return osenv.JujuXDGDataHomePath("store-usso-token")
}
