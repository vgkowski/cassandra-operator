package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/api/core/v1"
	cassandrav1 "github.com/vgkowski/cassandra-operator/pkg/apis/cassandra/v1"

	"encoding/xml"
)

func (c *Controller) DeleteService(svcName string) error{
	err := c.kubeClientset.CoreV1().Services(c.namespace).Delete(svcName, &metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		err = nil
	}
	return err
}

func (c *Controller) CreateOrUpdateService(cc *cassandrav1.CassandraCluster) error {
	// build the service
	svc := c.BuildHeadlessService(cc)

	client := c.kubeClientset.CoreV1().Services(c.namespace)
	// TODO use lister instead of direct query ?
	service, err := c.servicesLister.Services(c.namespace).Get(cc.Name)
	//service, err := client.Get(svc.Name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if errors.IsNotFound(err) {
		_, err = client.Create(svc)
		if err != nil {
			return err
		}
	} else {
		svc.ResourceVersion = service.ResourceVersion
		svc.Spec.ClusterIP = service.Spec.ClusterIP
		_, err := client.Update(svc)
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) BuildHeadlessService(cc *cassandrav1.CassandraCluster) *v1.Service{
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: cc.Name+"-node",
			Annotations: map[string]string{
				"operatorVersion": cassandrav1.CassandraCluster.APIVersion,

			},
			Labels: map[string]string{
				"cassandraCluster": cc.Name,
				"role": "cassandraCluster",
			},
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"cassandraCluster": cc.Name,
			},
			Ports: []v1.ServicePort{
				{
					Name: "cassandra",
					Port: 9042,
				},
			},
			ClusterIP: "None",
		},
	}
	return service
}

