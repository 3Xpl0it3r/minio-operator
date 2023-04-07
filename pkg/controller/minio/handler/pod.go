package handler

import (
	apicorev1 "k8s.io/api/core/v1"
	listercorev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	crlisterv1alpha1 "github.com/3Xpl0it3r/minio-operator/pkg/client/listers/miniooperator.3xpl0it3r.cn/v1alpha1"
	crconfig "github.com/3Xpl0it3r/minio-operator/pkg/config"
)

// podEventHandler represent podeventhandler
type podEventHandler struct {
	podLister   listercorev1.PodLister
	enqueueFn   func(key interface{})
	minioLister crlisterv1alpha1.MinioLister
}

func NewPodEventHandler(enqueueFn func(key interface{}), podListener listercorev1.PodLister, minioLister crlisterv1alpha1.MinioLister) *podEventHandler {
	return &podEventHandler{
		podLister:   podListener,
		enqueueFn:   enqueueFn,
		minioLister: minioLister,
	}
}

// podEventHandler represent podeventhandler
func (podeventhandler *podEventHandler) OnAdd(obj interface{}) {
	pod, ok := obj.(*apicorev1.Pod)
	if !ok {
		return
	}
	podeventhandler.enqueueMinioForPodUpdate(pod)
}

// podEventHandler represent podeventhandler
func (podeventhandler *podEventHandler) OnUpdate(oldObj, newObj interface{}) {
	oldPod, ok := oldObj.(*apicorev1.Pod)
	if !ok {
		return
	}
	newPod, ok := newObj.(*apicorev1.Pod)
	if !ok {
		return
	}
	if newPod.ResourceVersion == oldPod.ResourceVersion {
		return
	}
	podeventhandler.enqueueMinioForPodUpdate(newPod)
}

// podEventHandler represent podeventhandler
func (podeventhandler *podEventHandler) OnDelete(obj interface{}) {
	var deletedPod *apicorev1.Pod
	switch obj.(type) {
	case *apicorev1.Pod:
		deletedPod = obj.(*apicorev1.Pod)
	case cache.DeletedFinalStateUnknown:
		deletedObj := obj.(cache.DeletedFinalStateUnknown).Obj
		deletedPod = deletedObj.(*apicorev1.Pod)
	default:
		return
	}
	podeventhandler.enqueueMinioForPodUpdate(deletedPod)
}

// podEventHandler represent podeventhandler
func (podeventhandler *podEventHandler) enqueueMinioForPodUpdate(pod *apicorev1.Pod) {
	appName, ok := pod.Labels[crconfig.MinioAppNameLabel]
	if !ok {
		return
	}
	minio, err := podeventhandler.minioLister.Minios(pod.GetNamespace()).Get(appName)
	if err != nil {
		return
	}
	podeventhandler.enqueueFn(minio)
}
