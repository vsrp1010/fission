package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gomodule/redigo/redis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
	"strconv"
	"github.com/gorilla/mux"
)

func NewClient() redis.Conn {
	c, err := redis.Dial("tcp", "10.103.152.70:6379")
	if err != nil {
		log.Fatalf("Could not connect: %v\n", err)
	}
	return c
}

func (a *API) RecordsApiListAll(w http.ResponseWriter, r *http.Request) {

}

// Input: `from` (hours ago, between 0 [today] and 5) and `to` (same units)
// TODO: End range (validate as well)
func (a *API) RecordsApiFilterByTime(w http.ResponseWriter, r *http.Request) {
	//from := r.FormValue("from")
	vars := mux.Vars(r)
	from := vars["from"]
	from2, _ := strconv.Atoi(from)

	//to := r.FormValue("to")
	now := time.Now()
	year, month, day := now.Date()
	d := time.Date(year, month, day, now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), time.UTC)	// TODO: Location
	then := now.Add(time.Duration(-from2) * time.Hour)
	log.Info("Range from: %v, to: %v", then, d)
	range_start := then.UnixNano()
	log.Info("Start range: ", range_start)

	//

	client := NewClient()

	iter := 0
	var keys []string
	for {
		arr, err := redis.Values(client.Do("SCAN", iter))
		if err != nil {
			log.Fatal(err)		// TODO: Is this the right thing to do?
		}
		iter, _ = redis.Int(arr[0], nil)
		k, _ := redis.Strings(arr[1], nil)
		for _, part := range k {
			if strings.HasPrefix(part, "REQ") {
				keys = append(keys, part)
			}
		}

		if iter == 0 {
			break
		}
	}

	var filtered []string

	for _, key := range keys {
		val, err := redis.Strings(client.Do("HMGET", key, "Timestamp"))
		if err != nil {
			log.Fatal(err)
			// return err
		}
		ts, _ := strconv.Atoi(val[0]) 				// TODO: Get int64 precision from here
		if int64(ts) > range_start {				// TODO: Eh ... can't assume this is a valid check for all scenarios
			filtered = append(filtered, key)
		}
	}

	// Placeholder
	resp, err := json.Marshal(filtered)
	if err != nil {
		a.respondWithError(w, err)
		return
	}
	a.respondWithSuccess(w, resp)
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
		if recorder.Spec.Function.Name == query {
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
		filteredReqUIDs = append(filteredReqUIDs, strings.Join(val, ","))
	}

	resp, err := json.Marshal(filteredReqUIDs)
	if err != nil {
		a.respondWithError(w, err)
		return
	}
	a.respondWithSuccess(w, resp)
}
