/*
Copyright 2018 The Fission Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package controller

import (
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fission/fission/redis"
	"fmt"
	"io/ioutil"
)

func (a *API) DebugGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	query := vars["reqUID"]

	functions, err := a.fissionClient.Functions(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		a.respondWithError(w, err)
		return
	}

	environments, err := a.fissionClient.Environments(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		a.respondWithError(w, err)
		return
	}

	debugImg, correspFn, err := redis.DebugByReqUID(query, functions, environments)
	if err != nil {
		a.respondWithError(w, err)
		return
	}

	// After identifying the debug image, issue a request to the router at the function URL with a special header set
	// containing the image.
	// Router should invoke the executor, passing along the debug image, which should create a new deploy and return the
	// service IP.

	client := http.DefaultClient

	routerUrl := fmt.Sprintf("http://router.%v/fission-function/%v", podNamespace, correspFn)

	req, err := http.NewRequest("GET", routerUrl, nil)
	if err != nil {
		a.respondWithError(w, errors.New(fmt.Sprintf("failed to create new HTTP request: %v", err)))
	}
	req.Header.Add("X-Fission-Replay-Debug", debugImg)

	resp, err := client.Do(req)
	if err != nil {
		a.respondWithError(w, errors.New(fmt.Sprintf("failed to fetch new debug pod service IP: %v", err)))
	}

	log.Print("In controller, after debug request made: ", resp.Status)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		a.respondWithError(w, errors.New(fmt.Sprintf("failed to read response: %v", err)))
	}

	log.Printf("THIS IS SUPPOSED TO BE THE SERVICE > %v", string(body))

	a.respondWithSuccess(w, body)
}
