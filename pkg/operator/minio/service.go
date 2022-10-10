package minio

import (
	crapiv1alpha1 "github.com/3Xpl0it3r/minio-operator/pkg/apis/miniooperator.3xpl0it3r.cn/v1alpha1"
	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func newService(minio *crapiv1alpha1.Minio) *apicorev1.Service {
	svc := &apicorev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            getInternalServiceName(minio),
			Namespace:       minio.GetNamespace(),
			Labels:          getResourceLabels(minio),
			Annotations:     getResourceAnnotations(minio, ""),
			OwnerReferences: getResourceOwnerReference(minio),
		},
		Spec: apicorev1.ServiceSpec{
			Ports: []apicorev1.ServicePort{
				{
					Name:       "api",
					Port:       9000,
					TargetPort: intstr.IntOrString{IntVal: 9000},
				},
				{
					Name:       "http",
					Port:       9001,
					TargetPort: intstr.IntOrString{IntVal: 9001},
				},
			},
			Selector:   getResourceLabels(minio),
			ClusterIPs: []string{},
			ClusterIP:  "None",
			Type:       apicorev1.ServiceTypeClusterIP,
		},
		Status: apicorev1.ServiceStatus{},
	}
	return svc
}
