package redis

import (
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
	"github.com/fission/fission/redis/build/gen"
	"strings"
	"strconv"
	"time"
	"fmt"
)

func NewClient() redis.Conn {
	c, err := redis.Dial("tcp", "10.103.152.70:6379")
	if err != nil {
		log.Fatalf("Could not connect: %v\n", err)
	}
	return c
}

func EndRecord(triggerName string, recorderName string, reqUID string, request *http.Request, response *http.Response, namespace string, timestamp int64) {
	// Case where the function should not have been recorded
	if len(reqUID) == 0 {
		return
	}

	client := NewClient()

	url := make(map[string]string)
	url["Host"] = request.URL.Host
	url["Path"] = request.URL.Path

	header := make(map[string]string)
	for key, value := range request.Header {
		header[key] = strings.Join(value, ",")
	}

	form := make(map[string]string)
	for key, value := range request.Form {
		form[key] = strings.Join(value, ",")
	}

	postForm := make(map[string]string)
	for key, value := range request.PostForm {
		postForm[key] = strings.Join(value, ",")
	}

	req := &redisCache.Request{
		Method:   "GET",
		URL:      url,
		Header:   header,
		Host:     request.Host,
		Form:     form,
		PostForm: postForm,
	}

	resp := &redisCache.Response{
		Status: response.Status,
		StatusCode: int32(response.StatusCode),
	}

	ureq := &redisCache.UniqueRequest {
		Req: req,
		Resp: resp,
		Trigger: triggerName,
	}

	data, err := proto.Marshal(ureq)
	if err != nil {
		log.Fatal("Marshalling UniqueRequest error: ", err)
	}

	_, err = client.Do("HMSET", reqUID, "ReqResponse", data, "Timestamp", timestamp, "Trigger", triggerName)
	if err != nil {
		panic(err)
	}

	//_, err = client.Do("LPUSH", recorderName, reqUID)
	//if err != nil {
	//	panic(err)
	//}

	// FilterByTime(100.00)
}

// Currently only prints the records from the past n seconds
// TODO: Units of start and end time specified by user?
func FilterByTime(pastN float64) error {
	// Needs to scan all records, converting each record's nanosecond int64 timestamp to time.Time type
	// and subtract that time from time.Now() to see if it's within pastN seconds. If so, print that record.

	client := NewClient()

	iter := 0
	var keys []string
	for {
		arr, err := redis.Values(client.Do("SCAN", iter))
		if err != nil {
			//log.Fatal(err)
			return err
		}
		iter, _ = redis.Int(arr[0], nil)
		k, _ := redis.Strings(arr[1], nil)
		keys = append(keys, k...)

		if iter == 0 {
			break
		}
	}

	now := time.Unix(0, int64(time.Now().UnixNano()))
	var filtered []string

	for _, key := range keys {
		val, err := redis.Strings(client.Do("HMGET", key, "Timestamp"))
		if err != nil {
			// log.Fatal(err)
			return err
		}
		ts, _ := strconv.Atoi(val[0]) 				// TODO: Get int64 precision from here
		uts := time.Unix(0, int64(ts))
		difference := now.Sub(uts).Seconds()
		fmt.Println(fmt.Sprintf("Recorded %v %v seconds ago: ", key, difference))
		if difference < pastN {
			filtered = append(filtered, key)
		}
	}

	log.Info("Filtered reqUIDs: ", filtered)
	return nil
}
