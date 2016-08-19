// Copyright 2014 Canonical Ltd.  All rights reserved.

package cmd_test

import (
	gc "gopkg.in/check.v1"
	stdtesting "testing"

	jujucmd "github.com/juju/cmd"
	"github.com/juju/cmd/cmdtesting"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	"launchpad.net/gnuflag"

	"github.com/CanonicalLtd/plans-client/cmd"
)

func TestPackage(t *stdtesting.T) {
	gc.TestingT(t)
}

type HttpClientSuite struct {
	testing.CleanupSuite
	caCert string
}

var _ = gc.Suite(&HttpClientSuite{})

type testCommand struct {
	cmd.FlaggedHttpClientCommand
}

func (c *testCommand) Info() *jujucmd.Info {
	return &jujucmd.Info{Name: "test"}
}

func (c *testCommand) SetFlags(f *gnuflag.FlagSet) {
	c.FlaggedHttpClientCommand.SetFlags(f)
}

func (c *testCommand) Run(ctx *jujucmd.Context) error {
	return nil
}

func (s *HttpClientSuite) TestCmdCommand(c *gc.C) {
	basecmd := &testCommand{}

	var obEndpoint = "https://test.canonical.com"

	_, err := cmdtesting.RunCommand(c, basecmd, "--url", obEndpoint)

	c.Assert(err, jc.ErrorIsNil)
	c.Assert(basecmd.ServiceURL, gc.Equals, obEndpoint)
}

func (s *HttpClientSuite) TestNewClient(c *gc.C) {
	basecmd := &testCommand{}

	var obEndpoint = "https://test.canonical.com"

	_, err := cmdtesting.RunCommand(c, basecmd, "--url", obEndpoint)
	c.Assert(err, jc.ErrorIsNil)

	client, err := basecmd.NewClient()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(client, gc.NotNil)
}

func (s *HttpClientSuite) TestNewClientNoHttps(c *gc.C) {
	basecmd := &testCommand{}

	var obEndpoint = "http://test.canonical.com"

	_, err := cmdtesting.RunCommand(c, basecmd, "--url", obEndpoint)
	c.Assert(err, jc.ErrorIsNil)

	client, err := basecmd.NewClient()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(client.Transport, gc.IsNil)
}
