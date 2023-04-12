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
	"k8s.io/klog/v2"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

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
	var minioCopy *crapiv1alpha1.Minio

	if minio, err := o.minioLister.Minios(namespace).Get(name); err != nil {
		if k8serror.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("%s/%s get minio failed %v", namespace, name, err)
	} else {
		minioCopy = minio.DeepCopy()
	}

	// defaulter
	crapiv1alpha1.MinioDefaulter(minioCopy)

	var nodes []string
	// sync all nodes, this step is used to list all nodes , and upate minio. replicas according the number of nodes
	// if only has one node in kubernetes cluster, then we set replicas of minio to 1
	if err = o.syncNodes(minioCopy, &nodes); err != nil {
		return fmt.Errorf("%s/%s sync node failed, err %v", namespace, name, err)
	}

	// sync Service
	if _, err = o.syncInternalService(minioCopy); err != nil {
		return fmt.Errorf("%s/%s sync service failed %s", namespace, name, err)
	}
	if _, err = o.syncExternalService(minioCopy); err != nil {
		return fmt.Errorf("%s/%s sync service failed %s", namespace, name, err)
	}
	// sync pods
	var shouldUpdate bool
	_, shouldUpdate, err = o.syncPods(minioCopy, nodes)
	if shouldUpdate {
		preErr := err
		if minioCopy, err = o.minioClient.MiniooperatorV1alpha1().Minios(namespace).Update(context.TODO(), minioCopy, metav1.UpdateOptions{}); err != nil {
			err = errors.Wrapf(preErr, "update minio'annno failed: %v", err)
		}
	}
	if err != nil {
		return err
	}
	if err = o.syncMinioApplication(minioCopy, 60*time.Second); err != nil {
		return fmt.Errorf("Sync minio application failed: %v", err)
	}
	minioCopy.Status.Inited = "Ok"
	if _, err = o.minioClient.MiniooperatorV1alpha1().Minios(namespace).UpdateStatus(context.TODO(), minioCopy, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("Update minio apps status failed %v", err)
	}

	return nil

}

// operator represent operator
func (o *operator) syncNodes(minio *crapiv1alpha1.Minio, nodes *[]string) error {
	nodeListItems, err := o.nodeLister.List(labels.Everything())
	if err != nil {
		return err
	}
	for _, nodeItem := range nodeListItems {
		for _, cond := range nodeItem.Status.Conditions {
			switch cond.Type {
			case apicorev1.NodeReady:
				if cond.Status == apicorev1.ConditionTrue {
					*nodes = append(*nodes, nodeItem.GetName())
				}
			}
		}
	}
	if len(*nodes) == 1 {
		minio.Spec.Replicas = 1
	}
	return nil
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
		pickedNode := nodeNameForSchedulePod(podName, minio, nodeResPoll)
		pod, err := o.kubeClientSet.CoreV1().Pods(minio.GetNamespace()).Create(context.TODO(), newPod(podName, minio, pickedNode), metav1.CreateOptions{})
		if err != nil {
			return nil, crIsUpdate, err
		}
		// here means schedule is validate
		crIsUpdate = true
		minio.Annotations[podName] = pickedNode
		if err := o.waitForPodReady(pod, 30*time.Second); err != nil {
			return nil, crIsUpdate, err
		}
	}
	return nil, crIsUpdate, nil
}

// operator represent operator
func (o *operator) syncInternalService(minio *crapiv1alpha1.Minio) (*apicorev1.Service, error) {
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
	svc, err = o.kubeClientSet.CoreV1().Services(minio.GetNamespace()).Create(context.TODO(), newInternalService(minio), metav1.CreateOptions{})
	return svc, err
}

// operator represent operator
func (o *operator) syncExternalService(minio *crapiv1alpha1.Minio) (*apicorev1.Service, error) {
	// if service is existed, then return nil
	svc, err := o.serviceLister.Services(minio.GetNamespace()).Get(getExternalServiceName(minio))
	if err == nil {
		return svc, nil
	}
	// get service failed, buf not because sevice is not existed, for some other reasone
	if !k8serror.IsNotFound(err) {
		return nil, err
	}
	// service is not existed, then create new one
	svc, err = o.kubeClientSet.CoreV1().Services(minio.GetNamespace()).Create(context.TODO(), newExternalService(minio), metav1.CreateOptions{})
	return svc, err
}

// wait for all pod ready
func (o *operator) waitForPodReady(pod *apicorev1.Pod, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()

	var (
		latestPod *apicorev1.Pod
		err       error
	)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(1 * time.Second)
		}

		if latestPod, err = o.podLister.Pods(pod.GetNamespace()).Get(pod.GetName()); err != nil {
			continue
		}

		for _, condition := range latestPod.Status.Conditions {
			switch condition.Type {
			case apicorev1.PodReady:
				if condition.Status == apicorev1.ConditionTrue {
					return nil
				}
			}
		}
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

// operator represent operator
func (o *operator) syncMinioApplication(minioobject *crapiv1alpha1.Minio, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()

	var (
		minioClient *minio.Client
		err         error
	)

	var (
		endpoint  = fmt.Sprintf("%s.%s:9000", getExternalServiceName(minioobject), minioobject.GetNamespace())
		createOpt = minio.MakeBucketOptions{Region: "cn-north-1", ObjectLocking: true}
	)

	// for not in erasure codeed mode, ObjectLocking feature is not supported
	if minioobject.Spec.Replicas == 1 {
		createOpt.ObjectLocking = false
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(10 * time.Second)
		}

		if minioClient, err = minio.New(
			endpoint,
			&minio.Options{
				Creds:  credentials.NewStaticV4(minioobject.Spec.Credential.AccessKey, minioobject.Spec.Credential.SecretKey, ""),
				Secure: false},
		); err != nil {
			continue
		}

		if !minioClient.IsOnline() {
			continue
		}
		// 由于minio没有提供检测server是不是已经初始化完的api的, 因此这里通过创建一个testbucket来确认minio是不是已经初始化完成了
		_, err = minioClient.GetBucketLocation(context.Background(), "testbucket")
		if err != nil {
			if err := minioClient.MakeBucket(context.TODO(), "testbucket", createOpt); err != nil {
				klog.Errorf("get || create testbucket failed: %v", err)
				continue
			}
		}
		// if detacted minio is online , we should this is ok, event if create bucket failed
		if strings.Compare(minioobject.Status.Inited, "Ok") != 0 {
			for _, bucketName := range minioobject.Spec.Buckets {
				if err := minioClient.MakeBucket(context.TODO(), bucketName, createOpt); err != nil {
					klog.Errorf("create minio bucket failed %v", err)
				}
			}
		}
	}
}
