package app

import (
	cfg "github.com/marcosQuesada/k8s-swarm/pkg/config"
	log "github.com/sirupsen/logrus"
	"sync"
)

type App struct {
	state []cfg.Job
	mutex sync.Mutex
}

func NewApp() *App {
	return &App{}
}

func (a *App) Assign(version int64, keySet []cfg.Job) error {
	log.Infof("App Workloads updated version %d Workloads %v", version, keySet)

	return nil
}

func (a *App) Run() {
	// @TODO: Pending to implement
}

func (a *App) Terminate() {

}
