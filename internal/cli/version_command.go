package cli

import (
	"fmt"
	"io"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

type VersionCmd struct{}

func (c *VersionCmd) Run(streams *Streams) error {
	_, err := io.WriteString(streams.Stdout, fmt.Sprintf("pimctl %s\ncommit: %s\ndate: %s\n", version, commit, date))
	return err
}
