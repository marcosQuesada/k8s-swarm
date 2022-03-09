package swarm

import (
	"fmt"
	"github.com/marcosQuesada/k8s-swarm/pkg/config"
	log "github.com/sirupsen/logrus"
	"math"
	"sync"
)

type state struct {
	setName string
	keySet  []config.Job
	config  *config.Workloads
	mutex   sync.RWMutex
}

func NewState(keySet []config.Job, setName string) *state {
	log.Infof("Mongo sharder initialized with %d total Workloads", len(keySet))
	return &state{
		keySet:  keySet,
		setName: setName,
		config:  &config.Workloads{Workloads: map[string]*config.Workload{}},
	}
}

func (a *state) BalanceWorkload(totalWorkers int, version int64) error {
	log.Infof("State balance started, Recalculate assignations total workers: %d", totalWorkers)

	a.mutex.Lock()
	defer a.mutex.Unlock()

	if totalWorkers == 0 {
		a.cleanAssignations(0)
		return nil
	}

	totalParts := int(math.Ceil(float64(len(a.keySet)) / float64(totalWorkers)))
	for i := 0; i < totalWorkers; i++ {
		workerName := fmt.Sprintf("%s_%d", a.setName, i)
		if _, ok := a.config.Workloads[workerName]; !ok {
			a.config.Workloads[workerName] = &config.Workload{}
		}

		start := i * totalParts
		end := (i + 1) * totalParts
		if end > len(a.keySet) {
			end = len(a.keySet)
		}

		a.config.Workloads[workerName] = &config.Workload{Jobs: a.keySet[start:end]}
	}

	a.config.Version = version

	// on Downscaling
	if totalWorkers < len(a.config.Workloads) {
		a.cleanAssignations(totalWorkers)
	}

	return nil
}

func (a *state) cleanAssignations(totalWorkers int) {
	orgSize := len(a.config.Workloads)
	for i := totalWorkers; i < orgSize; i++ {
		delete(a.config.Workloads, fmt.Sprintf("%s_%d", a.setName, i))
	}
}

func (a *state) Workload(workerIdx int) (*config.Workload, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	asg, ok := a.config.Workloads[fmt.Sprintf("%s_%d", a.setName, workerIdx)]
	if !ok {
		return nil, fmt.Errorf("Workloads not found on index %d", workerIdx)
	}

	return asg, nil
}

func (a *state) Workloads() map[string]*config.Workload {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.config.Workloads
}

func (a *state) size() int {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return len(a.config.Workloads)
}
