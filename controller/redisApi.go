package controller

import (
	"errors"
	"encoding/json"
	"net/http"
	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
	"strconv"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gorilla/mux"
	"github.com/fission/fission/pkg/apis/fission.io/v1"
)

func NewClient() redis.Conn {
	c, err := redis.Dial("tcp", "10.103.152.70:6379")
	if err != nil {
		log.Fatalf("Could not connect: %v\n", err)
	}
	return c
}

func (a *API) RecordsApiListAll(w http.ResponseWriter, r *http.Request) {
	client := NewClient()

	iter := 0
	var filtered []string

	for {
		arr, err := redis.Values(client.Do("SCAN", iter))
		if err != nil {
			log.Fatal(err)		// TODO: Is this the right thing to do?
		}
		iter, _ = redis.Int(arr[0], nil)
		k, _ := redis.Strings(arr[1], nil)
		for _, key := range k {
			if strings.HasPrefix(key, "REQ") {
				filtered = append(filtered, key)
			}
		}

		if iter == 0 {
			break
		}
	}

	log.Info("Printing records: ", filtered)

	resp, err := json.Marshal(filtered)
	if err != nil {
		a.respondWithError(w, err)
		return
	}
	a.respondWithSuccess(w, resp)
}

/*
func wellFormedTime(time string) (int64, error) {
	r, err := regexp.Compile(`^\d*[hsmd]$`)

	if err != nil {
		return -1, errors.New("There is a problem with your regexp.\n")
	}

	rd, err := regexp.Compile(`[^\d]`)
	if err != nil {
		return -1, errors.New("There is a problem with your regexp.\n")
	}

	num := rd.Split(time,-1)[0]

	// Checks input matches pattern
	// Checks that lengths add up (any number of digits + single character) == len(input)
	// || len(num) + 1 != len(time)
	if !r.MatchString(time) {
		return -1, errors.New("Improperly formed time input")
	}

	digit, err := strconv.Atoi(num)
	if err != nil {
		return -1, err
	}

	log.Info("Matched pattern, got digit: ", digit, " and unit: ", time[len(time)-1])
	return int64(digit), nil
}

func unit(c string) time.Duration {
	if len(c) != 1 {
		log.Fatal("Can't interpret unit")
	}
	switch c {
	case "s":
		return time.Second
	case "m":
		return time.Minute
	case "h":
		return time.Hour
	case "d":
		return 24 * time.Hour
	default:
		return time.Hour		//TODO: Think of this case
	}
}
*/

func validateSplit(timeInput string) (int64, time.Duration, error) {
	num := timeInput[0:len(timeInput)-1]
	unit := string(timeInput[len(timeInput)-1:])

	num2, err := strconv.Atoi(num)
	if err != nil {
		return -1, time.Hour, err		// Return nil time struct?
	}

	num3 := int64(num2)

	log.Info("Parsed time thusly: ", num3, unit, len(unit))

	switch unit {
	case "s":
		return num3, time.Second, nil
	case "m":
		return num3, time.Minute, nil
	case "h":
		return num3, time.Hour, nil
	case "d":
		return num3, 24 * time.Hour, nil
	default:
		log.Info("Failed to default.")
		return -1, time.Hour, errors.New("Invalid time unit")		//TODO: Think of this case
	}
}

// Input: `from` (hours ago, between 0 [today] and 5) and `to` (same units)
// TODO: End range (validate as well)
// Note: Fractional values don't seem to work -- document that for the user
func (a *API) RecordsApiFilterByTime(w http.ResponseWriter, r *http.Request) {
	fromInput := r.FormValue("from")
	toInput := r.FormValue("to")

	// TODO: Reduce duplicate code
	fromMultiplier, fromUnit, err := validateSplit(fromInput)
	if err != nil {
		a.respondWithError(w, err)
		return
	}

	toMultiplier, toUnit, err := validateSplit(toInput)
	if err != nil {
		a.respondWithError(w, err)
		return
	}

	//fromMultiplier := 24
	//fromUnit := time.Hour
	//
	//toMultiplier := 0
	//toUnit := time.Hour

	now := time.Now()
	year, month, day := now.Date()
	d := time.Date(year, month, day, now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), time.UTC)	// TODO: Location
	then := now.Add(time.Duration(-fromMultiplier) * fromUnit)

	log.Info("Range fromInput: %v, toInput: %v", then, d)
	rangeStart := then.UnixNano()
	log.Info("Start range: ", rangeStart)

	// End Range
	until := now.Add(time.Duration(-toMultiplier) * toUnit)
	rangeEnd := until.UnixNano()

	client := NewClient()

	iter := 0
	var filtered []string

	for {
		arr, err := redis.Values(client.Do("SCAN", iter))
		if err != nil {
			log.Fatal(err)		// TODO: Is this the right thing toInput do?
		}
		iter, _ = redis.Int(arr[0], nil)
		k, _ := redis.Strings(arr[1], nil)
		for _, key := range k {
			if strings.HasPrefix(key, "REQ") {
				val, err := redis.Strings(client.Do("HMGET", key, "Timestamp"))
				if err != nil {
					log.Fatal(err)
					// return err
				}
				tsO, _ := strconv.Atoi(val[0])				// TODO: Get int64 precision fromInput here
				ts := int64(tsO)
				if ts > rangeStart && ts < rangeEnd {
					filtered = append(filtered, key)
				}
			}
		}

		if iter == 0 {
			break
		}
	}

	resp, err := json.Marshal(filtered)
	if err != nil {
		a.respondWithError(w, err)
		return
	}
	a.respondWithSuccess(w, resp)
}

