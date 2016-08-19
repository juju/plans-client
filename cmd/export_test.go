// Copyright 2016 Canonical Ltd.  All rights reserved.

package cmd

var (
	ReadFile  = &readFile
	NewClient = &newClient
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
