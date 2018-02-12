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

package v1

import (
	v1 "github.com/vgkowski/cassandra-operator/pkg/apis/cassandra/v1"
	"github.com/vgkowski/cassandra-operator/pkg/client/clientset/versioned/scheme"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	rest "k8s.io/client-go/rest"
)

type CassandraV1Interface interface {
	RESTClient() rest.Interface
	CassandraClustersGetter
}

// CassandraV1Client is used to interact with features provided by the cassandra group.
type CassandraV1Client struct {
	restClient rest.Interface
}

func (c *CassandraV1Client) CassandraClusters(namespace string) CassandraClusterInterface {
	return newCassandraClusters(c, namespace)
}

// NewForConfig creates a new CassandraV1Client for the given config.
func NewForConfig(c *rest.Config) (*CassandraV1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &CassandraV1Client{client}, nil
}

// NewForConfigOrDie creates a new CassandraV1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *CassandraV1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new CassandraV1Client for the given RESTClient.
func New(c rest.Interface) *CassandraV1Client {
	return &CassandraV1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *CassandraV1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
