package router

import (
	"log"
	"k8s.io/apimachinery/pkg/fields"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sCache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/rest"

	"github.com/fission/fission"
	"github.com/fission/fission/crd"
	"time"

)

type RecorderSet struct {
	parent               *HTTPTriggerSet

	functionRecorderMap  map[string]*crd.Recorder
	triggerRecorderMap   map[string]*crd.Recorder

	crdClient            *rest.RESTClient

	recorders            []crd.Recorder
	recStore             k8sCache.Store
	recController        k8sCache.Controller
}

func MakeRecorderSet(parent *HTTPTriggerSet, crdClient *rest.RESTClient) (*RecorderSet, k8sCache.Store) {
	var rStore k8sCache.Store
	var rController k8sCache.Controller
	recorderSet := &RecorderSet{
		parent: parent,
		functionRecorderMap: make(map[string]*crd.Recorder),
		triggerRecorderMap: make(map[string]*crd.Recorder),
		crdClient: crdClient,
		recorders: []crd.Recorder{},
		recStore: rStore,
		recController: rController,
	}
	return recorderSet, rStore
}

func (rs *RecorderSet) initRecorderController() (k8sCache.Store, k8sCache.Controller) {
	resyncPeriod := 30 * time.Second
	listWatch := k8sCache.NewListWatchFromClient(rs.crdClient, "recorders", metav1.NamespaceAll, fields.Everything())
	store, controller := k8sCache.NewInformer(listWatch, &crd.Recorder{}, resyncPeriod,
		k8sCache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				recorder := obj.(*crd.Recorder)
				go rs.newRecorder(recorder)
			},
			DeleteFunc: func(obj interface{}) {
				recorder := obj.(*crd.Recorder)
				go rs.disableRecorder(recorder)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				oldRecorder := oldObj.(*crd.Recorder)
				newRecorder := newObj.(*crd.Recorder)
				go rs.updateRecorder(oldRecorder, newRecorder)
			},
		},
	)
	return store, controller
}

// All new recorders are by default enabled
func (rs *RecorderSet) newRecorder(r *crd.Recorder) {
	functions := r.Spec.Functions
	triggers := r.Spec.Triggers

	// If triggers are not explicitly specified during the creation of this recorder,
	// keep track of those associated with the function(s) specified [implicitly added triggers]
	needTrack := len(triggers) == 0
	var trackFunction map[fission.FunctionReference]bool

	if len(functions) != 0 {
		for _, function := range functions {
			rs.functionRecorderMap[function.Name] = r
			// If we
			if needTrack {
				trackFunction[function] = true
			}
		}
	}

	// Account for implicitly added triggers
	for _, trigger := range rs.parent.triggers {
		if trackFunction[trigger.Spec.FunctionReference] {
			rs.triggerRecorderMap[trigger.Metadata.Name] = r
		}
	}

	if len(triggers) != 0 {
		for _, trigger := range triggers {
			rs.triggerRecorderMap[trigger.Name] = r
		}
	}

	log.Printf("New recorder! ", r.Metadata)
}

// TODO: Delete or disable?
func (rs *RecorderSet) disableRecorder(r *crd.Recorder) {
	functions := r.Spec.Functions
	triggers := r.Spec.Triggers

	if len(functions) != 0 {
		for _, function := range functions {
			delete(rs.functionRecorderMap, function.Name)		// Alternatively set the value to false
		}
	}

	if len(triggers) != 0 {
		for _, trigger := range triggers {
			delete(rs.triggerRecorderMap, trigger.Name)
		}
	}
	// Reset doRecord
	rs.parent.forceNewRouter()
}

func (rs *RecorderSet) updateRecorder(old *crd.Recorder, new *crd.Recorder) {
	if new.Spec.Enabled == true {
		rs.newRecorder(new)
	} else {
		rs.disableRecorder(new)
	}
}

func (rs *RecorderSet) triggerDeleted(trigger *crd.HTTPTrigger) {
	delete(rs.triggerRecorderMap, trigger.Metadata.Name)
}

func (rs *RecorderSet) funcDeleted(function *crd.Function) {
	delete(rs.functionRecorderMap, function.Metadata.Name)
}