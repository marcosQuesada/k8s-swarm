package configmap

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/marcosQuesada/k8s-swarm/pkg/config"
	"github.com/marcosQuesada/k8s-swarm/services/controller/internal/infra/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestNewProvider_ItUpdatesConfigMapOnAssignWorkload(t *testing.T) {
	var namespace = "swarm"
	var configMapName = "swarm-worker-config"
	var workerPoolName = "swarm-worker"

	clientset := k8s.BuildExternalClient()
	cm, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.Background(), configMapName, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	spew.Dump(cm.Data)

	p := NewProvider(clientset, namespace, configMapName, workerPoolName)
	w := &config.Workloads{
		Version: 1,
		Workloads: map[string]*config.Workload{
			"foo_0": {Jobs: []config.Job{"rtve1", "cctv1", "euktv", "ccmeg01"}},
			"foo_1": {Jobs: []config.Job{"zoom0", "zrtve1", "zcctv1", "zeuktv"}},
			"foo_2": {Jobs: []config.Job{"xfoo", "xrtve1", "xcctv1", "xeuktv"}},
		},
	}

	if err := p.Set(context.Background(), w); err != nil {
		t.Fatalf("unexepcted error setting workload %v, got %v", w, err)
	}
}
