package configmap

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/marcosQuesada/k8s-swarm/pkg/config"
	"github.com/marcosQuesada/k8s-swarm/services/controller/internal/infra/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

// @TODO: Use k8s client mock
func TestNewProvider_ItUpdatesConfigMapOnAssignWorkload(t *testing.T) {
	t.Skip()
	var namespace = "swarm"
	var configMapName = "swarm-worker-config"
	var deploymentName = "swarm-worker"

	clientset := k8s.BuildExternalClient()
	cm, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.Background(), configMapName, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	spew.Dump(cm.Data)

	p := NewProvider(clientset, namespace, configMapName, deploymentName)
	w := &config.Workloads{
		Version: 1,
		Workloads: map[string]*config.Workload{
			"swarm-worker-0": {Jobs: []config.Job{"rtve1", "cctv1", "euktv", "ccmeg01"}},
			"swarm-worker-1": {Jobs: []config.Job{"zoom0", "zrtve1", "zcctv1", "zeuktv"}},
			"swarm-worker-2": {Jobs: []config.Job{"xfoo", "xrtve1", "xcctv1", "xeuktv"}},
		},
	}

	if err := p.Set(context.Background(), w); err != nil {
		t.Fatalf("unexepcted error setting workload %v, got %v", w, err)
	}
}

// @TODO: Use k8s client mock
func TestNewProvider_ItGetsWorkloadsFromConfigMap(t *testing.T) {
	t.Skip()
	var namespace = "swarm"
	var configMapName = "swarm-worker-config"
	var deploymentName = "swarm-worker"

	clientset := k8s.BuildExternalClient()
	cm, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.Background(), configMapName, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	spew.Dump(cm.Data)

	p := NewProvider(clientset, namespace, configMapName, deploymentName)
	w, err := p.Get(context.Background())
	if err != nil {
		t.Fatalf("unexepcted error setting workload %v, got %v", w, err)
	}

	spew.Dump(w)
}
