package validate

import (
	"fmt"

	"github.com/urfave/cli"
)

const ExitNum = 2

// ServerArgs validates that the necessary flags are not missing.
func ServerArgs(cltx *cli.Context) error {
	for _, param := range []string{
		"arangodb-pass",
		"arangodb-database",
		"arangodb-user",
		"nats-host",
		"nats-port",
	} {
		if len(cltx.String(param)) == 0 {
			return cli.NewExitError(
				fmt.Sprintf("argument %s is missing", param),
				ExitNum,
			)
		}
	}

	return nil
}
