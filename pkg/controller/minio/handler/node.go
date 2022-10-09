package handler

import (
	crlisterv1alpha1 "github.com/3Xpl0it3r/minio-operator/pkg/client/listers/miniooperator.3xpl0it3r.cn/v1alpha1"
	listercorev1 "k8s.io/client-go/listers/core/v1"
)

// nodeEventHandler represent nodeeventhandler
type nodeEventHandler struct {
	nodeLister  listercorev1.NodeLister
	minioLister crlisterv1alpha1.MinioLister
	enqueueFn   func(key interface{})
}

// nodeEventHandler represent nodeeventhandler
func (nodeeventhandler *nodeEventHandler) OnAdd(obj interface{}) {
    return
}

// nodeEventHandler represent nodeeventhandler
func (nodeeventhandler *nodeEventHandler) OnDelete(obj interface{}) {
}

// nodeEventHandler represent nodeeventhandler
func (nodeeventhandler *nodeEventHandler) OnUpdate(oldObj, newObj interface{}) {
}

// enqueue minio  if only node in the following case 
// when node delete ,then we should enqueue minio for do some check
// when node update for draint, then we should queue minio for do some check
// here we should list all mino on the current nodes, then enqueue all of them
func (nodeeventhandler *nodeEventHandler) enqueueMinioForNodeUpdate(obj interface{}) {
}
