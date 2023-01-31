package main

import (
	"flag"
	"os"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Flags struct {
	*flag.FlagSet
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewFlags(name string, args []string) (*Flags, error) {
	flags := new(Flags)
	flags.FlagSet = flag.NewFlagSet(name, flag.ContinueOnError)

	// Register flags
	flags.Bool("debug", false, "Enable debug logging")
	flags.String("token", "${GOSERVER_TOKEN}", "Bearer token")
	flags.String("url", "${GOSERVER_ENDPOINT}", "Endpoint URL")
	flags.Duration("timeout", 0, "Timeout")
	flags.String("router", "/api/gateway/v1", "Router API path")

	// Parse command line
	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	// Return success
	return flags, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (flags *Flags) IsDebug() bool {
	return flags.Lookup("debug").Value.(flag.Getter).Get().(bool)
}

func (flags *Flags) Token() string {
	return os.ExpandEnv(flags.Lookup("token").Value.(flag.Getter).Get().(string))
}

func (flags *Flags) URL() string {
	return os.ExpandEnv(flags.Lookup("url").Value.(flag.Getter).Get().(string))
}

func (flags *Flags) Timeout() time.Duration {
	return flags.Lookup("timeout").Value.(flag.Getter).Get().(time.Duration)
}

func (flags *Flags) Router() string {
	return flags.Lookup("router").Value.(flag.Getter).Get().(string)
}
