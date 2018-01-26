package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"

)

func (c *Controller) deleteCassandraCluster(name string) error {
	// deleted the statefulset
	err := c.DeleteStatefulSet(name)
	if err != nil {
		return err
	}
	// delete the pvc
	err = c.DeletePVC(name)
	if err != nil {
		return err
	}
	// delete the service
	err = c.DeleteService(name)
	if err != nil {
		return err
	}
	return err
}

func (c *Controller) createCassandraCluster(name string) error {
	// create the statefulset object
	err := c.DeleteStatefulSet(name)
	if err != nil {
		return err
	}
	err = c.DeletePVC(name)
	if err != nil {
		return err
	}
	err = c.DeleteService(name)
	if err != nil {
		return err
	}
	return err
}

