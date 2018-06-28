package router

import (
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
	recorderSet := &RecorderSet{
		parent: parent,
		functionRecorderMap: make(map[string]*crd.Recorder),
		triggerRecorderMap: make(map[string]*crd.Recorder),
		crdClient: crdClient,
		recorders: []crd.Recorder{},
		recStore: rStore,
	}
	_, recorderSet.recController = recorderSet.initRecorderController()
	return recorderSet, rStore
}

func (rs *RecorderSet) initRecorderController() (k8sCache.Store, k8sCache.Controller) {
	resyncPeriod := 100 * time.Second
	//resyncPeriod := 0 * time.Second
	listWatch := k8sCache.NewListWatchFromClient(rs.crdClient, "recorders", metav1.NamespaceAll, fields.Everything())
	store, controller := k8sCache.NewInformer(listWatch, &crd.Recorder{}, resyncPeriod,
		k8sCache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				recorder := obj.(*crd.Recorder)
				rs.newRecorder(recorder)
			},
			DeleteFunc: func(obj interface{}) {
				recorder := obj.(*crd.Recorder)
				rs.disableRecorder(recorder)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				oldRecorder := oldObj.(*crd.Recorder)
				newRecorder := newObj.(*crd.Recorder)
				rs.updateRecorder(oldRecorder, newRecorder)
			},
		},
	)
	return store, controller
}

// All new recorders are by default enabled
func (rs *RecorderSet) newRecorder(r *crd.Recorder) {
	function := r.Spec.Function
	triggers := r.Spec.Triggers

	// If triggers are not explicitly specified during the creation of this recorder,
	// keep track of those associated with the function(s) specified [implicitly added triggers]
	needTrack := len(triggers) == 0
	trackFunction := make(map[fission.FunctionReference]bool)

	//log.Info("New recorder ! Need track? ", needTrack)

	/*
	if len(functions) != 0 {
		for _, function := range functions {
			rs.functionRecorderMap[function.Name] = r
			// If we
			if needTrack {
				trackFunction[function] = true
			}
		}
	}
	*/
	rs.functionRecorderMap[function.Name] = r
	if needTrack {
		trackFunction[function] = true
	}

	// Account for implicitly added triggers
	if needTrack {
		for _, t := range rs.parent.triggerStore.List() {
			trigger := *t.(*crd.HTTPTrigger)
			if trackFunction[trigger.Spec.FunctionReference] {
				rs.triggerRecorderMap[trigger.Metadata.Name] = r
			}
		}
	}

	if len(triggers) != 0 {
		for _, trigger := range triggers {
			rs.triggerRecorderMap[trigger.Name] = r
		}
	}

	//log.Info("See updated trigger map: ", keys(rs.triggerRecorderMap))
	//log.Info("See updated function map: ", keys(rs.functionRecorderMap))
}

func keys(m map[string]*crd.Recorder) (keys []string) {
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// TODO: Delete or disable?
func (rs *RecorderSet) disableRecorder(r *crd.Recorder) {
	function := r.Spec.Function
	triggers := r.Spec.Triggers

	//log.Info("Disabling recorder !")

	//if len(functions) != 0 {
	//	for _, function := range functions {
			delete(rs.functionRecorderMap, function.Name)		// Alternatively set the value to false
			//rs.functionRecorderMap[function.Name] = nil
	//	}
	//}

	if len(triggers) != 0 {
		for _, trigger := range triggers {
			delete(rs.triggerRecorderMap, trigger.Name)
			// rs.triggerRecorderMap[trigger.Name] = nil
		}
	}
	// Reset doRecord
	rs.parent.forceNewRouter()

	//log.Info("See updated trigger map: ", keys(rs.triggerRecorderMap))
	//log.Info("See updated function map: ", keys(rs.functionRecorderMap))
}

func (rs *RecorderSet) updateRecorder(old *crd.Recorder, new *crd.Recorder) {
	if new.Spec.Enabled == true {
		rs.newRecorder(new)
	} else {
		rs.disableRecorder(new)
	}
}

func (rs *RecorderSet) deleteTrigger(trigger *crd.HTTPTrigger) {
	delete(rs.triggerRecorderMap, trigger.Metadata.Name)
}

func (rs *RecorderSet) deleteFunction(function *crd.Function) {
	delete(rs.functionRecorderMap, function.Metadata.Name)
}