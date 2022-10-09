/*
   Copyright 2022 The minio-operator Authors.
   Licensed under the Apache License, PROJECT_VERSION 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at
       http://www.apache.org/licenses/LICENSE-2.0
   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/
package minio

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	apicorev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	listercorev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	crclientset "github.com/3Xpl0it3r/minio-operator/pkg/client/clientset/versioned"
	crinformers "github.com/3Xpl0it3r/minio-operator/pkg/client/informers/externalversions"
	crlisterv1alpha1 "github.com/3Xpl0it3r/minio-operator/pkg/client/listers/miniooperator.3xpl0it3r.cn/v1alpha1"
	crcontroller "github.com/3Xpl0it3r/minio-operator/pkg/controller"
	crhandler "github.com/3Xpl0it3r/minio-operator/pkg/controller/minio/handler"
	croperator "github.com/3Xpl0it3r/minio-operator/pkg/operator"
	miniooperator "github.com/3Xpl0it3r/minio-operator/pkg/operator/minio"
)

// controller is implement Controller for Minio resources
type controller struct {
	crcontroller.Base
	register      prometheus.Registerer
	kubeClientSet kubeclientset.Interface
	crClientSet   crclientset.Interface
	queue         workqueue.RateLimitingInterface
	operator      croperator.Operator
	recorder      record.EventRecorder

	minioLister   crlisterv1alpha1.MinioLister
	serviceLister listercorev1.ServiceLister
	podLister     listercorev1.PodLister
	nodeLister    listercorev1.NodeLister

	cacheSynced []cache.InformerSynced
}

// NewController create a new controller for Minio resources
func NewController(kubeClientSet kubeclientset.Interface, kubeInformers informers.SharedInformerFactory, crClientSet crclientset.Interface,
	crInformers crinformers.SharedInformerFactory, reg prometheus.Registerer) crcontroller.Controller {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.V(2).Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClientSet.CoreV1().Events(apicorev1.NamespaceAll)})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, apicorev1.EventSource{Component: "Minio-operator"})

	return newMinioController(kubeClientSet, kubeInformers, crClientSet, crInformers, recorder, reg)
}

// newMinioController is really
func newMinioController(kubeClientSet kubeclientset.Interface, kubeInformers informers.SharedInformerFactory, crClientSet crclientset.Interface,
	crInformers crinformers.SharedInformerFactory, recorder record.EventRecorder, reg prometheus.Registerer) *controller {
	c := &controller{
		register:      reg,
		kubeClientSet: kubeClientSet,
		crClientSet:   crClientSet,
		recorder:      recorder,
	}
	c.queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	minioInformer := crInformers.Miniooperator().V1alpha1().Minios()
	minioInformer.Informer().AddEventHandlerWithResyncPeriod(crhandler.NewMinioEventHandler(c.enqueueFunc, c.minioLister), 5*time.Second)
	c.minioLister = minioInformer.Lister()
	c.cacheSynced = append(c.cacheSynced, minioInformer.Informer().HasSynced)

	// add
	podInformer := kubeInformers.Core().V1().Pods()
	podInformer.Informer().AddEventHandlerWithResyncPeriod(crhandler.NewPodEventHandler(c.enqueueFunc, c.podLister, c.minioLister), 5*time.Second)
	c.podLister = podInformer.Lister()
	c.cacheSynced = append(c.cacheSynced, podInformer.Informer().HasSynced)

	serviceInformer := kubeInformers.Core().V1().Services()
	serviceInformer.Informer().AddEventHandlerWithResyncPeriod(crhandler.NewServiceEventHandler(c.serviceLister, c.enqueueFunc, c.minioLister), 5*time.Second)
	c.serviceLister = serviceInformer.Lister()
	c.cacheSynced = append(c.cacheSynced, serviceInformer.Informer().HasSynced)

	nodeInformer := kubeInformers.Core().V1().Nodes()
	c.nodeLister = nodeInformer.Lister()
	c.cacheSynced = append(c.cacheSynced, nodeInformer.Informer().HasSynced)

	c.operator = miniooperator.NewOperator(c.kubeClientSet, c.crClientSet, c.podLister, c.serviceLister, c.nodeLister, c.minioLister, c.recorder, c.register)
	return c
}

func (c *controller) Start(worker int, stopCh <-chan struct{}) error {
	// wait for all involved cached to be synced , before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, func() bool {
		for _, hasSyncdFn := range c.cacheSynced {
			if !hasSyncdFn() {
				return false
			}
		}
		return true
	}) {
		return fmt.Errorf("timeout wait for cache to be synced")
	}
	for i := 0; i < worker; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}
	return nil

}

// runWorker for loop
func (c *controller) runWorker() {
	defer utilruntime.HandleCrash()
	for c.processNextItem() {
	}
}

func (c *controller) processNextItem() bool {
	obj, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer func() {
		c.queue.Done(obj)
	}()
	if err := c.operator.Reconcile(obj); err != nil {
		c.queue.AddRateLimited(obj)
		utilruntime.HandleError(fmt.Errorf("failed to sync SparkApplication %q: %v", obj, err))
	}
	c.queue.Forget(obj)
	return true
}

func (c *controller) Stop() {
	klog.Info("Stopping the minio operator controller")
	c.queue.ShutDown()
}

func (c *controller) enqueueFunc(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		klog.Errorf("failed to get key for %v: %v", obj, err)
		return
	}
	c.queue.AddRateLimited(key)
}
