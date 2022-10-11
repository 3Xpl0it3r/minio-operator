package minio

import (
	"fmt"
	"path"

	crapiv1alpha1 "github.com/3Xpl0it3r/minio-operator/pkg/apis/miniooperator.3xpl0it3r.cn/v1alpha1"
	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newPod(podName string, minio *crapiv1alpha1.Minio, nodeName string) *apicorev1.Pod {
	// use fqdn to commuite each other
	var serverEndPoint string
	if minio.Spec.Replicas == 1 {
		serverEndPoint = "/data"
	} else {
		// pod-name-{0..N}.service-name.namespace.svc.cluster.local
		serverEndPoint = fmt.Sprintf("http://%s-{0...%d}.%s.%s.svc.cluster.local/data", getPodNamePrefix(minio), minio.Spec.Replicas-1, getInternalServiceName(minio), minio.GetNamespace())
	}
	var pod = &apicorev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            podName,
			Namespace:       minio.GetNamespace(),
			Labels:          getResourceLabels(minio),
			Annotations:     getResourceAnnotations(minio, nodeName),
			OwnerReferences: getResourceOwnerReference(minio),
		},
		Spec: apicorev1.PodSpec{
			Volumes: []apicorev1.Volume{
				{
					Name: podName,
					VolumeSource: apicorev1.VolumeSource{
						HostPath: &apicorev1.HostPathVolumeSource{
							Path: path.Join(minio.Spec.HostPath, podName),
						},
					},
				},
			},
			Containers: []apicorev1.Container{
				{
					Name:       minio.GetName(),
					Image:      minio.Spec.Image,
					Command:    []string{},
					Args:       []string{"server", "--console-address=0.0.0.0:9001", serverEndPoint},
					WorkingDir: "",
					Ports:      []apicorev1.ContainerPort{},
					Env: []apicorev1.EnvVar{
						{
							Name:  "MINIO_ACCESS_KEY",
							Value: minio.Spec.Credential.AccessKey,
						},
						{
							Name:  "MINIO_SECRET_KEY",
							Value: minio.Spec.Credential.SecretKey,
						},
						{
							Name:  "MINIO_ROOT_USER",
							Value: minio.Spec.Credential.AccessKey,
						},
						{
							Name:  "MINIO_ROOT_PASSWORD",
							Value: minio.Spec.Credential.SecretKey,
						},
					},
					Resources: apicorev1.ResourceRequirements{},
					VolumeMounts: []apicorev1.VolumeMount{
						{
							Name:      podName,
							MountPath: "/data",
						},
					},
					SecurityContext: &apicorev1.SecurityContext{},
					Stdin:           false,
					StdinOnce:       false,
					TTY:             false,
				},
			},
			RestartPolicy: apicorev1.RestartPolicyAlways,
			DNSPolicy:     apicorev1.DNSClusterFirstWithHostNet,
			//pin the pod to a specialfy node
			NodeName:           nodeName,
			Hostname:           podName, //
			Subdomain:          getInternalServiceName(minio),
			ReadinessGates:     []apicorev1.PodReadinessGate{},
			EnableServiceLinks: new(bool),
		},
		Status: apicorev1.PodStatus{},
	}
	return pod
}
