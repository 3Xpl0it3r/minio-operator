package register

import (
	"context"
	"os"
	"syscall"

	extensionapiv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	extensionclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/util/yaml"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RegisterCRDWithFile(namespace string, extClientSet extensionclientset.Interface, filename string) error {
	crd := new(extensionapiv1.CustomResourceDefinition)
	fp, err := os.OpenFile(filename, syscall.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	decoder := yaml.NewYAMLToJSONDecoder(fp)
	if err := decoder.Decode(crd); err != nil {
		return err
	}
	crd.SetNamespace(namespace)
	return RegisterCRDWithObject(extClientSet, crd)
}

// RegisterCRDWithObject register crd
func RegisterCRDWithObject(extClient extensionclientset.Interface, crdObj *extensionapiv1.CustomResourceDefinition) error {
	if _, err := extClient.ApiextensionsV1().CustomResourceDefinitions().Create(context.TODO(), crdObj, metav1.CreateOptions{}); err != nil {
		return err
	}
	return nil
}

func UnregisterCRD(extClientSet extensionclientset.Interface, crdName string) error {
	return extClientSet.ApiextensionsV1().CustomResourceDefinitions().Delete(context.TODO(), crdName, metav1.DeleteOptions{})
}
