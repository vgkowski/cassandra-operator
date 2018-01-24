package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Sirupsen/logrus"
	//"github.com/vgkowski/cassandra-operator/pkg/operator"
)

var (
	appVersion = "0.1"

	printVersion bool
	baseImage    string
	kubeCfgFile  string
	namespace	 string
	masterURL	 string
)

func init() {
	flag.StringVar(&baseImage, "baseImage", "sebmoule/cassandra:0.1-dev", "Base image to use when spinning up the Cassandra components.")
	flag.StringVar(&kubeCfgFile, "kubecfg-file", "", "Location of kubecfg file for access to kubernetes master service; --kube_master_url overrides the URL part of this; if neither this nor --kube_master_url are provided, defaults to service account tokens")
	flag.StringVar(&namespace, "namespace", os.Getenv("NAMESPACE"), "namespace to deploy the operator")
	flag.StringVar(&masterURL, "masterURL", "http://127.0.0.1:8001", "Full url to k8s api server")
	flag.Parse()
}

// Main entrypoint
func main()  {
	// Some possible improvements with parallelism and asynchronous to manage multiple clusters
	// in parallel
	if printVersion {
		fmt.Println("cassandra-operator", appVersion)
		os.Exit(0)
	}

	logrus.Info("Cassandra operator starting up!")

	// Print params configured
	logrus.Info("Using Variables:")
	logrus.Infof("   baseImage: %s", baseImage)
	logrus.Infof("   namespace: %s", namespace)

	sigs := make(chan os.Signal, 1) // Create channel to receive OS signals
	stop := make(chan struct{})     // Create channel to receive stop signal

	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGINT) // Register the sigs channel to receieve SIGTERM

	wg := &sync.WaitGroup{} // Goroutines can add themselves to this to be waited on so that they finish


	// Init
	controller, err := operator.NewCassandraClusterController(kubeCfgFile,namespace, baseImage)

	if err != nil {
		logrus.Error("Could not init Controller! ", err)
		panic(err.Error())
	}

	// Kick it off
	controller.Run(stop,wg)


	<-sigs // Wait for signals (this hangs until a signal arrives)
	logrus.Info("Shutting down...")

	close(stop) // Tell goroutines to stop themselves
	wg.Wait()   // Wait for all to be stopped
}