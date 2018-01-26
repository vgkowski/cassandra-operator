package controller

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/client-go/rest"


	clientset "github.com/vgkowski/cassandra-operator/pkg/client/clientset/versioned"
	listers "github.com/vgkowski/cassandra-operator/pkg/client/listers/cassandra/v1"
	informers "github.com/vgkowski/cassandra-operator/pkg/client/informers/externalversions"
	cassandraScheme "github.com/vgkowski/cassandra-operator/pkg/client/clientset/versioned/scheme"
	cassandrav1 "github.com/vgkowski/cassandra-operator/pkg/apis/cassandra/v1"
	"k8s.io/client-go/scale/scheme/appsv1beta2"
)

const controllerAgentName = "cassandraCluster-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a CassandraCluster is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a CassandraCluster fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by CassandraCluster"
	// MessageResourceSynced is the message used for an Event fired when a CassandraCluster
	// is synced successfully
	MessageResourceSynced = "CassandraCluster synced successfully"
)

// Controller is the controller implementation for CassandraCluster resources
type Controller struct {
	// the configuration used for clients and remoteexec
	config *rest.Config
	// kubeclientset is a standard kubernetes clientset
	kubeClientset kubernetes.Interface
	// namespace where the controller operates
	namespace string
	// cassandraClusterClientset is a clientset for our own API group
	cassandraClusterClientset clientset.Interface

	statefulsetsLister appslisters.StatefulSetLister
	statefulsetsSynced cache.InformerSynced
	CassandraClustersLister        listers.CassandraClusterLister
	CassandraClustersSynced        cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new cassandraCluster controller
func NewController(
	config *rest.Config,
	kubeClientset kubernetes.Interface,
	namespace string,
	cassandraClusterClientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	cassandraClusterInformerFactory informers.SharedInformerFactory) *Controller {

	// obtain references to shared index informers for the Statefulset and CassandraCluster
	// types.
	statefulsetInformer := kubeInformerFactory.Apps().V1().StatefulSets()
	CassandraClusterInformer := cassandraClusterInformerFactory.Cassandra().V1().CassandraClusters()

	// Create event broadcaster
	// Add cassandraCluster-controller types to the default Kubernetes Scheme so Events can be
	// logged for cassandraCluster-controller types.
	cassandraScheme.AddToScheme(scheme.Scheme)
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		config:			   config,
		kubeClientset:     kubeClientset,
		namespace: namespace,
		cassandraClusterClientset:   cassandraClusterClientset,
		statefulsetsLister: statefulsetInformer.Lister(),
		statefulsetsSynced: statefulsetInformer.Informer().HasSynced,
		CassandraClustersLister:        CassandraClusterInformer.Lister(),
		CassandraClustersSynced:        CassandraClusterInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "CassandraClusters"),
		recorder:          recorder,
	}

	glog.Info("Setting up event handlers")
	// Set up an event handler for when CassandraCluster resources change
	CassandraClusterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueCassandraCluster,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueCassandraCluster(new)
		},
		DeleteFunc: controller.enqueueCassandraCluster,
	})
	// Set up an event handler for when Statefulset resources change. This
	// handler will lookup the owner of the given Deployment, and if it is
	// owned by a CassandraCluster resource will enqueue that CassandraCluster resource for
	// processing. This way, we don't need to implement custom logic for
	// handling Statefulset resources. More info on this pattern:
	// https://github.com/kubernetes/community/blob/8cafef897a22026d42f5e5bb3f104febe7e29830/contributors/devel/controllers.md
	statefulsetInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			newSts := new.(*appsv1.StatefulSet)
			oldSts := old.(*appsv1.StatefulSet)
			if newSts.ResourceVersion == oldSts.ResourceVersion {
				// Periodic resync will send update events for all known Deployments.
				// Two different versions of the same Deployment will always have different RVs.
				return
			}
			controller.handleObject(new)
		},
		DeleteFunc: controller.handleObject,
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting CassandraCluster controller")

	// Wait for the caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.statefulsetsSynced, c.CassandraClustersSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")
	// Launch two workers to process CassandraCluster resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// CassandraCluster resource to be synced.
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		glog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the CassandraCluster resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the CassandraCluster resource with this namespace/name
	CassandraCluster, err := c.CassandraClustersLister.CassandraClusters(namespace).Get(name)
	if err != nil {
		// The CassandraCluster resource may no longer exist, in which case we consider as a deletion of a CassandraCluster
		// processing.
		if errors.IsNotFound(err) {
			glog.Infof("CassandraCluster '%s' in work queue no longer exists, deleting the CassandraCluster...", key)
			return c.deleteCassandraCluster(name)
		}

		return err
	} else {
		return c.createOrUpdateCassandraCluster(name)
	}

	statefulsetName := CassandraCluster.Name
	if statefulsetName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		runtime.HandleError(fmt.Errorf("%s: deployment name must be specified", key))
		return nil
	}

	// Get the deployment with the name specified in CassandraCluster.spec
	deployment, err := c.statefulsetsLister.StatefulSets(CassandraCluster.Namespace).Get(name)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		deployment, err = c.kubeClientset.AppsV1().Deployments(CassandraCluster.Namespace).Create(newDeployment(CassandraCluster))
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// If the Deployment is not controlled by this CassandraCluster resource, we should log
	// a warning to the event recorder and ret
	if !metav1.IsControlledBy(deployment, CassandraCluster) {
		msg := fmt.Sprintf(MessageResourceExists, deployment.Name)
		c.recorder.Event(CassandraCluster, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}

	// If this number of the replicas on the CassandraCluster resource is specified, and the
	// number does not equal the current desired replicas on the Deployment, we
	// should update the Deployment resource.
	if CassandraCluster.Spec.Replicas != nil && *CassandraCluster.Spec.Replicas != *deployment.Spec.Replicas {
		glog.V(4).Infof("CassandraClusterr: %d, deplR: %d", *CassandraCluster.Spec.Replicas, *deployment.Spec.Replicas)
		deployment, err = c.kubeclientset.AppsV1beta2().Deployments(CassandraCluster.Namespace).Update(newDeployment(CassandraCluster))
	}

	// If an error occurs during Update, we'll requeue the item so we can
	// attempt processing again later. THis could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// Finally, we update the status block of the CassandraCluster resource to reflect the
	// current state of the world
	err = c.updateCassandraClusterStatus(CassandraCluster, deployment)
	if err != nil {
		return err
	}

	c.recorder.Event(CassandraCluster, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) updateCassandraClusterStatus(CassandraCluster *cassandrav1.CassandraCluster, deployment *appsv1beta2.Deployment) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	CassandraClusterCopy := CassandraCluster.DeepCopy()
	CassandraClusterCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas
	// Until #38113 is merged, we must use Update instead of UpdateStatus to
	// update the Status block of the CassandraCluster resource. UpdateStatus will not
	// allow changes to the Spec of the resource, which is ideal for ensuring
	// nothing other than resource status has been updated.
	_, err := c.sampleclientset.SamplecontrollerV1alpha1().CassandraClusters(CassandraCluster.Namespace).Update(CassandraClusterCopy)
	return err
}

// enqueueCassandraCluster takes a CassandraCluster resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than CassandraCluster.
func (c *Controller) enqueueCassandraCluster(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the CassandraCluster resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that CassandraCluster resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		glog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	glog.V(4).Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a CassandraCluster, we should not do anything more
		// with it.
		if ownerRef.Kind != "CassandraCluster" {
			return
		}

		CassandraCluster, err := c.CassandraClustersLister.CassandraClusters(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			glog.V(4).Infof("ignoring orphaned object '%s' of CassandraCluster '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueueCassandraCluster(CassandraCluster)
		return
	}
}