package cmd

import (
	"fmt"
	"github.com/gorilla/mux"
	cfg "github.com/marcosQuesada/k8s-swarm/pkg/config"
	swarm2 "github.com/marcosQuesada/k8s-swarm/services/controller/internal/app/swarm"
	"github.com/marcosQuesada/k8s-swarm/services/controller/internal/infra/k8s"
	operator2 "github.com/marcosQuesada/k8s-swarm/services/controller/internal/infra/k8s/operator"
	pod2 "github.com/marcosQuesada/k8s-swarm/services/controller/internal/infra/k8s/operator/pod"
	statefulset2 "github.com/marcosQuesada/k8s-swarm/services/controller/internal/infra/k8s/operator/statefulset"
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

// externalCmd represents the external command
var externalCmd = &cobra.Command{
	Use:   "external",
	Short: "swarm external controller, useful on development path",
	Long:  `swarm internal controller balance configured keys between swarm peers, useful on development path`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("controller external listening on namespace %s label %s Version %s release date %s http server on port %s", namespace, watchLabel, cfg.Commit, cfg.Date, cfg.HttpPort)

		st := swarm2.NewState(config.Jobs)

		cl := k8s.BuildExternalClient()

		nop := swarm2.NewExecutor(nil) // @TODO: HERE!
		app := swarm2.NewApp(st, nop)

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
	},
}

func init() {
	rootCmd.AddCommand(externalCmd)
}
