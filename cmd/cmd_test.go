// Copyright 2017 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package cmd_test

import (
	gc "gopkg.in/check.v1"

	jujucmd "github.com/juju/cmd"
	"github.com/juju/cmd/cmdtesting"
	"github.com/juju/gnuflag"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"

	"github.com/juju/plans-client/cmd"
)

type baseCommandSuite struct {
	testing.CleanupSuite
	caCert string
}

var _ = gc.Suite(&baseCommandSuite{})

func newTestCommand() *testCommand {
	return &testCommand{cmd.NewBaseCommand()}
}

type testCommand struct {
	cmd.BaseCommand
}

func (c *testCommand) Info() *jujucmd.Info {
	return &jujucmd.Info{Name: "test"}
}

func (c *testCommand) SetFlags(f *gnuflag.FlagSet) {
	c.BaseCommand.SetFlags(f)
}

func (c *testCommand) Run(ctx *jujucmd.Context) error {
	return nil
}

func (s *baseCommandSuite) TestCmdCommand(c *gc.C) {
	basecmd := newTestCommand()

	var obEndpoint = "https://test.canonical.com"

	_, err := cmdtesting.RunCommand(c, basecmd, "--url", obEndpoint)

	c.Assert(err, jc.ErrorIsNil)
	c.Assert(basecmd.ServiceURL, gc.Equals, obEndpoint)
}

func (s *baseCommandSuite) TestNewClient(c *gc.C) {
	basecmd := newTestCommand()

	var obEndpoint = "https://test.canonical.com"

	_, err := cmdtesting.RunCommand(c, basecmd, "--url", obEndpoint)
	c.Assert(err, jc.ErrorIsNil)

	client, cleanup, err := basecmd.NewClient(cmdtesting.Context(c))
	c.Assert(err, jc.ErrorIsNil)
	defer cleanup()
	c.Assert(client, gc.NotNil)
}

func (s *baseCommandSuite) TestNewClientNoHttps(c *gc.C) {
	basecmd := newTestCommand()

	var obEndpoint = "http://test.canonical.com"

	_, err := cmdtesting.RunCommand(c, basecmd, "--url", obEndpoint)
	c.Assert(err, jc.ErrorIsNil)

	client, cleanup, err := basecmd.NewClient(cmdtesting.Context(c))
	c.Assert(err, jc.ErrorIsNil)
	defer cleanup()
	c.Assert(client.Transport, gc.IsNil)
}
