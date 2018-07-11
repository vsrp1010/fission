package replayer

import (
	"github.com/fission/fission/redis/build/gen"
	"fmt"
	"errors"
	"net/http"
	"io/ioutil"
	log "github.com/sirupsen/logrus"
)

// Make return value a proper Response object
func ReplayRequest(request *redisCache.Request) ([]string, error) {

	// Localhost?
	// URL should already be mapped
	// Router port is 31314

	path := request.URL["Path"]	// Slash included

	log.Info("URL to use: ", path)

	targetUrl := fmt.Sprintf("http://192.168.64.8:31314%v", path)
	targetUrl += "?replayed=true"	// Does nothing now

	resp, err := http.Get(targetUrl)
	if err != nil {
		return []string{}, errors.New(fmt.Sprintf("failed to make get request: %v", err))
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []string{}, errors.New(fmt.Sprintf("failed to read response: %v", err))
	}

	bodyStr := string(body)

	return []string{bodyStr}, nil
}