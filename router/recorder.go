package router

import (
	"k8s.io/apimachinery/pkg/fields"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sCache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/rest"

	"github.com/fission/fission/crd"
	"time"
	"fmt"
)

type RecorderSet struct {
	functionRecorderMap  map[string]bool		// TODO: Keep this as is or define in separate file?
	triggerRecorderMap   map[string]bool

	crdClient            *rest.RESTClient

	recorders            []crd.Recorder
	recStore             k8sCache.Store
	recController        k8sCache.Controller
}

// TODO: How many stores should we use?
// TODO: Originally passed in frmap from router.Start function.
func MakeRecorderSet(crdClient *rest.RESTClient) (*RecorderSet, k8sCache.Store) {
	var rStore k8sCache.Store
	var rController k8sCache.Controller
	recorderSet := &RecorderSet{
		functionRecorderMap: make(map[string]bool),
		triggerRecorderMap: make(map[string]bool),
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
				// Update recorder maps
				go rs.newRecorder(recorder)
				// Sync?
			},
			DeleteFunc: func(obj interface{}) {				//  When does this get invoked?
				recorder := obj.(*crd.Recorder)
				// Update recorder maps
				go rs.disableRecorder(recorder)
				// Sync?
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				oldRecorder := oldObj.(*crd.Recorder)
				newRecorder := newObj.(*crd.Recorder)
				go rs.updateRecorder(oldRecorder, newRecorder)
				// Sync?
			},
		},
	)
	return store, controller
}

// All new recorders are by default enabled
func (rs *RecorderSet) newRecorder(r *crd.Recorder) {
	functions := r.Spec.Functions
	triggers := r.Spec.Triggers

	if len(functions) != 0 {
		for _, function := range functions {
			rs.functionRecorderMap[function.Name] = true
		}
	}

	if len(triggers) != 0 {
		for _, trigger := range triggers {
			rs.triggerRecorderMap[trigger.Name] = true
		}
	}
}

// Delete or disable?
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
}

func (rs *RecorderSet) updateRecorder(old *crd.Recorder, new *crd.Recorder) {
	stateChange := fmt.Sprintf("%v,%v", old.Spec.Enabled, new.Spec.Enabled)
	switch stateChange {
	case "true,false":
		rs.disableRecorder(new)
	case "false,true":
		rs.newRecorder(new)
	case "true,true":
		rs.newRecorder(new)
	case "false,false":
		rs.disableRecorder(new)
	}

}