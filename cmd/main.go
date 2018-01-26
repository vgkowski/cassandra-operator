package main

import (
	"flag"
	"time"

	"github.com/golang/glog"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/vgkowski/cassandra-operator/pkg/signals"
	"os"

	clientset "github.com/vgkowski/cassandra-operator/pkg/client/clientset/versioned"
	informers "github.com/vgkowski/cassandra-operator/pkg/client/informers/externalversions"
	cassandraController "github.com/vgkowski/cassandra-operator/pkg/controller"
)

var (
	masterURL  string
	kubeconfig string
	baseImage string
	namespace string
)

func main() {
	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	cassandraClusterClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building example clientset: %s", err.Error())
	}

	// use informers that filter on the namespace where the operator is deployed
	kubeInformerFactory := kubeinformers.NewFilteredSharedInformerFactory(kubeClient, time.Second*30,namespace,nil)
	cassandraClusterInformerFactory := informers.NewFilteredSharedInformerFactory(cassandraClusterClient, time.Second*30,namespace,nil)

	controller := cassandraController.NewController(cfg,kubeClient,namespace, cassandraClusterClient, kubeInformerFactory, cassandraClusterInformerFactory)

	go kubeInformerFactory.Start(stopCh)
	go cassandraClusterInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&baseImage, "baseImage", "cassandra:3.0.15", "Base image to use when spinning up the Cassandra components.")
	flag.StringVar(&namespace, "namespace", os.Getenv("NAMESPACE"), "namespace to deploy the controller")
}
