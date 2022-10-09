package crd

import (
	"github.com/3Xpl0it3r/minio-operator/pkg/crd/minio"
	"github.com/3Xpl0it3r/minio-operator/pkg/crd/register"

	extensionapiv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	extensionclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
)

func InstallCustomResourceDefineToApiServer(extClientSet extensionclientset.Interface) error {
	crdResourceList := []*extensionapiv1.CustomResourceDefinition{}
	// register crd object
	crdResourceList = append(crdResourceList, minio.NewMinioResourceDefine())
	for _, crObj := range crdResourceList {
		if err := register.RegisterCRDWithObject(extClientSet, crObj); err != nil {
			return err
		}
		if err := register.WaitForCRDEstablished(extClientSet, crObj.GetName()); err != nil {
			return err
		}
	}
	return nil
}
