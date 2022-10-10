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
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	apicorev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"

	crapiv1alpha1 "github.com/3Xpl0it3r/minio-operator/pkg/apis/miniooperator.3xpl0it3r.cn/v1alpha1"
	crclientset "github.com/3Xpl0it3r/minio-operator/pkg/client/clientset/versioned"
	crlisterv1alpha1 "github.com/3Xpl0it3r/minio-operator/pkg/client/listers/miniooperator.3xpl0it3r.cn/v1alpha1"
	crconfig "github.com/3Xpl0it3r/minio-operator/pkg/config"
	croperator "github.com/3Xpl0it3r/minio-operator/pkg/operator"
	listercorev1 "k8s.io/client-go/listers/core/v1"
)

type operator struct {
	minioClient   crclientset.Interface
	kubeClientSet kubernetes.Interface
	recorder      record.EventRecorder
	minioLister   crlisterv1alpha1.MinioLister
	reg           prometheus.Registerer
	serviceLister listercorev1.ServiceLister
	podLister     listercorev1.PodLister
	nodeLister    listercorev1.NodeLister
}

func NewOperator(kubeClientSet kubernetes.Interface, crClientSet crclientset.Interface, podLister listercorev1.PodLister, serviceLister listercorev1.ServiceLister, nodeLister listercorev1.NodeLister, minioLister crlisterv1alpha1.MinioLister, recorder record.EventRecorder, reg prometheus.Registerer) croperator.Operator {
	return &operator{
		minioClient:   crClientSet,
		minioLister:   minioLister,
		reg:           reg,
		kubeClientSet: kubeClientSet,

		podLister:     podLister,
		serviceLister: serviceLister,
		nodeLister:    nodeLister,
	}
}

func (o *operator) Reconcile(object interface{}) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(object.(string))
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("failed to get the namespace and name from key: %v : %v", object, err))
		return nil
	}

	minio, err := o.minioLister.Minios(namespace).Get(name)
	if err != nil {
		if k8serror.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("%s/%s get minio failed %v", namespace, name, err)
	}
	// defaulter
	crapiv1alpha1.MinioDefaulter(minio)
	nodeListItems, err := o.nodeLister.List(labels.Everything())
	if err != nil {
		return fmt.Errorf("%s/%s list all node failed, %s", namespace, name, err)
	}
	nodes := []string{}
	for _, nodeItem := range nodeListItems {
		for _, cond := range nodeItem.Status.Conditions {
			switch cond.Type {
			case apicorev1.NodeReady:
				if cond.Status == apicorev1.ConditionTrue {
					nodes = append(nodes, nodeItem.GetName())
				}
			}
		}
	}
	// if only has one node in kubernetes cluster, then we set replicas of minio to 1
	if len(nodes) == 1 {
		minio.Spec.Replicas = 1
	}
	// sync Service
	if _, err := o.syncService(minio); err != nil {
		return fmt.Errorf("%s/%s sync service failed %s", namespace, name, err)
	}
	// sync pods
	_, shoudUpdate, err := o.syncPods(minio, nodes)
	if shoudUpdate {
		if _, e := o.minioClient.MiniooperatorV1alpha1().Minios(namespace).Update(context.TODO(), minio, metav1.UpdateOptions{}); e!= nil {
            return errors.Wrapf(err, "update minio'anno failed: %v", e)
        }
	}
    return err
}

// operator represent operator
func (o *operator) syncPods(minio *crapiv1alpha1.Minio, allNodes []string) ([]*apicorev1.Pod, bool, error) {
	nodeResPoll := make(map[int][]string, 64) //
	nodeResPoll[0] = allNodes
	podShoudCreate := []string{}
	crIsUpdate := false
	for index := 0; index < int(minio.Spec.Replicas); index++ {
		pod, err := o.podLister.Pods(minio.GetNamespace()).Get(getPodName(index, minio))
		if err != nil {
			if !k8serror.IsNotFound(err) {
				return nil, crIsUpdate, err
			}
			podShoudCreate = append(podShoudCreate, getPodName(index, minio))
			continue
		}
		// if pod existed ,then update nodeinfo
		nodeName, _ := pod.GetAnnotations()[crconfig.MinioAppLocation]
		updateNodeAllocatedInfo(nodeResPoll, nodeName)
	}
	// create some pods if necessary
	for _, podName := range podShoudCreate {
		crIsUpdate = true
		pickedNode := nodeNameForSchedulePod(podName, minio, nodeResPoll)
		pod, err := o.kubeClientSet.CoreV1().Pods(minio.GetNamespace()).Create(context.TODO(), newPod(podName, minio, pickedNode), metav1.CreateOptions{})
		if err != nil {
			return nil, crIsUpdate, err
		}
		if err := o.waitForPodReady(pod, 30*time.Second); err != nil {
			return nil, crIsUpdate, err
		}
		minio.Annotations[podName] = pickedNode
	}
	return nil, crIsUpdate, nil
}

