package configmap

import (
	"context"
	"fmt"
	cfg "github.com/marcosQuesada/k8s-swarm/pkg/config"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const defaultConfigKey = "config.yml"

type Provider struct {
	client        kubernetes.Interface
	namespace     string
	configMapName string
	setName       string
}

func NewProvider(cl kubernetes.Interface, namespace, configMapName, setName string) *Provider {
	return &Provider{
		client:        cl,
		namespace:     namespace,
		configMapName: configMapName,
		setName:       setName,
	}
}

func (p *Provider) Set(ctx context.Context, a *cfg.Workloads) error {
	cm, err := p.client.CoreV1().ConfigMaps(p.namespace).Get(context.Background(), p.configMapName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to get config map %v", err)
	}

	raw, err := yaml.Marshal(a)
	if err != nil {
		return fmt.Errorf("unable to Marshall config map, error %v", err)
	}
	cm.Data[defaultConfigKey] = string(raw)
	// @TODO: CM encode and update
	_, err = p.client.CoreV1().ConfigMaps(p.namespace).Update(ctx, cm, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("unable to update config map %v", err)
	}

	return nil
}

func (p *Provider) decode(cm *v1.ConfigMap) (*cfg.Workloads, error) {
	data := make(map[interface{}]interface{})
	if err := yaml.Unmarshal([]byte(cm.Data[defaultConfigKey]), &data); err != nil {
		return nil, fmt.Errorf("unable to unmarshall, error %v", err)
	}

	c := &cfg.Workloads{}
	if err := mapstructure.Decode(data, c); err != nil {
		return nil, fmt.Errorf("unable to decode, error %v", err)
	}

	return c, nil
}