func (a *API) RecordsApiFilterByTrigger(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	trigger := vars["trigger"]

	//trigger := a.extractQueryParamFromRequest(r, "trigger")
	log.Info("In redisApi, got trigger: ", trigger)

	ns := a.extractQueryParamFromRequest(r, "namespace")
	if len(ns) == 0 {
		ns = metav1.NamespaceAll
	}

	// Get all recorders and filter out the ones that aren't attached to the queried trigger
	recorders, err := a.fissionClient.Recorders(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		a.respondWithError(w, err)
		return
	}

	var matchingRecorders []string
	for _, recorder := range recorders.Items {
		if len(recorder.Spec.Triggers) > 0 {
			if includesTrigger(recorder.Spec.Triggers, trigger) {
				matchingRecorders = append(matchingRecorders, recorder.Spec.Name)
			}
		}
	}
	log.Info("Matching recorders: ", matchingRecorders)

	// For each matching recorder, for all its corresponding reqUIDs, if the value's Trigger field == queried trigger,
	// add that reqUID to filtered list

	client := NewClient()

	var filteredReqUIDs []string

	for _, key := range matchingRecorders {
		val, err := redis.Strings(client.Do("LRANGE", key, "0", "-1"))   // TODO: Prefix that distinguishes recorder lists
		if err != nil {
			a.respondWithError(w, err)
		}
		for _, reqUID := range val {
			val, err := redis.Strings(client.Do("HMGET", reqUID, "Trigger"))  // 1-to-1 reqUID - trigger?
			if err != nil {
				log.Fatal(err)
			}
			if val[0] == trigger {
				filteredReqUIDs = append(filteredReqUIDs, reqUID)
			}
		}
	}

	// Placeholder
	resp, err := json.Marshal(filteredReqUIDs)
	if err != nil {
		a.respondWithError(w, err)
		return
	}
	a.respondWithSuccess(w, resp)
}

func includesTrigger(triggers []v1.TriggerReference, query string) bool {
	for _, trigger := range triggers {
		if trigger.Name == query {
			return true
		}
	}
	return false
}

func (a *API) RecordsApiAll(w http.ResponseWriter, r *http.Request) {

}

func (a *API) RecordsApiFilterByFunction(w http.ResponseWriter, r *http.Request) {
	//query := r.FormValue("query")
	vars := mux.Vars(r)
	query := vars["function"]
	ns := a.extractQueryParamFromRequest(r, "namespace")
	if len(ns) == 0 {
		ns = metav1.NamespaceAll
	}

	recorders, err := a.fissionClient.Recorders(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		a.respondWithError(w, err)
		return
	}

	var matchingRecorders []string
	for _, recorder := range recorders.Items {
		if recorder.Spec.Function == query {
			matchingRecorders = append(matchingRecorders, recorder.Spec.Name)
		}
	}
	log.Info("Matching recorders: ", matchingRecorders)

	client := NewClient()

	var filteredReqUIDs []string

	for _, key := range matchingRecorders {
		val, err := redis.Strings(client.Do("LRANGE", key, "0", "-1"))   // TODO: Prefix that distinguishes recorder lists
		if err != nil {
			a.respondWithError(w, err)
		}
		// filteredReqUIDs = append(filteredReqUIDs, strings.Join(val, ","))
		filteredReqUIDs = append(filteredReqUIDs, val...)
	}

	resp, err := json.Marshal(filteredReqUIDs)
	if err != nil {
		a.respondWithError(w, err)
		return
	}
	a.respondWithSuccess(w, resp)
}
