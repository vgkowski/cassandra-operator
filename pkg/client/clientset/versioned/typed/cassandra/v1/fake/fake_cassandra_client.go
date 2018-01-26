package fake

import (
	v1 "github.com/vgkowski/cassandra-operator/pkg/client/clientset/versioned/typed/cassandra/v1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeCassandraV1 struct {
	*testing.Fake
}

func (c *FakeCassandraV1) CassandraClusters(namespace string) v1.CassandraClusterInterface {
	return &FakeCassandraClusters{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeCassandraV1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
