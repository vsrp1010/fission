package replayer

import (
	"github.com/fission/fission/redis/build/gen"
	"fmt"
	"errors"
	"net/http"
	"io/ioutil"
	log "github.com/sirupsen/logrus"
	"bytes"
	"strings"
)

// Make return value a proper Response object
func ReplayRequest(request *redisCache.Request) ([]string, error) {

	// Localhost?
	// URL should already be mapped
	// Router port is 31314

	path := request.URL["Path"]	// Slash included
	payload := request.URL["Payload"]
	escPayload := strings.Replace(payload, "\"", "\\\"", -1)
	//payloadExists := request.URL["PayloadExists"]

	targetUrl := fmt.Sprintf("http://192.168.64.8:31314%v", path)

	expected := "{\"title\":\"Ms\",\"name\":\"Williams\",\"item\":\"racket\"}"

	log.Info("PAYLOAD > ", payload, " and Escaped Payload > ", escPayload, " length: ", len(escPayload))
	log.Info("Expected PAYLOAD > ", expected, len(expected))

	targetUrl += "?replayed=true"	// TODO: Append to url queries if provided, otherwise will be invalid

	//resp, err := http.Post(targetUrl, "application/json", bytes.NewReader([]byte("{\"title\":\"Ms\",\"name\":\"Williams\",\"item\":\"racket\"}")))
	resp, err := http.Post(targetUrl, "application/json", bytes.NewReader([]byte(payload)))

	if err != nil {
		return []string{}, errors.New(fmt.Sprintf("failed to make request: %v", err))
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []string{}, errors.New(fmt.Sprintf("failed to read response: %v", err))
	}

	bodyStr := string(body)

	log.Info("GOT THIS BODY! ", bodyStr)

	return []string{bodyStr}, nil
}