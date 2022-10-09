package register

import (
	"context"
	"errors"
	"time"

	extensionapiv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	extensionclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

func WaitForCRDEstablished(extClientSet extensionclientset.Interface, crdName string) error {
	return wait.Poll(1250*time.Millisecond, 10*time.Second, func() (done bool, err error) {
		crd, err := extClientSet.ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), crdName, metav1.GetOptions{})
		klog.Infof("crd info : %v\n %v\n", crd.GetName(), crd.GetNamespace())
		if err != nil {
			return false, err
		}
		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case extensionapiv1.NamesAccepted:
				if cond.Status == extensionapiv1.ConditionFalse {
					return false, errors.New("CRD Name Conflict")
				}
			case extensionapiv1.Established:
				if cond.Status == extensionapiv1.ConditionTrue {
					return true, nil
				}
			}
		}
		return false, err
	})
}
