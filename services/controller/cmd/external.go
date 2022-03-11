package cmd

import (
	"fmt"
	"github.com/gorilla/mux"
	cfg "github.com/marcosQuesada/k8s-swarm/pkg/config"
	swarm2 "github.com/marcosQuesada/k8s-swarm/services/controller/internal/app/swarm"
	"github.com/marcosQuesada/k8s-swarm/services/controller/internal/infra/k8s"
	operator2 "github.com/marcosQuesada/k8s-swarm/services/controller/internal/infra/k8s/operator"
	"github.com/marcosQuesada/k8s-swarm/services/controller/internal/infra/k8s/operator/configmap"
	pod2 "github.com/marcosQuesada/k8s-swarm/services/controller/internal/infra/k8s/operator/pod"
	statefulset2 "github.com/marcosQuesada/k8s-swarm/services/controller/internal/infra/k8s/operator/statefulset"
	ht "github.com/marcosQuesada/k8s-swarm/services/controller/internal/infra/transport/http"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/workqueue"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// externalCmd represents the external command
var externalCmd = &cobra.Command{
	Use:   "external",
	Short: "swarm external controller, useful on development path",
	Long:  `swarm internal controller balance configured keys between swarm peers, useful on development path`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("controller external listening on namespace %s label %s Version %s release date %s http server on port %s", namespace, watchLabel, cfg.Commit, cfg.Date, cfg.HttpPort)

		cl := k8s.BuildExternalClient()
		dl := configmap.NewProvider(cl, namespace, workersConfigMapName, watchLabel)
		vst := ht.NewVersionProvider(cfg.HttpPort)
		pdl := pod2.NewProvider(cl, namespace)
		ex := swarm2.NewExecutor(dl, vst, pdl)
		st := swarm2.NewState(config.Jobs, watchLabel)
		app := swarm2.NewWorkerPool(st, ex)

		podLwa := pod2.NewListWatcherAdapter(cl, namespace)
		podH := pod2.NewHandler(app)
		podSelector := operator2.NewSelector(watchLabel)
		podEventQueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
		podEventHandler := operator2.NewResourceEventHandler(podSelector, podEventQueue)
		podEvp := operator2.NewEventProcessor(&apiv1.Pod{}, podLwa, podEventHandler, podH)
		podCtl := operator2.NewController(podEvp, podEventQueue)

		stsLwa := statefulset2.NewListWatcherAdapter(cl, namespace)
		stsH := statefulset2.NewHandler(app)
		stsSelector := operator2.NewSelector(watchLabel)
		stsEventQueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
		stsEventHandler := operator2.NewResourceEventHandler(stsSelector, stsEventQueue)
		stsEvp := operator2.NewEventProcessor(&appsv1.StatefulSet{}, stsLwa, stsEventHandler, stsH)
		stsCtl := operator2.NewController(stsEvp, stsEventQueue)

		stopCh := make(chan struct{})
		go podCtl.Run(stopCh)
		go stsCtl.Run(stopCh)

		router := mux.NewRouter()
		srv := &http.Server{
			Addr:         fmt.Sprintf(":%s", cfg.HttpPort),
			Handler:      router,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}

		go func(h *http.Server) {
			log.Infof("starting server on port %s", cfg.HttpPort)
			e := h.ListenAndServe()
			if e != nil && e != http.ErrServerClosed {
				log.Fatalf("Could not Listen and server, error %v", e)
			}
		}(srv)

		sigTerm := make(chan os.Signal, 1)
		signal.Notify(sigTerm, syscall.SIGTERM, syscall.SIGINT)
		<-sigTerm
		if err := srv.Close(); err != nil {
			log.Errorf("unexpected error on http server close %v", err)
		}
		close(stopCh)
		log.Info("Stopping controller")
	},
}

func init() {
	rootCmd.AddCommand(externalCmd)
}
