package main

import (
	"github.com/fission/fission/redis"
	"github.com/urfave/cli"
)

func recordsView(c *cli.Context) error {
	// err := redis.FilterByTime(90.00)
	rc := redis.MakeRecordsClient()

	err := rc.FilterByFunction("hello")

	checkErr(err, "records view")

	return nil
}
