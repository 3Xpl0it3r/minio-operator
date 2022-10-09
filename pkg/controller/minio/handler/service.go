package handler

import (
	crlisterv1alpha1 "github.com/3Xpl0it3r/minio-operator/pkg/client/listers/miniooperator.3xpl0it3r.cn/v1alpha1"
	crconfig "github.com/3Xpl0it3r/minio-operator/pkg/config"
	apicorev1 "k8s.io/api/core/v1"
	listenercorev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

// serviceEventHandler represent serviceeventhandler
type serviceEventHandler struct {
	serviceListener listenercorev1.ServiceLister
	enqueneFn       func(obj interface{})
	minioLister     crlisterv1alpha1.MinioLister
}

func NewServiceEventHandler(serviceListener listenercorev1.ServiceLister, enqueueFn func(obj interface{}), minioLister crlisterv1alpha1.MinioLister) *serviceEventHandler {
	return &serviceEventHandler{
		serviceListener: serviceListener,
		enqueneFn:       enqueueFn,
		minioLister:     minioLister,
	}
}

// serviceEventHandler represent serviceeventhandler
func (serviceeventhandler *serviceEventHandler) OnAdd(obj interface{}) {
	svc, ok := obj.(*apicorev1.Service)
	if !ok {
		return
	}
	serviceeventhandler.enqueueMinioForServiceUpdate(svc)
}

// serviceEventHandler represent serviceeventhandler
func (serviceeventhandler *serviceEventHandler) OnDelete(obj interface{}) {
	var deletedSvc *apicorev1.Service
	switch obj.(type) {
	case *apicorev1.Service:
		deletedSvc = obj.(*apicorev1.Service)
	case cache.DeletedFinalStateUnknown:
		deletedObj := obj.(cache.DeletedFinalStateUnknown).Obj
		deletedSvc = deletedObj.(*apicorev1.Service)
	default:
		return
	}
	serviceeventhandler.enqueueMinioForServiceUpdate(deletedSvc)
}

// serviceEventHandler represent serviceeventhandler
func (serviceeventhandler *serviceEventHandler) OnUpdate(oldObj, newObj interface{}) {
	oldSvc, ok := oldObj.(*apicorev1.Service)
	if !ok {
		return
	}
	newSvc, ok := newObj.(*apicorev1.Service)
	if !ok {
		return
	}
	if oldSvc.ResourceVersion == newSvc.ResourceVersion {
		return
	}
	serviceeventhandler.enqueueMinioForServiceUpdate(newSvc)
}

// serviceEventHandler represent serviceeventhandler
func (serviceeventhandler *serviceEventHandler) enqueueMinioForServiceUpdate(svc *apicorev1.Service) {
	appName, ok := svc.Labels[crconfig.MinioAppNameLabel]
	if !ok {
		return
	}
	app, err := serviceeventhandler.minioLister.Minios(svc.GetNamespace()).Get(appName)
	if err != nil {
		return
	}
	serviceeventhandler.enqueneFn(app)
}
