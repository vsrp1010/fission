package redis

import (
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
	"github.com/fission/fission/redis/build/gen"
	"strings"
	"net/url"
	"fmt"
	"net/http/httputil"
	"encoding/json"
)

func DebugRequest(request http.Request) {
	dump, err := httputil.DumpRequest(&request, true)
	if err != nil {
		log.Info(err)
	}
	log.Info(fmt.Sprintf("%q", dump))
}

func NewClient() redis.Conn {
	// TODO: Load redis ClusterIP from environment variable / configmap
	c, err := redis.Dial("tcp", "10.102.223.159:6379")
	if err != nil {
		log.Fatalf("Could not connect: %v\n", err)
	}
	return c
}

func EndRecord(triggerName string, recorderName string, reqUID string, request *http.Request, originalUrl url.URL, parsedBody string, response *http.Response, namespace string, timestamp int64) {
	// Case where the function should not have been recorded
	if len(reqUID) == 0 {
		return
	}

	replayed := originalUrl.Query().Get("replayed")

	log.Info("EndRecord: URL > ", originalUrl.String(), " with body: ", parsedBody)
	//log.Info("Debug request follows (if nothing, errored out)")
	//DebugRequest(originalReq)

	if replayed == "true" {
		log.Info("This was a replayed request.")
		return
	}

	/*
	err := request.ParseForm()
	if err != nil {
		log.Info("Problem parsing form: ", err)
	}

	var bodyRequest []byte
	bodyRequest, err = ioutil.ReadAll(request.Body)
	if err != nil {
		log.Info("Problem reading bytes of request body: ", err)
	}
	bodyString := string(bodyRequest)

	log.Info(fmt.Sprintf("1: {%v}, 2: {%v}, 3: {%v}", request.Form.Encode(), request.PostForm.Encode(), bodyString))
	*/

	//payload := originalUrl.RawQuery 		// TODO: Order? If both raw query and form entries given, use both? Test both.
	payload := parsedBody

	postFormEntries := request.PostForm.Encode()
	if len(postFormEntries) > 0 {
		payload += "&" + postFormEntries
	}

	fullPath := originalUrl.String() + postFormEntries

	escPayload := string(json.RawMessage(payload))

	log.Info("Escaped payload parsed: ", escPayload, " and FullPath > ", fullPath)

	client := NewClient()

	url := make(map[string]string)
	url["Host"] = request.URL.Host
	url["Path"] = fullPath // Previously originalUrl.String()	// Previously request.URL.Path
	url["Payload"] = escPayload
	url["PayloadExists"] = payload
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
		Trigger: triggerName,			// TODO: Why is this here when Trigger is set as a separate field?
	}

	data, err := proto.Marshal(ureq)
	if err != nil {
		log.Fatal("Marshalling UniqueRequest error: ", err)
	}

	_, err = client.Do("HMSET", reqUID, "ReqResponse", data, "Timestamp", timestamp, "Trigger", triggerName)
	if err != nil {
		panic(err)
	}

	_, err = client.Do("LPUSH", recorderName, reqUID)
	if err != nil {
		panic(err)
	}
}
