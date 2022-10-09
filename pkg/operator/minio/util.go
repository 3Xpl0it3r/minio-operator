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
	"strconv"

	crapiv1alpha1 "github.com/3Xpl0it3r/minio-operator/pkg/apis/miniooperator.3xpl0it3r.cn/v1alpha1"
	crconfig "github.com/3Xpl0it3r/minio-operator/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getMinioAppName(obj *crapiv1alpha1.Minio, target string) string {
	return obj.GetName() + "-" + target
}

// getResourceLabels generate labels according crResource object
func getResourceLabels(obj *crapiv1alpha1.Minio) map[string]string {
	labels := map[string]string{
		crconfig.MinioAppNameLabel: obj.GetName(),
	}
	return labels
}

// getResourceAnnotations generate annotations according crResource object
func getResourceAnnotations(obj *crapiv1alpha1.Minio, nodeName string) map[string]string {
	annotations := map[string]string{
		crconfig.MinioAppNameLabel: obj.GetName(),
	}
    if nodeName != ""{
        annotations[crconfig.MinioAppLocation] = nodeName
    }
	return annotations
}

// getResourceOwnerReference generate OwnerReference according crResource object
func getResourceOwnerReference(obj *crapiv1alpha1.Minio) []metav1.OwnerReference {
	ownerReference := []metav1.OwnerReference{}
	ownerReference = append(ownerReference, *metav1.NewControllerRef(obj, crapiv1alpha1.SchemeGroupVersion.WithKind("Minio")))
	return ownerReference
}


//go:inline
func getPodName(index int, minio *crapiv1alpha1.Minio) string {
	return getPodNamePrefix(minio) +  "-" + strconv.Itoa(index)
}

//go:inline
func getPodNamePrefix( minio *crapiv1alpha1.Minio) string {
	return minio.GetName() 
}
//go:inline
func getInternalServiceName(minio *crapiv1alpha1.Minio) string {
	return minio.GetName() + "-internal"
}
