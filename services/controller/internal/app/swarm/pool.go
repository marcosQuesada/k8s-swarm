package swarm

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/k8s-swarm/pkg/config"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
	"net"
	"sort"
	"sync"
	"time"
)

const defaultSleepBeforeNotify = time.Second * 5

type assigner interface {
	BalanceWorkload(totalWorkers int, version int64) error
	Workloads() *config.Workloads
}

type Delegated interface {
	Assign(ctx context.Context, w *config.Workloads) error
	Assignation(ctx context.Context, w *Worker) (*config.Workload, error)
	RestartWorker(ctx context.Context, name string) error
}

type pool struct {
	index          map[string]*Worker
	state          assigner
	delegated      Delegated
	version        int64
	expectedSize   int
	underVariation bool
	refreshedPool  bool // @TODO: Rethink! Note Scheduled execution Ts
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

func (a *pool) UpdateExpectedSize(newSize int) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.expectedSize == newSize {
		return
	}

	previousSize := a.expectedSize
	a.expectedSize = newSize
	a.underVariation = true
	a.version++

	log.Infof("Pool Version Update %d Size From %d to %d", a.version, previousSize, newSize)

	if err := a.state.BalanceWorkload(newSize, a.version); err != nil {
		log.Errorf("err on balance started %v", err)
	}

	if err := a.delegated.Assign(context.Background(), a.state.Workloads()); err != nil {
		log.Errorf("config error %v", err)
	}

	if previousSize == 0 {
		return
	}

	ws := a.geAllWorkers()
	totalToRefresh := newSize
	if newSize > previousSize {
		totalToRefresh = previousSize
	}
	a.refreshedPool = false
	log.Infof("Total %d Workers marked to Refresh %d to version %d", len(ws), totalToRefresh, a.version)
	for i := 0; i < totalToRefresh; i++ {
		if i > len(ws)-1 { // @TODO: TEST IT PROPERLY!
			return
		}
		w := ws[i]
		w.MarkToRefresh()
	}
}

func (a *pool) AddWorkerIfNotExists(idx int, name string, IP net.IP) bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if _, ok := a.index[name]; ok {
		log.Infof("pod %s already on pool", name)
		return false
	}

	a.index[name] = newWorker(idx, name, IP, a.delegated)

	log.Debugf("Added Worker to Pool Name %s IP %s length %d, expectedSize %d", name, IP, len(a.index), a.expectedSize)

	return true
}

func (a *pool) RemoveWorkerByName(name string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	w, ok := a.index[name]
	if !ok {
		return
	}

	log.Infof("Removing Worker %s from Pool Name", name)
	w.Terminate()
	delete(a.index, name)

	if len(a.index) != a.expectedSize {
		a.underVariation = true
	}
}

func (a *pool) Terminate() {
	close(a.stopChan)
}

func (a *pool) worker(name string) (*Worker, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	v, ok := a.index[name]
	if !ok {
		log.Infof("pod %s already on pool", name)
		return nil, fmt.Errorf("no worker %s found", name)
	}
	return v, nil
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

	if a.refreshedPool {
		a.underVariation = false
		return
	}

	for _, w := range a.geAllWorkers() {
		if !w.NeedsRefresh() {
			continue
		}

		log.Infof("Request scheduled restart to %s", w.Name)
		go a.requestRestart(context.Background(), w)
	}
	a.refreshedPool = true

	log.Info("Stopping conciliation loop, variation completed!")
	a.underVariation = false
}

func (a *pool) requestRestart(ctx context.Context, w *Worker) error {
	log.Infof("Scheduling worker %s refresh", w.Name)
	time.Sleep(defaultSleepBeforeNotify)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := w.delegated.RestartWorker(ctx, w.Name); err != nil {
		log.Errorf("unable to restart worker, error %v", err)
		return err
	}
	w.MarkRefreshed()
	return nil
}

func (a *pool) geAllWorkers() []*Worker {
	var res workerList
	for _, w := range a.index {
		res = append(res, w)
	}
	sort.Sort(res)

	return res
}
