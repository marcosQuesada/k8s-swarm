package swarm

import (
	"fmt"
	"github.com/marcosQuesada/k8s-swarm/pkg/config"
	log "github.com/sirupsen/logrus"
	"sync"
)

type state struct {
	setName string
	jobs    []config.Job
	config  *config.Workloads
	mutex   sync.RWMutex
}

func NewState(keySet []config.Job, setName string) *state {
	log.Infof("Swarm sharder initialized with %d total Workloads", len(keySet))
	return &state{
		jobs:    keySet,
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

	partSize := len(a.jobs) / totalWorkers
	modulePartSize := len(a.jobs) % totalWorkers
	for i := 0; i < totalWorkers; i++ {
		workerName := fmt.Sprintf("%s-%d", a.setName, i)
		if _, ok := a.config.Workloads[workerName]; !ok {
			a.config.Workloads[workerName] = &config.Workload{}
		}

		size := partSize
		if i < modulePartSize {
			size++
		}
		start := i * size
		end := (i + 1) * size
		if end > len(a.jobs) {
			end = len(a.jobs)
		}

		a.config.Workloads[workerName] = &config.Workload{Jobs: a.jobs[start:end]}

		log.Infof("Worker %s total jobs %d", workerName, len(a.jobs[start:end]))
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
		delete(a.config.Workloads, fmt.Sprintf("%s-%d", a.setName, i))
	}
}

func (a *state) Workloads() *config.Workloads {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.config
}

func (a *state) Workload(workerIdx int) (*config.Workload, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	asg, ok := a.config.Workloads[fmt.Sprintf("%s-%d", a.setName, workerIdx)]
	if !ok {
		return nil, fmt.Errorf("Workloads not found on index %d", workerIdx)
	}

	return asg, nil
}

func (a *state) size() int {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return len(a.config.Workloads)
}
