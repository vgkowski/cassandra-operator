package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (c *Controller) DeletePVC(name string) error{
	pvcClient := c.kubeClientset.CoreV1().PersistentVolumeClaims(c.namespace)
	// use the owner of the PVC to get PVC associated with cassandraCluster
	pvcs,err := pvcClient.List(metav1.ListOptions{LabelSelector: "cassandraCluster="+name})
	if err != nil {
		return err
	}

	nbPVC := len(pvcs.Items)
	for i := 0 ; i < nbPVC ; i++ {
		err := pvcClient.Delete(pvcs.Items[i].Name, &metav1.DeleteOptions{})
		if errors.IsNotFound(err) {
			err = nil
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}
