package redis

import (
	"github.com/fission/fission/crd"
	"github.com/gomodule/redigo/redis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	log "github.com/sirupsen/logrus"
)

type RecordsClient struct {
	crdClient   *crd.FissionClient
	redisClient redis.Conn
}

func MakeRecordsClient(crdClient *crd.FissionClient) *RecordsClient {
	redisClient := NewClient()
	return &RecordsClient{
		crdClient,
		redisClient,
	}
}

func (rc *RecordsClient) FilterByTime() error {
	return nil
}

// For a given trigger, find the associated recorder and all the requests recorded for that recorder
// TODO: Does this work if multiple triggers are attached to a single recorder?
func (rc *RecordsClient) FilterByTrigger(triggerName string) error {
	return nil
}

func (rc *RecordsClient) FilterByFunction(query string) error {
	recorders, err := rc.crdClient.Recorders(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	var matchingRecorders []string
	for _, recorder := range recorders.Items {
		if recorder.Spec.Function.Name == query {
			matchingRecorders = append(matchingRecorders, recorder.Spec.Name)
		}
	}
	log.Info("Matching recorders: ", matchingRecorders)
	return nil
}

func (rc *RecordsClient) FilterByRecorder(recName string) error {
	return nil
}