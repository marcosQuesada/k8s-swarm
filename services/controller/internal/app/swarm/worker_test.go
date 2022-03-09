package swarm

import (
	ap "github.com/marcosQuesada/k8s-swarm/pkg/config"
	"k8s.io/client-go/util/workqueue"
	"testing"
)

//
//func TestItRemainsOnWaitingAssignationUntilAssigned(t *testing.T) {
//	call := &fakeCaller{}
//	w1Index := 1
//	w1IP := net.ParseIP("8.8.8.1")
//	w1Name := "worker-1"
//	w := newWorker(w1Index, w1Name, w1IP, call, time.Millisecond*10)
//	time.Sleep(time.Millisecond * 200)
//
//	if expected, got := WaitingAssignation, w.GetState(); expected != string(got) {
//		t.Errorf("state does not match, expected %s got %s", expected, got)
//	}
//}
//
//func TestItRequestAssignationToRemoteOnceWorkerIsAssigned(t *testing.T) {
//	call := &fakeCaller{}
//	w1Index := 1
//	w1IP := net.ParseIP("8.8.8.1")
//	w1Name := "worker-1"
//	w := newWorker(w1Index, w1Name, w1IP, call, time.Millisecond*10)
//
//	asg := &Workload{
//		Version: 1,
//		Jobs: []*ap.Job{
//			{Entry: "world", Tag: "aatoupdate_temp"},
//			{Entry: "world", Tag: "stream:sportnews1"},
//			{Entry: "world", Tag: "discussions"},
//			{Entry: "world", Tag: "comments"},
//		},
//	}
//	w.Assign(asg)
//
//	time.Sleep(time.Millisecond * 100)
//
//	if expected, got := WaitingSyncComplete, w.GetState(); expected != string(got) {
//		t.Errorf("state does not match, expected %s got %s", expected, got)
//	}
//	if atLeast, got := int32(1), atomic.LoadInt32(&call.assigns); got < atLeast {
//		t.Errorf("min Workloads calls not match, at least %d got %d", atLeast, got)
//	}
//}
//
//func TestItDoesNotAssignOnceRemoteAssignationMatches(t *testing.T) {
//	asg := &Workload{
//		Version: 1,
//		Jobs: []*ap.Job{
//			{Entry: "world", Tag: "aatoupdate_temp"},
//			{Entry: "world", Tag: "stream:sportnews1"},
//			{Entry: "world", Tag: "discussions"},
//			{Entry: "world", Tag: "comments"},
//		},
//	}
//	call := &fakeCaller{assignation: asg}
//	w1Index := 1
//	w1IP := net.ParseIP("8.8.8.1")
//	w1Name := "worker-1"
//	w := newWorker(w1Index, w1Name, w1IP, call, time.Millisecond*10)
//
//	w.Assign(asg)
//
//	time.Sleep(time.Millisecond * 100)
//
//	if expected, got := Synced, w.GetState(); expected != string(got) {
//		t.Errorf("state does not match, expected %s got %s", expected, got)
//	}
//	if atLeast, got := int32(0), atomic.LoadInt32(&call.assigns); got < atLeast {
//		t.Errorf("min Workloads calls not match, at least %d got %d", atLeast, got)
//	}
//}
////
//func TestItGetSyncedOnceAssignationMatchesToRemoteOnceWorkerIsAssigned(t *testing.T) {
//	call := &fakeCaller{}
//	w1Index := 1
//	w1IP := net.ParseIP("8.8.8.1")
//	w1Name := "worker-1"
//	w := newWorker(w1Index, w1Name, w1IP, call, time.Millisecond*10)
//	asg := &Workload{
//		Version: 1,
//		Jobs: []*ap.Job{
//			{Entry: "world", Tag: "aatoupdate_temp"},
//			{Entry: "world", Tag: "stream:sportnews1"},
//			{Entry: "world", Tag: "discussions"},
//			{Entry: "world", Tag: "comments"},
//		},
//	}
//
//	w.Assign(asg)
//
//	time.Sleep(time.Millisecond * 100)
//
//	if expected, got := Synced, w.GetState(); expected != string(got) {
//		t.Errorf("state does not match, expected %s got %s", expected, got)
//	}
//	if atLeast, got := int32(1), atomic.LoadInt32(&call.assigns); got < atLeast {
//		t.Errorf("min Workloads calls not match, at least %d got %d", atLeast, got)
//	}
//}

// @TODO: REMOVE IT!
func TestQueueBehaviour(t *testing.T) {
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
