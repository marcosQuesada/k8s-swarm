package cmd

import (
	"context"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/marcosQuesada/k8s-swarm/services/controller/internal/infra/k8s"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

// configmaperCmd represents the configmaper command
var configmaperCmd = &cobra.Command{
	Use:   "configmaper",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("configmaper called")
		var namespace string = "swarm"
		var configMapName string = "swarm-worker-config"

		clientset := k8s.BuildExternalClient()

		cm, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.Background(), configMapName, metav1.GetOptions{})
		if err != nil {
			log.Fatal(err)
		}
		spew.Dump(cm.Data)

		//allConfig := cfg.Config{}
		//
		//data := make(map[interface{}]interface{})
		//if err := yaml.Unmarshal([]byte(cm.Data["config.yml"]), &data); err != nil {
		//	log.Fatal(err)
		//}
		//
		//if err := mapstructure.Decode(data, &allConfig); err != nil {
		//	log.Fatal(err)
		//}
		//
		//cm.Data["foo"] = "zoom"
		//_, err = clientset.CoreV1().ConfigMaps(namespace).Update(context.Background(), cm, metav1.UpdateOptions{})
		//if err != nil {
		//	log.Fatal(err)
		//}

		deploymentName := "swarm-worker"
		data := fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"%s"}}}}}`, time.Now().String())
		resultDeployment, err := clientset.AppsV1().Deployments(namespace).Patch(context.Background(), deploymentName, types.StrategicMergePatchType, []byte(data), metav1.PatchOptions{FieldManager: "kubectl-rollout"})
		if err != nil {
			log.Fatal(err)
		}
		spew.Dump(resultDeployment)

	},
}

func init() {
	rootCmd.AddCommand(configmaperCmd)
}
