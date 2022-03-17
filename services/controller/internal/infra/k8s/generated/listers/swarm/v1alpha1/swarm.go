/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/marcosQuesada/k8s-swarm/services/controller/internal/infra/k8s/apis/swarm/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// SwarmLister helps list Swarms.
// All objects returned here must be treated as read-only.
type SwarmLister interface {
	// List lists all Swarms in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Swarm, err error)
	// Swarms returns an object that can list and get Swarms.
	Swarms(namespace string) SwarmNamespaceLister
	SwarmListerExpansion
}

// swarmLister implements the SwarmLister interface.
type swarmLister struct {
	indexer cache.Indexer
}

// NewSwarmLister returns a new SwarmLister.
func NewSwarmLister(indexer cache.Indexer) SwarmLister {
	return &swarmLister{indexer: indexer}
}

// List lists all Swarms in the indexer.
func (s *swarmLister) List(selector labels.Selector) (ret []*v1alpha1.Swarm, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Swarm))
	})
	return ret, err
}

// Swarms returns an object that can list and get Swarms.
func (s *swarmLister) Swarms(namespace string) SwarmNamespaceLister {
	return swarmNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// SwarmNamespaceLister helps list and get Swarms.
// All objects returned here must be treated as read-only.
type SwarmNamespaceLister interface {
	// List lists all Swarms in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Swarm, err error)
	// Get retrieves the Swarm from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.Swarm, error)
	SwarmNamespaceListerExpansion
}

// swarmNamespaceLister implements the SwarmNamespaceLister
// interface.
type swarmNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Swarms in the indexer for a given namespace.
func (s swarmNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Swarm, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Swarm))
	})
	return ret, err
}

// Get retrieves the Swarm from the indexer for a given namespace and name.
func (s swarmNamespaceLister) Get(name string) (*v1alpha1.Swarm, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("swarm"), name)
	}
	return obj.(*v1alpha1.Swarm), nil
}
