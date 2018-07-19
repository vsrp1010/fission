package main

import (
	"github.com/urfave/cli"
	"text/tabwriter"
	"os"
	"fmt"
)

// TODO: More options besides by ReqUID
func replay(c *cli.Context) error {
	fc := getClient(c.GlobalString("server"))

	reqUID := c.String("reqUID")
	if len(reqUID) == 0 {
		fatal("Need a reqUID, use --reqUID flag to specify")
	}

	// TODO: There should only be a single response, not []string of responses
	responses, err := fc.ReplayByReqUID(reqUID)
	checkErr(err, "replay records")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	for _, resp := range responses {
		fmt.Fprintf(w, "%v",
			resp,
		)
	}

	w.Flush()

	return nil
}