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
	out := fmt.Sprintf(
		"%s\ncommit: %s\ndate: %s\n",
		versionLine(),
		commit,
		date,
	)
	_, err := io.WriteString(streams.Stdout, out)
	return err
}

func versionLine() string {
	return fmt.Sprintf("azkit %s", version)
}
