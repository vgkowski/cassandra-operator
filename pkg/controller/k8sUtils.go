package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/apimachinery/pkg/api/errors"
	"fmt"
	"bytes"
	"k8s.io/api/core/v1"
)

func (c *Controller) DeleteStatefulSet(stsName string) error{
	err := c.kubeClientset.AppsV1().StatefulSets(c.namespace).Delete(stsName, &metav1.DeleteOptions{
		PropagationPolicy: func() *metav1.DeletionPropagation {
			foreground := metav1.DeletePropagationForeground
			return &foreground
		}(),
	})
	if errors.IsNotFound(err) {
		err = nil
	}
	return err
}

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

func (c *Controller) DeleteService(svcName string) error{
	err := c.kubeClientset.CoreV1().Services(c.namespace).Delete(svcName, &metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		err = nil
	}
	return err
}

func (c *Controller) ExecCmd(podName string,cmd [] string) (string,string,error) {
	// get the pod from the name
	pod, err := c.kubeClientset.CoreV1().Pods(c.namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		return "", "",fmt.Errorf("could not get pod info: %v", err)
	}
	if len(pod.Spec.Containers) != 1 {
		return "", "", fmt.Errorf("could not determine which container to use")
	}

	// build the remoteexec
	req := c.kubeClientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(c.namespace).
		SubResource("exec")
	req.VersionedParams(&v1.PodExecOptions{
		Container: pod.Spec.Containers[0].Name,
		Command:   cmd,
		Stdin:	   false,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return "", "",fmt.Errorf("could not init remote executor: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout:             &stdout,
		Stderr:             &stderr,
		Tty:                false,
	})

	return stdout.String(), stderr.String(),err
}