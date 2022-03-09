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

type delegatedStorage interface {
	Get(ctx context.Context) (*ap.Workloads, error)
	Set(ctx context.Context, a *ap.Workloads) error
	RefreshWorkerPool(ctx context.Context) error
}

type executor struct {
	storage delegatedStorage
	remotes workerProvider
}

func NewExecutor(s delegatedStorage, p workerProvider) *executor {
	return &executor{
		storage: s,
		remotes: p,
	}
}

func (e *executor) Assign(ctx context.Context, w *ap.Workloads) (err error) {
	log.Infof("Trying to assign to %v", w)

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

func (e *executor) RestartWorkerPool(ctx context.Context) error {
	log.Info("Restarting worker ppol")
	return e.storage.RefreshWorkerPool(ctx)
}
