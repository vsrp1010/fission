/*
Copyright 2018 The Fission Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/urfave/cli"
)

func replay(c *cli.Context) error {
	fc := getClient(c.GlobalString("server"))

	reqUID := c.String("reqUID")
	if len(reqUID) == 0 {
		log.Fatal("Need a reqUID, use --reqUID flag to specify")
	}

	debug := c.Bool("debug")
	if debug {
		// For now all this does is tell the executor it wants a new pod running the appropriate debug image
		// and returns the service IP to the user; the user must manually connect to it through port-forwarding
		svcIP, err := fc.SpecializeDebugPod(reqUID)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("Connect to debug process @%v\n\n", svcIP)
		}
	}

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
