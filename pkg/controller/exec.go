package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/kubernetes/scheme"
	"fmt"
	"bytes"
	"k8s.io/api/core/v1"
)

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
