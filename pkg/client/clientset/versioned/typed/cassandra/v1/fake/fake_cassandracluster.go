package fake

import (
	cassandra_v1 "github.com/vgkowski/cassandra-operator/pkg/apis/cassandra/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeCassandraClusters implements CassandraClusterInterface
type FakeCassandraClusters struct {
	Fake *FakeCassandraV1
	ns   string
}

var cassandraclustersResource = schema.GroupVersionResource{Group: "cassandra", Version: "v1", Resource: "cassandraclusters"}

var cassandraclustersKind = schema.GroupVersionKind{Group: "cassandra", Version: "v1", Kind: "CassandraCluster"}

// Get takes name of the cassandraCluster, and returns the corresponding cassandraCluster object, and an error if there is any.
func (c *FakeCassandraClusters) Get(name string, options v1.GetOptions) (result *cassandra_v1.CassandraCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(cassandraclustersResource, c.ns, name), &cassandra_v1.CassandraCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*cassandra_v1.CassandraCluster), err
}

// List takes label and field selectors, and returns the list of CassandraClusters that match those selectors.
func (c *FakeCassandraClusters) List(opts v1.ListOptions) (result *cassandra_v1.CassandraClusterList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(cassandraclustersResource, cassandraclustersKind, c.ns, opts), &cassandra_v1.CassandraClusterList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &cassandra_v1.CassandraClusterList{}
	for _, item := range obj.(*cassandra_v1.CassandraClusterList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested cassandraClusters.
func (c *FakeCassandraClusters) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(cassandraclustersResource, c.ns, opts))

}

// Create takes the representation of a cassandraCluster and creates it.  Returns the server's representation of the cassandraCluster, and an error, if there is any.
func (c *FakeCassandraClusters) Create(cassandraCluster *cassandra_v1.CassandraCluster) (result *cassandra_v1.CassandraCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(cassandraclustersResource, c.ns, cassandraCluster), &cassandra_v1.CassandraCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*cassandra_v1.CassandraCluster), err
}

// Update takes the representation of a cassandraCluster and updates it. Returns the server's representation of the cassandraCluster, and an error, if there is any.
func (c *FakeCassandraClusters) Update(cassandraCluster *cassandra_v1.CassandraCluster) (result *cassandra_v1.CassandraCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(cassandraclustersResource, c.ns, cassandraCluster), &cassandra_v1.CassandraCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*cassandra_v1.CassandraCluster), err
}

// Delete takes name of the cassandraCluster and deletes it. Returns an error if one occurs.
func (c *FakeCassandraClusters) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(cassandraclustersResource, c.ns, name), &cassandra_v1.CassandraCluster{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCassandraClusters) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(cassandraclustersResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &cassandra_v1.CassandraClusterList{})
	return err
}

// Patch applies the patch and returns the patched cassandraCluster.
func (c *FakeCassandraClusters) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *cassandra_v1.CassandraCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(cassandraclustersResource, c.ns, name, data, subresources...), &cassandra_v1.CassandraCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*cassandra_v1.CassandraCluster), err
}
