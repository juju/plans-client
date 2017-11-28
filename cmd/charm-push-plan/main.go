// Copyright 2017 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package main

import (
	"fmt"
	"os"

	"github.com/juju/cmd"

	pcmd "github.com/juju/plans-client/cmd"
)

func main() {
	ctx, err := cmd.DefaultContext()
	if err != nil {
		fmt.Printf("failed to get command context: %v\n", err)
		os.Exit(2)
	}
	c := pcmd.NewPushCommand()
	args := os.Args
	os.Exit(cmd.Main(c, ctx, args[1:]))
}
