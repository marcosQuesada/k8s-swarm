package swarm

import (
	"context"
	"github.com/marcosQuesada/k8s-swarm/pkg/config"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
	"net"
	"sort"
	"sync"
	"time"
)

type assigner interface {
	BalanceWorkload(totalWorkers int, version int64) error
	Workload(workerIdx int) (*config.Workload, error) // @TODO: REMOVE IT!
	Workloads() *config.Workloads
}

type Delegated interface {
	Assign(ctx context.Context, w *config.Workloads) error
	Assignation(ctx context.Context, w *Worker) (*config.Workload, error)
	RestartWorkerPool(ctx context.Context) error
}

type pool struct {
	index          map[string]*Worker
	state          assigner
	delegated      Delegated
	version        int64
	expectedSize   int
	underVariation bool
	refreshedPool  bool
	stopChan       chan struct{}
	mutex          sync.RWMutex
}

func NewApp(cmp assigner, not Delegated) *pool {
	s := &pool{
		index:          make(map[string]*Worker),
		state:          cmp,
		delegated:      not,
		underVariation: true,
		stopChan:       make(chan struct{}),
	}

	go wait.Until(s.conciliate, DefaultWorkerFrequency, s.stopChan)

	return s
}

func (a *pool) UpdateExpectedSize(size int) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.expectedSize == size {
		return
	}

	a.expectedSize = size
	a.underVariation = true
	a.version++ // @TODO: HOW TO!

	log.Infof("Pool Version Update %d Size From %d to %d", a.version, a.expectedSize, size)

	if err := a.state.BalanceWorkload(size, a.version); err != nil {
		log.Errorf("err on balance started %v", err)
	}

	// @TODO
	log.Infof("state workload updarte to version %d, expected slaves: %d on index %d", a.version, a.expectedSize, len(a.index))
	if err := a.delegated.Assign(context.Background(), a.state.Workloads()); err != nil {
		log.Errorf("config error %v", err)
	}
}

func (a *pool) AddWorkerIfNotExists(idx int, name string, IP net.IP) bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if _, ok := a.index[name]; ok {
		log.Infof("pod %s already on pool", name)
		return false
	}

	// @TODO. Explicit ScaleUp Mode
	a.index[name] = newWorker(idx, name, IP, a.delegated, DefaultWorkerFrequency)

	log.Debugf("Added Worker to Pool Name %s IP %s length %d, expectedSize %d", name, IP, len(a.index), a.expectedSize)

	return true
}

func (a *pool) RemoveWorkerByName(name string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	log.Infof("Removing Worker from Pool Name %s", name)

	w, ok := a.index[name]
	if !ok {
		return
	}

	w.Terminate()
	delete(a.index, name)

	// @TODO. Explicit ScaleDown Mode
	if len(a.index) != a.expectedSize {
		a.underVariation = true
	}
}

func (a *pool) Terminate() {
	close(a.stopChan)
}

func (a *pool) Events() map[string][]Event {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	e := map[string][]Event{}
	for _, worker := range a.geAllWorkers() {
		e[worker.Name] = worker.Events()
	}

	return e
}

func (a *pool) conciliate() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.expectedSize == 0 {
		return
	}

	if !a.underVariation {
		return
	}

	log.Infof("state contiliation version %d, expected slaves: %d on index %d", a.version, a.expectedSize, len(a.index))

	if !a.refreshedPool {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if err := a.delegated.RestartWorkerPool(ctx); err != nil {
			log.Errorf("unable to refresh worker pool, error %v", err)
		} else {
			a.refreshedPool = true
		}
	}

	for _, w := range a.geAllWorkers() {
		asg, err := a.state.Workload(w.Index)
		if err != nil {
			log.Errorf("unexpected error getting config %v", err)
			return
		}

		if asg.Equals(w.GetAssignation()) {
			continue
		}

		log.Infof("Conciliation loop, config to slave %s on Version %d", w.Name, a.version)
		w.Assign(asg)
	}

	if a.expectedSize != len(a.index) {
		log.Infof("Pool still on variation, expected %d got %d", a.expectedSize, len(a.index))
		return
	}

	// ensure all workers are in the same version
	for _, w := range a.index {
		wv := w.GetVersion()
		if wv != a.version {
			log.Infof("Pool still on variation, worker %s still on version %d", w.Name, wv)
			return
		}
	}

	log.Info("Stopping conciliation loop, variation completed!")
	a.underVariation = false
}

func (a *pool) geAllWorkers() []*Worker {
	var res workerList
	for _, w := range a.index {
		res = append(res, w)
	}
	sort.Sort(res)

	return res
}
