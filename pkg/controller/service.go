package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/api/core/v1"
	cassandrav1 "github.com/vgkowski/cassandra-operator/pkg/apis/cassandra/v1"

)

func (c *Controller) DeleteService(svcName string) error{
	err := c.kubeClientset.CoreV1().Services(c.namespace).Delete(svcName, &metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		err = nil
	}
	return err
}

func (c *Controller) CreateOrUpdateService(svc *v1.Service) error {
	client := c.kubeClientset.CoreV1().Services(c.namespace)
	service, err := client.Get(svc.Name, metav1.GetOptions{})
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

func (c *Controller) BuildService(cc *cassandrav1.CassandraCluster) *v1.Service{
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: cc.Name,
			Annotations: map[string]string{
				// todo: add operator name in labels
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

