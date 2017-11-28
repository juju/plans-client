// Copyright 2017 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package cmd

var (
	ReadFile  = &readFile
	NewClient = &newClient
	FromWire  = fromWire
)

// BaseCommand type is exported for test purposes.
type BaseCommand struct {
	*baseCommand
}

// NewBaseCommand returns a new instance
// of the BaseCommand.
func NewBaseCommand() BaseCommand {
	return BaseCommand{&baseCommand{}}
}
