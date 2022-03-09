package operator

import (
	"context"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Indexer interface {
	GetByKey(key string) (item interface{}, exists bool, err error)
}

type eventHandler struct {
	filter  Filter
	queue   workqueue.RateLimitingInterface
	indexer Indexer
	handler Handler
}

func NewEventHandler(f Filter, q workqueue.RateLimitingInterface, i Indexer, h Handler) *eventHandler { // @TODO: interface
	return &eventHandler{
		filter:  f,
		queue:   q,
		indexer: i,
		handler: h,
	}
}

func (e *eventHandler) AddFunc(obj interface{}) {
	if obj == nil {
		log.Error("Add with nil obj, skip")
		return
	}

	o := obj.(runtime.Object)
	if err := e.filter.Validate(o); err != nil {
		return
	}

	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		log.Errorf("Add MetaNamespaceKeyFunc error %v", err)
		return
	}

	log.Debugf("Add %T: %s", obj, key)
	e.queue.Add(&event{
		key: key,
		obj: o.DeepCopyObject(),
	})
}

func (e *eventHandler) UpdateFunc(oldObj, newObj interface{}) {
	if newObj == nil || oldObj == nil {
		log.Errorf("Update with Nil Object old: %T new: %T", oldObj, newObj)
		return
	}

	n := newObj.(runtime.Object)
	if err := e.filter.Validate(n); err != nil {
		return
	}

	key, err := cache.MetaNamespaceKeyFunc(newObj)
	if err != nil {
		log.Errorf("Patch MetaNamespaceKeyFunc error %v", err)
		return
	}

	log.Debugf("Update %T: %s", oldObj, key)
	o := oldObj.(runtime.Object)
	e.queue.Add(&updateEvent{
		key:    key,
		newObj: n.DeepCopyObject(),
		oldObj: o.DeepCopyObject(),
	})
}

func (e *eventHandler) DeleteFunc(obj interface{}) {
	if obj == nil {
		log.Error("Delete with nil obj, skip")
		return
	}

	o := obj.(runtime.Object)
	if err := e.filter.Validate(o); err != nil {
		return
	}

	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		log.Errorf("Delete DeletionHandlingMetaNamespaceKeyFunc error %v", err)
		return
	}

	log.Debugf("Delete %T: %s", obj, key)
	e.queue.Add(&event{
		key: key,
		obj: o.DeepCopyObject(),
	})
}

func (c *eventHandler) ProcessNextItem() bool {
	ev, quit := c.queue.Get()
	if quit {
		return false
	}

	defer c.queue.Done(ev)

	e := ev.(Event)
	err := c.handleEvent(e)
	c.handleErr(err, e.GetKey())

	return true
}

func (c *eventHandler) handleEvent(e Event) error {
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

// handleErr checks if an error happened and makes sure we will retry later.
func (c *eventHandler) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.queue.NumRequeues(key) < 5 {
		log.Infof("Error syncing pod %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	utilruntime.HandleError(err)
	log.Infof("Dropping pod %q out of the queue: %v", key, err)
}
