package main

import (
	"errors"
	"github.com/urfave/cli"
	"fmt"
)

func recordsView(c *cli.Context) error {
	function := c.String("function")
	from := c.String("from")

	// TODO: Support or refuse multiple filters
	if len(function) != 0 {
		return recordsByFunction(function, c)
	}
	if len(from) != 0 {
		return recordsByTime(from, c)
	}
	checkErr(errors.New("fallback"), "view records")			// TODO: View all records by default or last 10?
	return nil
}

// TODO: More accurate function name (function filter)
func recordsByFunction(function string, c *cli.Context) error {
	fc := getClient(c.GlobalString("server"))

	/*
	rc := redis.MakeRecordsClient(fc)
	err := rc.FilterByFunction("hello")
	*/

	records, err := fc.RecordsByFunction(function)
	checkErr(err, "view records")

	fmt.Println(records)

	return nil
}

func recordsByTime(from string, c *cli.Context) error {
	fc := getClient(c.GlobalString("server"))

	records, err := fc.RecordsByTime(from)
	checkErr(err, "view records")

	fmt.Println(records)

	return nil
}