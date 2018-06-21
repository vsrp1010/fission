package redis

import (
	"net/http"
	"github.com/golang/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
	"github.com/fission/fission/redis/build/gen"
	"strings"
	"github.com/satori/go.uuid"
	"fmt"
)

func NewClient() redis.Conn {
	c, err := redis.Dial("tcp", "10.103.152.70:6379")
	if err != nil {
		log.Fatalf("Could not connect: %v\n", err)
	}
	return c
}

func serializeRequest(request *http.Request) []byte {
	// TODO: Capture more url fields if needed
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

	data, err := proto.Marshal(req)
	if err != nil {
		log.Fatal("Marshalling error: ", err)
	}
	return data
}

func BeginRecord(function *metav1.ObjectMeta, request *http.Request) {
	UID := strings.ToLower(uuid.NewV4().String())
	reqUID := function.Name + UID

	client := NewClient()

	sReq := serializeRequest(request)

	_, err := client.Do("SET", reqUID, sReq)
	if err != nil {
		panic(err)
	}

	val, err := redis.Bytes(client.Do("GET", reqUID))
	if err != nil {
		panic(err)
	}

	req := &redisCache.Request{}
	err = proto.Unmarshal(val, req)
	if err != nil {
		log.Fatal("Unmarshalling error: ", err)
	}

	log.Info(fmt.Sprintf("Obtained this key-value pair: %v : %v", reqUID, req))
}