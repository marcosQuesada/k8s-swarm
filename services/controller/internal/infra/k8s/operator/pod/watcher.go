package pod

import (
	"context"
	"github.com/marcosQuesada/k8s-swarm/services/controller/internal/infra/k8s/operator"
	"k8s.io/client-go/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

type listWatcherAdapter struct {
	client    kubernetes.Interface
	namespace string
}

// NewListWatcherAdapter instantiates pod list watcher adapter
func NewListWatcherAdapter(c kubernetes.Interface, namespace string) operator.ListWatcher {
	return &listWatcherAdapter{
		client:    c,
		namespace: namespace,
	}
}

// List handles pod listing in the specified namespace
func (a *listWatcherAdapter) List(options metav1.ListOptions) (runtime.Object, error) {
	return a.client.CoreV1().Pods(a.namespace).List(context.Background(), options)
}

// Watch creates a stream of pod events in watched namespace
func (a *listWatcherAdapter) Watch(options metav1.ListOptions) (watch.Interface, error) {
	return a.client.CoreV1().Pods(a.namespace).Watch(context.Background(), options)
}
