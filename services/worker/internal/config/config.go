package config

import (
	"fmt"
	"github.com/marcosQuesada/k8s-swarm/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"sync"
)

var (
	// keySetConfig defines static worker Workload Workloads
	keySetConfig Config
	keySetMutex  sync.RWMutex
)

// Config loaded app config
type Config struct {
	Workload map[string]*config.Workload `mapstructure:"workload"`
	Version  int64                       `mapstructure:"version"`
}

func (cfg *Config) HostKeySet(host string) ([]config.Job, error) {
	v, ok := cfg.Workload[host]
	if !ok {
		return nil, fmt.Errorf("unable to found host %s keySet config", host)
	}

	log.Infof("Loaded %d keys on Host %s", len(v.Jobs), host)

	return v.Jobs, nil
}

// HostKeySet gets assignation Workload to the Host
func HostKeySet(host string) ([]config.Job, error) {
	keySetMutex.RLock()
	defer keySetMutex.RUnlock()

	return keySetConfig.HostKeySet(host)
}

// Version reply version config
func Version() int64 {
	keySetMutex.RLock()
	defer keySetMutex.RUnlock()

	return keySetConfig.Version
}

type VersionAdapter struct {
	hostName string
}

func NewVersionAdapter(host string) *VersionAdapter {
	return &VersionAdapter{
		hostName: host,
	}
}

func (v *VersionAdapter) Version() int64 {
	return Version()
}

func (v *VersionAdapter) Workload() []config.Job {
	keySet, err := HostKeySet(v.hostName)
	if err != nil {
		return []config.Job{}
	}
	return keySet
}

func HostName(defaultHostName string) string {
	p := os.Getenv("HOSTNAME")
	if p == "" {
		return defaultHostName
	}

	return p
}

func LoadConfig(configFilePath, configFile string) error {
	viper.AddConfigPath(configFilePath)
	viper.SetConfigName(configFile)
	viper.SetConfigType("yml")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("unable to load config file %s from path %s, error %v", configFile, configFilePath, err)
	}

	log.Infof("Using config file: %s", viper.ConfigFileUsed())

	keySetMutex.Lock()
	defer keySetMutex.Unlock()
	if err := viper.Unmarshal(&keySetConfig); err != nil {
		return fmt.Errorf("unable to unMarshall config, error %v", err)
	}

	viper.WatchConfig()

	return nil
}
