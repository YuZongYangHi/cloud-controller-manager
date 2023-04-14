package controllers

import (
	"context"
	"errors"
	"fmt"
	"github.com/YuZongYangHi/cloud-controller-manager/pkg/cloudprovider/sdk"
	"github.com/YuZongYangHi/cloud-controller-manager/pkg/config"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"time"
)

const (
	defaultSyncPeriod = 30 * time.Second
)

type LoadBalanceController struct {
	LoadBalanceConfig *config.LoadBalanceConfig

	// addition, deletion, modification and query of load balancing
	LoadBalanceClient *sdk.LoadBalanceClient

	kubeClient kubernetes.Interface

	kubeInformerFactory informers.SharedInformerFactory

	servicesLister v1.ServiceLister
	serviceSynced  cache.InformerSynced
	serviceQueue   workqueue.RateLimitingInterface
}

func NewLoaBalanceController(kubeClient kubernetes.Interface, loadBalanceConfig *config.LoadBalanceConfig) (*LoadBalanceController, error) {
	sharedInformerFactory := informers.NewSharedInformerFactory(kubeClient, defaultSyncPeriod)
	serviceInformer := sharedInformerFactory.Core().V1().Services()

	c := &LoadBalanceController{
		LoadBalanceConfig:   loadBalanceConfig,
		LoadBalanceClient:   sdk.NewLoadBalance(loadBalanceConfig),
		kubeClient:          kubeClient,
		kubeInformerFactory: sharedInformerFactory,
		servicesLister:      serviceInformer.Lister(),
		serviceSynced:       serviceInformer.Informer().HasSynced,
		serviceQueue:        workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
	}

	// wait for cache by lb list
	if !c.LoadBalanceClient.WaitForCacheSync() {
		return nil, errors.New("full sync loadbalances ip list fail")
	}

	serviceInformer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addService,
		UpdateFunc: c.updateService,
		DeleteFunc: c.deleteService,
	}, defaultSyncPeriod)
	return c, nil
}

func (c *LoadBalanceController) Run(workers int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.serviceQueue.ShuttingDown()

	go c.kubeInformerFactory.Start(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.serviceSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	klog.Info("Stopping loadbalance controller")
}

func (c *LoadBalanceController) runWorker() {
	for c.processNextItem() {
	}
}

func (c *LoadBalanceController) processNextItem() bool {
	key, quit := c.serviceQueue.Get()
	if quit {
		return false
	}

	defer c.serviceQueue.Done(key)

	namespace, name, err := cache.SplitMetaNamespaceKey(key.(string))
	if err != nil {
		klog.Errorf("split service namespace error: %s", key)
		return false
	}

	if err = c.syncLoadBalance(namespace, name); err != nil {
		klog.Errorf("sync loadbalance fail name: %s, namespace: %s, err: %s", name, namespace, err.Error())
		return false
	}
	return true
}

func (c *LoadBalanceController) syncLoadBalance(namespace, name string) error {
	service, err := c.servicesLister.Services(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return c.LoadBalanceClient.Unbind(name, namespace)
		}
		return err
	}

	var lb string

	if len(service.Status.LoadBalancer.Ingress) > 0 {
		lb = service.Status.LoadBalancer.Ingress[0].IP
	}

	if service.Spec.LoadBalancerIP != "" {
		lb = service.Spec.LoadBalancerIP
	}

	bindRequest := sdk.LoadBalance{
		Namespace:   namespace,
		ServiceName: name,
	}

	if lb == "" {
		lb, err = c.LoadBalanceClient.GetAvailableIp()
		if err != nil {
			return err
		}
	}

	bindRequest.Ip = lb
	if err = c.LoadBalanceClient.Bind(name, namespace, lb); err != nil {
		return err
	}

	if len(service.Status.LoadBalancer.Ingress) == 0 || service.Status.LoadBalancer.Ingress[0].IP != lb {
		service.Status.LoadBalancer = corev1.LoadBalancerStatus{Ingress: []corev1.LoadBalancerIngress{{IP: lb}}}
		_, err = c.kubeClient.CoreV1().Services(namespace).UpdateStatus(context.Background(), service, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorf("update service %s namespace: %s ip: %s error: %s", name, namespace, lb, err.Error())
			return err
		}
		klog.Infof("ip: %s bound by service: %s, namespace: %s", lb, name, namespace)
	}
	return nil
}

func (c *LoadBalanceController) addService(obj interface{}) {
	service := obj.(*corev1.Service)
	if service.Spec.Type != corev1.ServiceTypeLoadBalancer {
		return
	}
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err == nil {
		c.serviceQueue.Add(key)
	}
}

func (c *LoadBalanceController) updateService(src, dest interface{}) {
	oldService := src.(*corev1.Service)
	newService := dest.(*corev1.Service)

	if oldService.ResourceVersion == newService.ResourceVersion {
		return
	}
	c.addService(newService)
}

func (c *LoadBalanceController) deleteService(obj interface{}) {
	if _, ok := obj.(*corev1.Service); ok {
		c.addService(obj)
		return
	}
	tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
	if !ok {
		klog.Errorf("Couldn't get object from tombstone %#v", obj)
		return
	}

	_, ok = tombstone.Obj.(*corev1.Service)
	if !ok {
		klog.Errorf("Tombstone contained object that is not a Service: %#v", obj)
		return
	}
	c.addService(obj)
	return
}
