package swarm

import (
	ap "github.com/marcosQuesada/k8s-swarm/pkg/config"
	"k8s.io/client-go/util/workqueue"
	"testing"
)

// @TODO: Development test, remove IT!
func TestQueueBehaviourDevelopmentTest(t *testing.T) {
	queue := workqueue.New()
	queue.Add(ap.Job("fooo_0"))
	queue.Add(ap.Job("fooo_1"))
	queue.Add(ap.Job("fooo_2"))
	queue.Add(ap.Job("fooo_3"))
	queue.Add(ap.Job("fooo_4"))

	if expected, got := 5, queue.Len(); expected != got {
		t.Fatalf("unexpected queue size, expected %d got %d", expected, got)
	}

	k0, _ := queue.Get()
	key0, ok := k0.(ap.Job)
	if !ok {
		t.Fatalf("unexpected key type, got %T", k0)
	}

	if expected, got := "fooo_0", string(key0); expected != got {
		t.Fatalf("unexpected collection key, expected %s got %s", expected, got)
	}
	queue.Done(k0)

	k1, _ := queue.Get()
	queue.Done(k1)

	k2, _ := queue.Get()
	queue.Done(k2)

	if expected, got := 2, queue.Len(); expected != got {
		t.Fatalf("unexpected queue size, expected %d got %d", expected, got)
	}

	k3, _ := queue.Get()
	queue.Done(k3)

	k4, _ := queue.Get()
	queue.Done(k4)

	if expected, got := 0, queue.Len(); expected != got {
		t.Fatalf("unexpected queue size, expected %d got %d", expected, got)
	}

	queue.ShutDown()
	_, shut := queue.Get()
	if !shut {
		t.Error("expected queue on shutdown")
	}
}
