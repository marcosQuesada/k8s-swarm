package swarm

import (
	"context"
	"github.com/marcosQuesada/k8s-swarm/pkg/config"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
	"net"
	"sync"
	"time"
)

type Worker struct {
	Name        string
	IP          net.IP
	Index       int
	assignation *config.Workload
	version     int64
	state       State
	events      []Event
	stateMutex  sync.RWMutex
	stopChan    chan struct{}
	delegated   Delegated
}

func newWorker(idx int, name string, IP net.IP, d Delegated, freq time.Duration) *Worker {
	w := &Worker{
		Name:      name,
		IP:        IP,
		Index:     idx,
		state:     WaitingAssignation,
		events:    []Event{newStateEvent(Booting)},
		stopChan:  make(chan struct{}),
		delegated: d,
	}

	go wait.Until(w.conciliate, freq, w.stopChan)
	return w
}

func (w *Worker) conciliate() {
	w.stateMutex.Lock()
	defer w.stateMutex.Unlock()

	switch w.state {
	case WaitingAssignation:
		if w.assignation == nil {
			return
		}
		w.state = CheckingVersion
		w.events = append(w.events, newStateEvent(w.state))
	case CheckingVersion, WaitingSyncComplete:
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		asg, err := w.delegated.Assignation(ctx, w) // @TODO: NOT NEEDED ! REMOVE IT!
		if err != nil {
			log.Errorf("error notifying worker %s Workloads %v", w.Name, err)
			w.events = append(w.events, newErrorEvent(err))
			return
		}

		var ev State = Synced
		if !w.assignation.Equals(asg) {
			ev = WaitingSyncComplete
			if w.state == CheckingVersion {
				ev = Syncing
			}
			log.Infof("Worker %s on state %s goes to %s", w.Name, w.state, ev)
		}

		w.state = ev
		w.events = append(w.events, newStateEvent(w.state))
	case Synced:
		if w.version == w.assignation.Version {
			return
		}
		w.version = w.assignation.Version
		w.state = Synced
		w.events = append(w.events, newVersionUpdateEvent(w.version))
	}

	if w.state != Synced {
		log.Infof("Worker %s state %s", w.Name, w.state)
	}
}

func (w *Worker) Assign(a *config.Workload) {
	w.stateMutex.Lock()
	defer w.stateMutex.Unlock()

	w.assignation = a
	var st State = CheckingVersion
	// on synced slaves refresh config
	if w.state == Synced {
		st = Syncing
	}
	w.state = st
	w.events = append(w.events, newStateEvent(AssignationDefined))
	w.events = append(w.events, newStateEvent(st))
}

func (w *Worker) GetAssignation() *config.Workload {
	w.stateMutex.RLock()
	defer w.stateMutex.RUnlock()

	return w.assignation
}

func (w *Worker) GetVersion() int64 {
	w.stateMutex.RLock()
	defer w.stateMutex.RUnlock()

	return w.version
}

func (w *Worker) GetState() State {
	w.stateMutex.RLock()
	defer w.stateMutex.RUnlock()

	return w.state
}

func (w *Worker) Events() []Event {
	w.stateMutex.RLock()
	defer w.stateMutex.RUnlock()

	return w.events
}
func (w *Worker) Terminate() {
	close(w.stopChan)
}

type workerList []*Worker

func (e workerList) Len() int {
	return len(e)
}

func (e workerList) Less(i, j int) bool {
	return e[i].Index < e[j].Index
}

func (e workerList) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
