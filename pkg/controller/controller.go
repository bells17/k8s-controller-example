package controller

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

// Controller is the controller implementation for Foo resources
type Controller struct {
	kubeclientset     kubernetes.Interface
	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced
	recorder          record.EventRecorder
	workqueue         workqueue.RateLimitingInterface
}

// NewController returns a new sample controller
func NewController(
	ctx context.Context,
	kubeclientset kubernetes.Interface,
	deploymentInformer appsinformers.DeploymentInformer) *Controller {
	logger := klog.FromContext(ctx)

	logger.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "sample-controller"})

	controller := &Controller{
		kubeclientset:     kubeclientset,
		deploymentsLister: deploymentInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "sample-controller"),
		recorder:          recorder,
	}

	logger.Info("Setting up event handlers")
	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(new interface{}) {
			controller.enqueue(new)
		},
		UpdateFunc: func(old, new interface{}) {
			newDepl := new.(*appsv1.Deployment)
			oldDepl := old.(*appsv1.Deployment)
			if newDepl.ResourceVersion == oldDepl.ResourceVersion {
				return
			}
			controller.enqueue(new)
		},
	})

	return controller
}

func (c *Controller) enqueue(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *Controller) Run(ctx context.Context, workers int) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	logger := klog.FromContext(ctx)
	logger.Info("Starting sample-controller")
	defer logger.Info("Shutting down sample-controller")

	if ok := cache.WaitForCacheSync(ctx.Done(), c.deploymentsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	logger.Info("Starting workers", "count", workers)
	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, c.runWorker, time.Second)
	}

	<-ctx.Done()
	return nil
}

func (c *Controller) runWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

func (c *Controller) processNextWorkItem(ctx context.Context) bool {
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	logger := klog.FromContext(ctx)
	defer c.workqueue.Done(obj)

	var key string
	var ok bool
	if key, ok = obj.(string); !ok || key == "" {
		c.workqueue.Forget(obj)
		utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
		return true
	}

	logger = klog.LoggerWithValues(logger, "deployment", key)
	if err := c.syncHandler(klog.NewContext(ctx, logger), key); err != nil {
		logger.Error(err, "Error syncing")
		c.workqueue.AddRateLimited(key)
		utilruntime.HandleError(err)
		return true
	}

	c.workqueue.Forget(obj)
	logger.Info("Successfully synced")
	return true
}

func (c *Controller) syncHandler(ctx context.Context, key string) error {
	logger := klog.FromContext(ctx)
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	deploy, err := c.deploymentsLister.Deployments(ns).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.V(4).Info("Deployment not found")
			return nil
		}
		return fmt.Errorf("failed to get deployment %s", err)
	}

	if deploy.Annotations != nil {
		val, found := deploy.Annotations["sample-controller"]
		if found && val == "True" {
			logger.V(4).Info("sample-controller annotation is already set")
			return nil
		}
	}

	deploy2 := deploy.DeepCopy()
	deploy2.Annotations["sample-controller"] = "True"
	logger.V(4).Info("Need to set sample-controller annotation")
	_, err = c.kubeclientset.AppsV1().Deployments(ns).Update(context.TODO(), deploy2, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	logger.V(4).Info("Successfully set sample-controller annotation to deployment")
	c.recorder.Event(deploy2, corev1.EventTypeNormal, "Synced", "Successfully set sample-controller annotation to deployment")
	return nil
}