// operator represent operator
func (o *operator) syncService(minio *crapiv1alpha1.Minio) (*apicorev1.Service, error) {
	// if service is existed, then return nil
	svc, err := o.serviceLister.Services(minio.GetNamespace()).Get(getInternalServiceName(minio))
	if err == nil {
		return svc, nil
	}
	// get service failed, buf not because sevice is not existed, for some other reasone
	if !k8serror.IsNotFound(err) {
		return nil, err
	}
	// service is not existed, then create new one
	svc, err = o.kubeClientSet.CoreV1().Services(minio.GetNamespace()).Create(context.TODO(), newService(minio), metav1.CreateOptions{})

	return svc, err
}

// operator represent operator
func (o *operator) getAllReadyPodsAndRemoveUnReady(pods []*apicorev1.Pod) ([]*apicorev1.Pod, error) {
	// if pods is not existed ,then return nil , error is set to nil for we should create some pods
	if len(pods) == 0 {
		return nil, nil
	}
	readyPods := []*apicorev1.Pod{}
	unreadyPods := []*apicorev1.Pod{}
	// get all Ready Pods
	for _, pod := range pods {
		for _, condition := range pod.Status.Conditions {
			switch condition.Type {
			case apicorev1.PodReady:
				if condition.Status == apicorev1.ConditionTrue {
					readyPods = append(readyPods, pod)
				} else {
					unreadyPods = append(unreadyPods, pod)
				}
			}
		}
	}
	// delete all unready pods
	for _, pod := range unreadyPods {
		if err := o.kubeClientSet.CoreV1().Pods(pod.GetNamespace()).Delete(context.TODO(), pod.GetName(), metav1.DeleteOptions{}); err != nil {
			return readyPods, err
		}
	}

	return readyPods, nil
}

// wait for all pod ready
func (o *operator) waitForPodReady(pod *apicorev1.Pod, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()
	ready := make(chan struct{}, 0)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			latestPod, err := o.podLister.Pods(pod.GetNamespace()).Get(pod.GetName())
			if err != nil {
				continue
			}
			for _, condition := range latestPod.Status.Conditions {
				switch condition.Type {
				case apicorev1.PodReady:
					if condition.Status == apicorev1.ConditionTrue {
						ready <- struct{}{}
					}
				}
			}
		}
	}()
	select {
	case <-ctx.Done():
		return fmt.Errorf("Timeout wait for pod %s ready", pod.GetName())
	case <-ready:
		return nil
	}
}

func nodeNameForSchedulePod(podName string, minio *crapiv1alpha1.Minio, nodes map[int][]string) string {
	// if pod has be created, then it has be deleted for some unexpected reason, so in this case ,the pod and node map has been write into the status of minio
	// we should get restaore this
	picked, ok := minio.GetAnnotations()[podName]
	if ok {
		updateNodeAllocatedInfo(nodes, picked)
		return picked
	}
	// new pod, has not create before
	for level := 0; level < len(nodes); level++ {
		resSize := len(nodes[level])
		if resSize == 0 {
			continue
		}
		picked := nodes[level][resSize-1]
		nodes[level+1] = append(nodes[level+1], picked)
		if resSize == 1 {
			nodes[level] = nodes[level][:0]
		} else {
			nodes[level] = nodes[level][:resSize-1]
		}
		return picked
	}
	return ""
}

func updateNodeAllocatedInfo(nodes map[int][]string, node string) {
	for level, nodesList := range nodes {
		for index, name := range nodesList {
			if strings.Compare(node, name) == 0 {
				nodes[level] = append(nodes[level][:index], nodes[index+1]...)
				if level+1 > len(nodes) {
					panic(fmt.Sprintf("too many pods %v %v", level, len(nodes)))
				}
				nodes[level+1] = append(nodes[level+1], name)
				return
			}
		}
	}
}
