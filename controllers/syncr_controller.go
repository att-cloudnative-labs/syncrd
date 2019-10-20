/*

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

package controllers

import (
	"context"
	"fmt"
	"github.com/att-cloud-native-labs/syncrd/internal/informercache"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	syncrv1beta1 "github.com/att-cloud-native-labs/syncrd/api/v1beta1"
)

const (
	SynchronizeNamespaceAnnotation = "syncrd.atteg.com/synchronize"
)

// SyncSetting holds settings for Resource to sync
type SyncSetting struct {
	SyncEnabled bool
	MatchLabels map[string]string
}

// SyncrReconciler reconciles a Syncr object
type SyncrReconciler struct {
	client.Client
	Log           logr.Logger
	dynamicClient dynamic.Interface
	informerCache *informercache.InformerCache
	syncables     map[schema.GroupVersionResource]map[types.NamespacedName]SyncSetting
}

// +kubebuilder:rbac:groups=syncrd.atteg.com,resources=syncrs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=syncrd.atteg.com,resources=syncrs/status,verbs=get;update;patch

func (r *SyncrReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("syncr", req.NamespacedName)

	// your logic here
	var syncr syncrv1beta1.Syncr
	if err := r.Get(ctx, req.NamespacedName, &syncr); err != nil {
		log.Error(err, "unable to fetch Syncr")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	gvr := schema.GroupVersionResource{Group: syncr.Spec.APIGroup, Version: syncr.Spec.APIVersion, Resource: syncr.Spec.APIResource}
	r.AddSyncable(gvr, syncr.Spec.Namespace, syncr.Spec.Name, syncr.Spec.MatchLabels)
	_ = r.informerCache.GetInformer(gvr)
	return ctrl.Result{}, nil
}

func (r *SyncrReconciler) AddSyncable(gvr schema.GroupVersionResource, namespace string, name string, matchLabels map[string]string) {
	if r.syncables[gvr] == nil {
		r.syncables[gvr] = make(map[types.NamespacedName]SyncSetting, 0)
	}
	r.syncables[gvr][types.NamespacedName{Namespace: namespace, Name: name}] = SyncSetting{SyncEnabled: true, MatchLabels: matchLabels}
}

func (r *SyncrReconciler) RemoveSyncable(gvr schema.GroupVersionResource, namespace string, name string) {
	if r.syncables[gvr] == nil {
		return
	}
	r.syncables[gvr][types.NamespacedName{Namespace: namespace, Name: name}] = SyncSetting{SyncEnabled: false}
}

func (r *SyncrReconciler) syncResource(gvr schema.GroupVersionResource, resource *unstructured.Unstructured, matchLabels map[string]string) {
	r.Log.Info("starting syncResource")
	ctx := context.Background()
	var matchingLabels client.MatchingLabels = matchLabels

	var namespaceList v1.NamespaceList
	if err := r.List(ctx, &namespaceList, matchingLabels); err != nil {
		r.Log.Error(err, "error getting namespace list")
	}
	client := r.dynamicClient.Resource(gvr)
	r.Log.Info("retrieved namespaces")
	for _, namespace := range namespaceList.Items {
		r.Log.Info("syncing namespace", "namespace", namespace)
		if namespace.Annotations[SynchronizeNamespaceAnnotation] == "true" && namespace.Name != resource.GetName() {
			resourceCopy := resource.DeepCopy()
			resourceCopy.SetNamespace(namespace.Name)
			r.Log.Info("copied resource")
			existing, err := client.Namespace(namespace.Name).Get(resource.GetName(), metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					r.Log.Info("resource didn't exist in target...creating")
					resourceCopy.SetResourceVersion("")
					_, err := client.Namespace(namespace.Name).Create(resourceCopy, metav1.CreateOptions{})
					if err != nil {
						r.Log.Error(err, "error creating resource")
					} else {
						r.Log.Info(fmt.Sprintf("created %s, %s in namespace %s", resourceCopy.GetKind(), resourceCopy.GetName(), resourceCopy.GetNamespace()))
					}
					continue
				} else {
					r.Log.Error(err, "error checking for existing resource")
					return
				}
			}
			r.Log.Info("checking equality of existing")
			if !reflect.DeepEqual(resource.Object["spec"], existing.Object["spec"]) ||
				!reflect.DeepEqual(resource.Object["data"], existing.Object["data"]) {
				r.Log.Info("resources don't match, updating")
				resourceCopy.SetResourceVersion(existing.GetResourceVersion())
				resourceCopy.SetUID(existing.GetUID())
				_, err = client.Namespace(namespace.GetName()).Update(resourceCopy, metav1.UpdateOptions{})
				if err != nil {
					r.Log.Error(err, "error updating resource")
					return
				} else {
					r.Log.Info(fmt.Sprintf("updated %s, %s in namespace %s", resourceCopy.GetKind(), resourceCopy.GetName(), resourceCopy.GetNamespace()))
				}
			} else {
				r.Log.Info("resources matched")
			}
		}
	}
}

func (r *SyncrReconciler) AddFunc(obj interface{}) {
	un := obj.(*unstructured.Unstructured)
	// Not sure how to handle this without using unsafe guess to resource, only
	// other option is to create a mapping sourced from Syncrs where each
	// object must include both resource and kind.
	gvr, _ := meta.UnsafeGuessKindToResource(un.GroupVersionKind())
	nsn := types.NamespacedName{Namespace: un.GetNamespace(), Name: un.GetName()}

	if r.syncables[gvr][nsn].SyncEnabled {
		r.syncResource(gvr, un, r.syncables[gvr][nsn].MatchLabels)
	}
}

func (r *SyncrReconciler) UpdateFunc(oldObj, newObj interface{}) {
	un := newObj.(*unstructured.Unstructured)
	// Not sure how to handle this without using unsafe guess to resource, only
	// other option is to create a mapping sourced from Syncrs where each
	// object must include both resource and kind.
	gvr, _ := meta.UnsafeGuessKindToResource(un.GroupVersionKind())
	nsn := types.NamespacedName{Namespace: un.GetNamespace(), Name: un.GetName()}

	if r.syncables[gvr][nsn].SyncEnabled {
		r.syncResource(gvr, un, r.syncables[gvr][nsn].MatchLabels)
	}
}

func (r *SyncrReconciler) DeleteFunc(obj interface{}) {

}

func (r *SyncrReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.dynamicClient == nil {
		client, err := dynamic.NewForConfig(mgr.GetConfig())
		if err != nil {
			return fmt.Errorf("error creating dynamic client from config: %s", err.Error())
		}
		r.dynamicClient = client
	}
	handlerFuncs := cache.ResourceEventHandlerFuncs{
		AddFunc:    r.AddFunc,
		UpdateFunc: r.UpdateFunc,
		DeleteFunc: r.DeleteFunc,
	}
	if r.informerCache == nil {
		r.informerCache = informercache.NewInformerCache(r.dynamicClient, handlerFuncs)
	}
	if r.syncables == nil {
		r.syncables = make(map[schema.GroupVersionResource]map[types.NamespacedName]SyncSetting)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&syncrv1beta1.Syncr{}).
		Complete(r)
}
