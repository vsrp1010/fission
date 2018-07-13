package replayer

import (
	"github.com/fission/fission/redis/build/gen"
	"fmt"
	"errors"
	"net/http"
	"io/ioutil"
	log "github.com/sirupsen/logrus"
	"bytes"
)

// Make return value a proper Response object
func ReplayRequest(request *redisCache.Request) ([]string, error) {

	// Localhost?
	// URL should already be mapped
	// Router port is 31314

	path := request.URL["Path"]	// Slash included
	payload := request.URL["Payload"]

	targetUrl := fmt.Sprintf("http://192.168.64.8:31314%v", path)

	log.Info("Payload > ", payload)

	targetUrl += "?replayed=true"	// TODO: Append url queries if provided, otherwise will be invalid (?)

	var resp *http.Response
	var err error
	if request.Method == http.MethodGet {
		resp, err = http.Get(targetUrl)
	} else if request.Method == http.MethodPost {
		resp, err = http.Post(targetUrl, "application/json", bytes.NewReader([]byte(payload)))
	}

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