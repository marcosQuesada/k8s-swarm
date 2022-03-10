package pod

import (
	"context"
	"github.com/marcosQuesada/k8s-swarm/services/controller/internal/infra/k8s"
	"testing"
)

// @TODO: Use k8s client mock
func TestNewProvider_ItDeletesPodToRefreshByNewOne(t *testing.T) {
	t.Skip()
	var namespace = "swarm"

	clientset := k8s.BuildExternalClient()

	p := NewProvider(clientset, namespace)
	name := "swarm-worker-2"
	if err := p.RefreshPod(context.Background(), name); err != nil {
		t.Fatalf("unexepcted error refreshing pod %s, got %v", name, err)
	}
}
