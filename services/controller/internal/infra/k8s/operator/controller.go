package operator

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const conciliationFrequency = time.Second * 5

var ErrNoAppLabelFound = errors.New("no app label found ")

type handler interface {
	Created(ctx context.Context, obj runtime.Object)
	Updated(ctx context.Context, old, new runtime.Object)
	Deleted(ctx context.Context, obj runtime.Object)
}

// ListWatcher defines list and watch methods
type ListWatcher interface {
	List(options metav1.ListOptions) (runtime.Object, error)
	Watch(options metav1.ListOptions) (watch.Interface, error)
}

// Filter segregates object selection
type Filter interface {
	Object() runtime.Object
	Validate(runtime.Object) error
}

type controller struct {
	indexer  cache.Indexer
	informer cache.Controller
	queue    workqueue.RateLimitingInterface

	handler handler
}

// @TODO: Clean out! It must be testeable!!!
func Build(handler handler, filter Filter, watcher ListWatcher) *controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	indexer, informer := cache.NewIndexerInformer(
		&cache.ListWatch{
			ListFunc:  watcher.List,
			WatchFunc: watcher.Watch,
		},
		filter.Object(),
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if obj == nil {
					log.Error("Add with nil obj, skip")
					return
				}

				o := obj.(runtime.Object)
				if err := filter.Validate(o); err != nil {
					return
				}

				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err != nil {
					log.Errorf("Add MetaNamespaceKeyFunc error %v", err)
					return
				}

				log.Debugf("Add %T: %s", obj, key)
				queue.Add(&event{
					key: key,
					obj: o.DeepCopyObject(),
				})
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				if newObj == nil || oldObj == nil {
					log.Errorf("Update with Nil Object old: %T new: %T", oldObj, newObj)
					return
				}

				n := newObj.(runtime.Object)
				if err := filter.Validate(n); err != nil {
					return
				}

				key, err := cache.MetaNamespaceKeyFunc(newObj)
				if err != nil {
					log.Errorf("Patch MetaNamespaceKeyFunc error %v", err)
					return
				}

				log.Debugf("Update %T: %s", oldObj, key)
				o := oldObj.(runtime.Object)
				queue.Add(&updateEvent{
					key:    key,
					newObj: n.DeepCopyObject(),
					oldObj: o.DeepCopyObject(),
				})
			},
			DeleteFunc: func(obj interface{}) {
				if obj == nil {
					log.Error("Delete with nil obj, skip")
					return
				}

				o := obj.(runtime.Object)
				if err := filter.Validate(o); err != nil {
					return
				}

				key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
				if err != nil {
					log.Errorf("Delete DeletionHandlingMetaNamespaceKeyFunc error %v", err)
					return
				}

				log.Debugf("Delete %T: %s", obj, key)
				queue.Add(&event{
					key: key,
					obj: o.DeepCopyObject(),
				})
			},
		},
		cache.Indexers{},
	)
	return &controller{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
		handler:  handler,
	}
}

// Run begins watching and syncing.
func (c *controller) Run(workers int, stopCh chan struct{}) {
	defer utilruntime.HandleCrash()

	defer c.queue.ShutDown()
	log.Infof("Starting controller runner with %d workers", workers)

	go c.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		utilruntime.HandleError(errors.New("timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, conciliationFrequency, stopCh)
	}

	<-stopCh
	log.Info("Stopping controller")
}

// runWorker executes the loop to process new items added to the queue
func (c *controller) runWorker() {
	for c.processNextItem() {
	}
}

func (c *controller) processNextItem() bool {
	ev, quit := c.queue.Get()
	if quit {
		return false
	}

	defer c.queue.Done(ev)

	e := ev.(Event)
	err := c.handleEvent(e)
	c.handleError(err, e.GetKey())

	return true
}

func (c *controller) handleEvent(e Event) error {
	ctx := context.Background()
	_, exists, err := c.indexer.GetByKey(e.GetKey())
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", e.GetKey(), err)
		return err
	}

	if !exists {
		log.Infof("Object %s does not exist anymore, delete! event: %T", e.GetKey(), e)
		if ev, ok := e.(*event); ok {
			c.handler.Deleted(ctx, ev.obj)
		}
		if ev, ok := e.(*updateEvent); ok {
			c.handler.Deleted(ctx, ev.oldObj)
		}
		return nil
	}

	switch ev := e.(type) {
	case *event:
		c.handler.Created(ctx, ev.obj)
	case *updateEvent:
		c.handler.Updated(ctx, ev.newObj, ev.oldObj)
	}

	return nil
}

func (c *controller) handleError(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}
	if c.queue.NumRequeues(key) < 5 {
		log.Errorf("Error syncing pod %v: %v", key, err)
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	utilruntime.HandleError(err)
	log.Errorf("Dropping pod %q out of the queue: %v", key, err)
}
