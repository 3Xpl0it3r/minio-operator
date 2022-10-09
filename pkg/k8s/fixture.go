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
package k8s

import (
	"fmt"
	apiappsv1 "k8s.io/api/apps/v1"
	apicorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	k8stest "k8s.io/client-go/testing"

	// crapiv1alpha1 "github.com/3Xpl0it3r/minio-operator/pkg/apis/miniooperator.3xpl0it3r.cn/v1alpha1"
	crfakeclients "github.com/3Xpl0it3r/minio-operator/pkg/client/clientset/versioned/fake"
	crinformers "github.com/3Xpl0it3r/minio-operator/pkg/client/informers/externalversions"
	"github.com/3Xpl0it3r/minio-operator/pkg/operator"
)

type Fixture struct {
	kubeClient *k8sfake.Clientset
	crClient   *crfakeclients.Clientset

	// Object to put in the store
	podLister        []*apicorev1.Pod
	deploymentLister []*apiappsv1.Deployment
	statefulSet      []*apiappsv1.StatefulSet
	daemonSet        []*apiappsv1.DaemonSet
	serviceLister    []*apicorev1.Service
	configMapLister  []*apicorev1.ConfigMap
	// for custom resource
	customListers []*runtime.Object

	// todo write your code here
	kubeInformers informers.SharedInformerFactory
	crInformers   crinformers.SharedInformerFactory

	// Actions expected to happen on client.
	kubeActions           []k8stest.Action
	customResourceActions []k8stest.Action

	// Objects from here preloaded into NewSimpleFake
	kubeObjects           []runtime.Object
	customResourceObjects []runtime.Object

	// operator
	operator operator.Operator
}

func NewFixture(k8sFakeClient *k8sfake.Clientset, crFakeClient *crfakeclients.Clientset,
	kubeInformers informers.SharedInformerFactory, crInformers crinformers.SharedInformerFactory) *Fixture {
	return &Fixture{kubeClient: k8sFakeClient, crClient: crFakeClient, kubeInformers: kubeInformers, crInformers: crInformers}
}

func (f *Fixture) AddPodLister(pods ...*apicorev1.Pod) error {
	for _, pod := range pods {
		if err := f.kubeInformers.Core().V1().Pods().Informer().GetIndexer().Add(pod); err != nil {
			return err
		}
	}
	return nil
}

func (f *Fixture) AddDeploymentLister(dpls ...*apiappsv1.Deployment) error {
	for _, dpl := range dpls {
		if err := f.kubeInformers.Apps().V1().Deployments().Informer().GetIndexer().Add(dpl); err != nil {
			return err
		}
	}
	return nil
}

func (f *Fixture) AddStatefulSetLister(sts ...*apiappsv1.StatefulSet) error {
	for _, st := range sts {
		if err := f.kubeInformers.Apps().V1().StatefulSets().Informer().GetIndexer().Add(st); err != nil {
			return err
		}
	}
	return nil
}

func (f *Fixture) AddDaemonSetLister(dss ...*apiappsv1.DaemonSet) error {
	for _, ds := range dss {
		if err := f.kubeInformers.Apps().V1().DaemonSets().Informer().GetIndexer().Add(ds); err != nil {
			return err
		}
	}
	return nil
}

func (f *Fixture) AddServiceLister(svs ...*apicorev1.Service) error {
	for _, sv := range svs {
		if err := f.kubeInformers.Core().V1().Services().Informer().GetIndexer().Add(sv); err != nil {
			return err
		}
	}
	return nil
}

func (f *Fixture) AddConfigMapLister(cms ...*apicorev1.ConfigMap) error {
	for _, cm := range cms {
		if err := f.kubeInformers.Core().V1().ConfigMaps().Informer().GetIndexer().Add(cm); err != nil {
			return err
		}
	}
	return nil
}

func (f *Fixture) AddCustomResourceLister(cr runtime.Object) error {
	f.customResourceObjects = append(f.customResourceObjects, cr)
	switch cr.GetObjectKind().GroupVersionKind().Kind {
	default:
		return fmt.Errorf("Unexpect Custom Resource Type %s %s ", cr.GetObjectKind().GroupVersionKind().Kind, cr.GetObjectKind().GroupVersionKind().GroupVersion())
	}
	return nil
}

// add expect actions
func (f *Fixture) PutKubeActions(kubeActions ...k8stest.Action) {
	f.kubeActions = append(f.kubeActions, kubeActions...)
}

func (f *Fixture) PutCustomResourceActions(crActions ...k8stest.Action) {
	f.customResourceActions = append(f.customResourceActions, crActions...)
}

func (f *Fixture) GetKubeActions() []k8stest.Action {
	return f.kubeActions
}

func (f *Fixture) GetCustomResourceActions() []k8stest.Action {
	return f.customResourceActions
}
