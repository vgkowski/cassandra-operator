package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
	"k8s.io/apimachinery/pkg/api/resource"
	cassandrav1 "github.com/vgkowski/cassandra-operator/pkg/apis/cassandra/v1"

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

func (c *Controller) CreateOrUpdateStatefulSet(sts *v1.StatefulSet) error{

	client := c.kubeClientset.AppsV1().StatefulSets(c.namespace)

	statefulSet, err := client.Get(sts.Name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if errors.IsNotFound(err) {
		_, err = client.Create(sts)
		if err != nil {
			return err
		}
	} else {
		sts.ResourceVersion = statefulSet.ResourceVersion
		_, err := client.Update(sts)
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

// query API server until the stateful set is completely deployed (use an exponential back off and a timeout)
func (c *Controller) WaitForStatefulSet(sts *v1.StatefulSet) error {
	glog.V(2).Infof("waiting for statefulset %s to be ready", sts.Name)
	return wait.Poll(5*time.Second, 30*time.Second, func() (bool, error) {
		statefulSet,err := c.kubeClientset.AppsV1().StatefulSets(c.namespace).Get(sts.Name, metav1.GetOptions{})
		if err != nil {
			return false,err
		}
		if statefulSet.Status.ReadyReplicas < *sts.Spec.Replicas {
			return false,nil
		}
		return true, nil
	})
}

func (c *Controller) BuildStatefulSet(cc *cassandrav1.CassandraCluster) *v1.StatefulSet{

	limitCPU, _ := resource.ParseQuantity(cc.Spec.Cpu)
	limitMemory, _ := resource.ParseQuantity(cc.Spec.Memory)
	requestCPU, _ := resource.ParseQuantity(cc.Spec.Cpu)
	requestMemory, _ := resource.ParseQuantity(cc.Spec.Memory)
	requestDataStorage,_ := resource.ParseQuantity(cc.Spec.Data.StorageVolume)

	var antiAffinity *corev1.Affinity
	if (cc.Spec.AntiAffinity == true){
		antiAffinity = &corev1.Affinity{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "mongodbCluster",
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{cc.ObjectMeta.Name},
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		}
	}else{
		antiAffinity = nil
	}

	statefulSet := &v1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: cc.Name,
			Labels: map[string]string{
				"cassandraCluster": cc.Name,
				"role": "cassandraCluster",
			},
			Annotations: map[string]string{
				// todo: add operator name in labels
				"operatorVersion": cassandrav1.SchemeGroupVersion.Version,

			},
		},
		Spec: v1.StatefulSetSpec{
			ServiceName: cc.Name,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string {
					"cassandraCluster": cc.Name,
				},
			},
			UpdateStrategy: v1.StatefulSetUpdateStrategy{
				v1.RollingUpdateStatefulSetStrategyType,
				&v1.RollingUpdateStatefulSetStrategy{
					func(i int) *int32 { j:=int32(i);return &j}(0),
				},
			},
			Replicas: cc.Spec.NbNodes,
			//ServiceName: mc.ObjectMeta.Name,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"cassandraCluster": cc.Name,
						"role": "cassandraCluster",
					},
				},
				Spec: corev1.PodSpec{

					Affinity: antiAffinity,
					TerminationGracePeriodSeconds: func(i int64) *int64 { return &i}(10),
					Volumes: []corev1.Volume{
						{
							Name:	"secret",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: cc.Name+"-node-token",
									DefaultMode: func(i int32) *int32 {return &i}(256),
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "cassandra",
							Image:           cc.Spec.BaseImage,
							ImagePullPolicy: "Always",
							Command: []string{
								"numactl",
								"--interleave=all",
								"mongod",
								"--replSet",
								cc.Name,
								"--auth",
								"--clusterAuthMode",
								"keyFile",
								"--keyFile",
								"/etc/secrets-volume/node-token",
								"--setParameter",
								"authenticationMechanisms=SCRAM-SHA-1",
							},
							Env: []corev1.EnvVar{
								{
									Name: "NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "mongo",
									ContainerPort: 27017,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/data/db",
								},
								{
									Name:		"secret",
									MountPath:	"/etc/secrets-volume",
									ReadOnly: 	true,
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									"cpu":    limitCPU,
									"memory": limitMemory,
								},
								Requests: corev1.ResourceList{
									"cpu":    requestCPU,
									"memory": requestMemory,
								},
							},
							// TODO add readiness probe
							//ReadinessProbe:,
							// TODO add liveness probe
							//LivenessProbe:,
							/*livenessProbe:
							tcpSocket:
							port: 27017
							initialDelaySeconds: 30
							timeoutSeconds: 1
						periodSeconds: 10
						successThreshold: 1
							failureThreshold: 3
							readinessProbe:
							exec:
							command:
							- /bin/sh
							- '-i'
                			- '-c'
                			- >-
						mongo $MONGO_URI --eval="quit()"*/
						},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
						Annotations: map[string]string{
							"volume.beta.kubernetes.io/storage-class": cc.Spec.Data.StorageClass,
						},
						Labels: map[string]string{
							"name":      cc.Name,
						},
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: requestDataStorage,
							},
						},
					},
				},
			},
		},
	}
	return statefulSet
}
