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

package handler

import (
	crapiv1alpha1 "github.com/3Xpl0it3r/minio-operator/pkg/apis/miniooperator.3xpl0it3r.cn/v1alpha1"
	crlisterv1alpha1 "github.com/3Xpl0it3r/minio-operator/pkg/client/listers/miniooperator.3xpl0it3r.cn/v1alpha1"
)


type minioEventHandler struct {
	minioLister crlisterv1alpha1.MinioLister
	enqueueFn   func(key interface{})
}

func (h *minioEventHandler) OnAdd(obj interface{}) {
	if minio, ok := obj.(*crapiv1alpha1.Minio); ok {
		h.enqueueFn(minio)
	}
}

func (h *minioEventHandler) OnUpdate(oldObj, newObj interface{}) {
    oldMinio,ok := oldObj.(*crapiv1alpha1.Minio)
    if !ok {
        return
    }
    newMinio, ok := newObj.(*crapiv1alpha1.Minio)
    if !ok {
        return
    }
    if oldMinio.ResourceVersion == newMinio.ResourceVersion {
        h.enqueueFn(newMinio)
    }
}

func (h *minioEventHandler) OnDelete(obj interface{}) {
    if minio,ok := obj.(*crapiv1alpha1.Minio); ok  {
        h.enqueueFn(minio)
    }
}

func NewMinioEventHandler(enqueueFn func(key interface{}), lister crlisterv1alpha1.MinioLister) *minioEventHandler {
	return &minioEventHandler{
		minioLister: lister,
		enqueueFn:   enqueueFn,
	}
}
