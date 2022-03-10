package swarm

import (
	"github.com/marcosQuesada/k8s-swarm/pkg/config"
	log "github.com/sirupsen/logrus"
	"net"
	"sync"
)

type Worker struct {
	Name        string
	IP          net.IP
	Index       int
	assignation *config.Workload
	version     int64
	state       State
	stateMutex  sync.RWMutex
	delegated   Delegated
}

func newWorker(idx int, name string, IP net.IP, d Delegated) *Worker {
	return &Worker{
		Name:      name,
		IP:        IP,
		Index:     idx,
		state:     WaitingAssignation,
		delegated: d,
	}
}

func (w *Worker) NeedsRefresh() bool {
	w.stateMutex.RLock()
	defer w.stateMutex.RUnlock()

	return w.state == NeedsRefresh
}

func (w *Worker) MarkToRefresh() {
	log.Infof("worker %s marked to refresh", w.Name)
	w.stateMutex.Lock()
	defer w.stateMutex.Unlock()
	w.state = NeedsRefresh
}

func (w *Worker) MarkRefreshed() {
	log.Infof("worker %s marked refreshed", w.Name)
	w.stateMutex.Lock()
	defer w.stateMutex.Unlock()
	w.state = Syncing
}

func (w *Worker) Assign(a *config.Workload) {
	w.stateMutex.Lock()
	defer w.stateMutex.Unlock()

	w.assignation = a
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

func (w *Worker) Terminate() {
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
