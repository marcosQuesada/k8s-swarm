package cmd

import (
	"fmt"
	"github.com/gorilla/mux"
	config2 "github.com/marcosQuesada/k8s-swarm/pkg/config"
	ht "github.com/marcosQuesada/k8s-swarm/pkg/infra/transport/http"
	swarm2 "github.com/marcosQuesada/k8s-swarm/services/controller/internal/app/swarm"
	"github.com/marcosQuesada/k8s-swarm/services/controller/internal/infra/k8s"
	operator2 "github.com/marcosQuesada/k8s-swarm/services/controller/internal/infra/k8s/operator"
	"github.com/marcosQuesada/k8s-swarm/services/controller/internal/infra/k8s/operator/configmap"
	pod2 "github.com/marcosQuesada/k8s-swarm/services/controller/internal/infra/k8s/operator/pod"
	statefulset2 "github.com/marcosQuesada/k8s-swarm/services/controller/internal/infra/k8s/operator/statefulset"
	cht "github.com/marcosQuesada/k8s-swarm/services/controller/internal/infra/transport/http"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// internalCmd represents the internal command
var internalCmd = &cobra.Command{
	Use:   "internal",
	Short: "swarm internal controller",
	Long:  `swarm internal controller balance configured keys between swarm peers`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("controller internal listening on namespace %s label %s Version %s release date %s http server on port %s", namespace, watchLabel, config2.Commit, config2.Date, config2.HttpPort)

		cl := k8s.BuildInternalClient()

		dl := configmap.NewProvider(cl, namespace, workersConfigMapName, watchLabel)
		vst := cht.NewVersionProvider(config2.HttpPort)
		pdl := pod2.NewProvider(cl, namespace)
		ex := swarm2.NewExecutor(dl, vst, pdl)
		st := swarm2.NewState(config.Jobs, watchLabel)
		app := swarm2.NewWorkerPool(st, ex)

		podCla := pod2.NewListWatcherAdapter(cl, namespace)
		podH := pod2.NewHandler(app)
		podFilter := operator2.NewFilter(watchLabel, &apiv1.Pod{})
		podCtl := operator2.Build(podH, podFilter, podCla)

		stlCla := statefulset2.NewListWatcherAdapter(cl, namespace)
		stlH := statefulset2.NewHandler(app)
		stsFilter := operator2.NewFilter(watchLabel, &appsv1.StatefulSet{})
		stlCtl := operator2.Build(stlH, stsFilter, stlCla)

		stopCh := make(chan struct{})
		defer close(stopCh)
		go podCtl.Run(podControllerWorkers, stopCh)
		go stlCtl.Run(stsControllerWorkers, stopCh)

		router := mux.NewRouter()
		ch := ht.NewChecker(config2.Commit, config2.Date)
		ch.Routes(router)

		srv := &http.Server{
			Addr:         fmt.Sprintf(":%s", config2.HttpPort),
			Handler:      router,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}

		go func(h *http.Server) {
			log.Infof("starting server on port %s", config2.HttpPort)
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
	},
}

func init() {
	rootCmd.AddCommand(internalCmd)
}
