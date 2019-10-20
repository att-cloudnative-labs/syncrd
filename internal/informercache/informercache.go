package informercache

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

type InformerCache struct {
	informerFactory dynamicinformer.DynamicSharedInformerFactory
	informers       map[schema.GroupVersionResource]informers.GenericInformer
	handlers        cache.ResourceEventHandlerFuncs
	stopCh          chan struct{}
}

func NewInformerCache(client dynamic.Interface, handlers cache.ResourceEventHandlerFuncs) *InformerCache {
	return &InformerCache{
		dynamicinformer.NewDynamicSharedInformerFactory(client, 30*time.Second),
		make(map[schema.GroupVersionResource]informers.GenericInformer),
		handlers,
		make(chan struct{}),
	}
}

func (ic *InformerCache) GetInformer(gvr schema.GroupVersionResource) informers.GenericInformer {
	if ic.informers[gvr] == nil {
		informer := ic.informerFactory.ForResource(gvr)
		informer.Informer().AddEventHandler(ic.handlers)
		ic.informers[gvr] = informer
		ic.informerFactory.Start(ic.stopCh)
	}
	return ic.informers[gvr]
}

func NewInformerCacheTest() {
	cfg, err := clientcmd.BuildConfigFromFlags("", "/Users/georgebraxton/.kube/config")
	if err != nil {
		fmt.Printf("error: %s", err.Error())
	}
	client, err := dynamic.NewForConfig(cfg)
	if err != nil {
		fmt.Printf("error: %s", err.Error())
	}
	//resInterface := client.Resource(schema.GroupVersionResource{"networking.istio.io", "v1alpha3", "serviceentries"})

	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(client, 30*time.Second)
	informer := informerFactory.ForResource(schema.GroupVersionResource{Group: "networking.istio.io", Version: "v1alpha3", Resource: "serviceentries"})
	stopCh := make(chan struct{})
	informerFactory.Start(stopCh)
	for !informer.Informer().HasSynced() {
		fmt.Println("informer not synced yet. waiting 1s...")
		time.Sleep(time.Second)
	}
	podList, err := informer.Lister().ByNamespace("default").List(labels.Everything())
	if err != nil {
		fmt.Printf("error: %s", err.Error())
	}

	fmt.Printf("resourceList:\n")
	for _, pod := range podList {
		mo, err := meta.Accessor(pod)
		if err != nil {
			fmt.Printf("error: %s", err.Error())
			return
		}
		fmt.Printf("resource: %s\n", mo.GetName())
	}

	stopCh <- struct{}{}
}

func getGroup() {
	gvr := schema.GroupVersionResource{Group: "com.geob", Version: "v2", Resource: "bart"}
	fmt.Printf("gvr: %s\n", gvr.String())
	fmt.Printf("group: %s\n", gvr.Group)
	fmt.Printf("version: %s\n", gvr.Version)
	fmt.Printf("group: %s\n", gvr.Resource)
}
