/*
Copyright 2018 The Kubernetes Authors.

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

// This file was automatically generated by informer-gen

package v1

import (
	cassandra_v1 "github.com/vgkowski/cassandra-operator/pkg/apis/cassandra/v1"
	versioned "github.com/vgkowski/cassandra-operator/pkg/client/clientset/versioned"
	internalinterfaces "github.com/vgkowski/cassandra-operator/pkg/client/informers/externalversions/internalinterfaces"
	v1 "github.com/vgkowski/cassandra-operator/pkg/client/listers/cassandra/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
	time "time"
)

// CassandraClusterInformer provides access to a shared informer and lister for
// CassandraClusters.
type CassandraClusterInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.CassandraClusterLister
}

type cassandraClusterInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewCassandraClusterInformer constructs a new informer for CassandraCluster type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewCassandraClusterInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredCassandraClusterInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredCassandraClusterInformer constructs a new informer for CassandraCluster type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredCassandraClusterInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CassandraV1().CassandraClusters(namespace).List(options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CassandraV1().CassandraClusters(namespace).Watch(options)
			},
		},
		&cassandra_v1.CassandraCluster{},
		resyncPeriod,
		indexers,
	)
}

func (f *cassandraClusterInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredCassandraClusterInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *cassandraClusterInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&cassandra_v1.CassandraCluster{}, f.defaultInformer)
}

func (f *cassandraClusterInformer) Lister() v1.CassandraClusterLister {
	return v1.NewCassandraClusterLister(f.Informer().GetIndexer())
}
