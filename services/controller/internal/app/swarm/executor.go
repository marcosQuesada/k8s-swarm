package swarm

import (
	"context"
	ap "github.com/marcosQuesada/k8s-swarm/pkg/config"
	log "github.com/sirupsen/logrus"
	"net"
)

type Connector interface {
	Get(ctx context.Context) (version int64, include []*ap.Job, err error)
	Set(ctx context.Context, a *ap.Workload) error
}

type ConnCloser interface {
	Close() error
}

type configDumper interface {
	Build(ctx context.Context, workerName string, IP net.IP) (Connector, ConnCloser, error)
}

type provider interface {
	//Assignation(ctx context.Context, IP net.IP) (*VersionedWorkload, error)
}

type delegated interface {
}

type executor struct {
	builder configDumper
}

func NewExecutor(c configDumper) *executor {
	return &executor{
		builder: c,
	}
}

func (e *executor) Assign(ctx context.Context, w *ap.Workloads) (err error) {
	log.Infof("Trying to assign to %v", w)

	/*	conn, closer, err := e.builder.Build(ctx, w.Name, w.IP)
		if err != nil {
			return fmt.Errorf("unable to create client to IP %s error %v", w.IP, err)
		}

		defer func() {
			err = closer.Close()
			if err != nil {
				log.Errorf("unexpected error closing conn to %s, error: %v", w.IP.String(), err)
			}
			return
		}()

		if err := conn.Set(ctx, a); err != nil {
			return fmt.Errorf("unable to include on client to %s IP %s error %v", w.Name, w.IP, err)
		}*/

	return nil
}

func (e *executor) Assignation(ctx context.Context, w *Worker) (a *ap.Workload, err error) {
	log.Infof("Get config to %s IP %s", w.Name, w.IP.String())
	// http://localhost:9090/internal/version
	/*
		conn, closer, err := e.builder.Build(ctx, w.Name, w.IP)
		if err != nil {
			return nil, fmt.Errorf("unable to create client to IP %s error %v", w.IP, err)
		}

		defer func() {
			err = closer.Close()
			if err != nil {
				log.Errorf("unexpected error closing conn to %s, error: %v", w.IP.String(), err)
			}
			return
		}()

		version, inc, err := conn.Get(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to get config from client to %s IP %s error %v", w.Name, w.IP, err)
		}
	*/
	return &ap.Workload{ // @TODO: HERE!
		Jobs: nil,
	}, nil
}

func (e *executor) Restart(ctx context.Context, w *Worker) error {
	// @TODO:
	//data := fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"%s"}}}}}`, time.Now().String())
	//resultDeployment, err = p.Client.AppsV1().Deployments(p.Namespace).Patch(context.Background(), deployment.Name, types.StrategicMergePatchType, []byte(data), metav1.PatchOptions{FieldManager: "kubectl-rollout"})

	return nil
}
