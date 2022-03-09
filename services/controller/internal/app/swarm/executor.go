package swarm

import (
	"context"
	"fmt"
	ap "github.com/marcosQuesada/k8s-swarm/pkg/config"
	log "github.com/sirupsen/logrus"
	"net"
)

type workerProvider interface {
	Assignation(ctx context.Context, IP net.IP) (*ap.Workload, error)
}

type workerManager interface {
	RefreshPod(ctx context.Context, name string) error
}

type delegatedStorage interface {
	Get(ctx context.Context) (*ap.Workloads, error)
	Set(ctx context.Context, a *ap.Workloads) error
	RefreshWorkerPool(ctx context.Context) error
}

type executor struct {
	storage delegatedStorage
	remotes workerProvider
	manager workerManager
}

func NewExecutor(s delegatedStorage, p workerProvider, m workerManager) *executor {
	return &executor{
		storage: s,
		remotes: p,
		manager: m,
	}
}

func (e *executor) Assign(ctx context.Context, w *ap.Workloads) (err error) {
	log.Infof("Persist Workload version %d to assign to %v", w.Version, w.Workloads)
	return e.storage.Set(ctx, w)
}

func (e *executor) Assignation(ctx context.Context, w *Worker) (a *ap.Workload, err error) {
	log.Infof("Get config to %s IP %s", w.Name, w.IP.String())
	res, err := e.remotes.Assignation(ctx, w.IP)
	if err != nil {
		return nil, fmt.Errorf("unable to get remote assignation on %s error %v", w.Name, err)
	}
	log.Infof("Received config from %s IP %s version %d jobs %v", w.Name, w.IP.String(), res.Version, len(res.Jobs))
	return res, nil
}

func (e *executor) RestartWorker(ctx context.Context, name string) error {
	log.Infof("Restarting worker %s", name)
	return e.manager.RefreshPod(ctx, name)
}
