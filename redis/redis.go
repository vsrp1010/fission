package redis

import (
	"net/http"
	"strconv"

	"github.com/golang/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
	"github.com/fission/fission/redis/build/gen"
)

var generator int 		// Temporary

func NewClient() redis.Conn {
	c, err := redis.Dial("tcp", "10.103.152.70:6379")
	if err != nil {
		log.Fatalf("Could not connect: %v\n", err)
	}
	return c
}

func serializeRequest(request *http.Request) []byte {
	URL := make(map[string]string)
	URL["Host"] = request.URL.Host
	URL["Path"] = request.URL.Path
	req := &redisCache.Request{
		Method: proto.String(request.Method),
		URL: URL,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		log.Fatal("Marshalling error: ", err)
	}
	return data
}

func BeginRecord(function *metav1.ObjectMeta, request *http.Request) {
	// reqUID := strings.ToLower(uuid.NewV4().String())
	reqUID := function.Name + strconv.Itoa(generator)
	generator++

	client := NewClient()

	sreq := serializeRequest(request)

	_, err := client.Do("SET", reqUID, "This is a string")
	if err != nil {
		panic(err)
	}

	val, err := redis.String(client.Do("GET", reqUID))
	if err != nil {
		panic(err)
	}
	log.Info("Obtained this key-value pair: ", reqUID, val)
}