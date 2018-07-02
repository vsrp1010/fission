package main

import (
	"github.com/urfave/cli"
	"fmt"
	"text/tabwriter"
	"os"
)

func recordsView(c *cli.Context) error {
	function := c.String("function")
	trigger := c.String("trigger")
	from := c.String("from")
	to := c.String("to")

	// TODO: Support or refuse multiple filters
	if len(function) != 0 {
		return recordsByFunction(function, c)
	}
	if len(trigger) != 0 {
		return recordsByTrigger(trigger, c)
	}
	if len(from) != 0 {
		return recordsByTime(from, to, c)
	}
	err := recordsAll(c)
	checkErr(err, "view records")			// TODO: View all records by default or last 10?
	return nil
}

func recordsAll(c *cli.Context) error {
	fc := getClient(c.GlobalString("server"))

	records, err := fc.RecordsAll()
	checkErr(err, "view records")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\n",
		"REQUID", "REQUEST METHOD", "FUNCTION", "RESPONSE STATUS", "TRIGGER")
	for _, record := range records {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\n",
			record.ReqUID, record.Req.Method, record.Req.Header["X-Fission-Function-Name"], record.Resp.Status, record.Trigger)
	}
	w.Flush()

	return nil
}

func recordsByTrigger(trigger string, c *cli.Context) error {
	fc := getClient(c.GlobalString("server"))

	records, err := fc.RecordsByTrigger(trigger)
	checkErr(err, "view records")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\n",
		"REQUID", "REQUEST METHOD", "FUNCTION", "RESPONSE STATUS", "TRIGGER")
	for _, record := range records {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\n",
			record.ReqUID, record.Req.Method, record.Req.Header["X-Fission-Function-Name"], record.Resp.Status, record.Trigger)
	}
	w.Flush()

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

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	fmt.Fprintf(w, "%v\n",
		"REQUID")
	for _, record := range records {
		fmt.Fprintf(w, "%v\n", record)
	}
	w.Flush()

	return nil
}

func recordsByTime(from string, to string, c *cli.Context) error {
	fc := getClient(c.GlobalString("server"))

	records, err := fc.RecordsByTime(from, to)
	checkErr(err, "view records")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	fmt.Fprintf(w, "%v\n",
		"REQUID")
	for _, record := range records {
		fmt.Fprintf(w, "%v\n", record)
	}
	w.Flush()

	return nil
}