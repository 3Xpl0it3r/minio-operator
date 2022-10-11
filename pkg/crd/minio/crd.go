package minio

import (
	crdapiv1alpha1 "github.com/3Xpl0it3r/minio-operator/pkg/apis/miniooperator.3xpl0it3r.cn/v1alpha1"
	extensionapiv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	jsonSchemePropsTypeAsInteger string = "integer"
	jsonSchemePropsTypeAsString  string = "string"
	jsonSchemePropsTypeAsObject  string = "object"
	jsonSchemePropsTypesAsNumber string = "number"
	jsonSchemePropsTypeAsArray   string = "array"
)

func NewMinioResourceDefine() *extensionapiv1.CustomResourceDefinition {
	crd := &extensionapiv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "minios" + "." + crdapiv1alpha1.SchemeGroupVersion.Group,
		},
		Spec: extensionapiv1.CustomResourceDefinitionSpec{
			Group: crdapiv1alpha1.SchemeGroupVersion.Group,
			Names: extensionapiv1.CustomResourceDefinitionNames{
				Plural:   "minios",
				Singular: "minio",
				Kind:     "Minio",
				ListKind: "MinioList",
			},
			Scope: extensionapiv1.ResourceScope(extensionapiv1.NamespaceScoped),
			Versions: []extensionapiv1.CustomResourceDefinitionVersion{
				{
					Name:    crdapiv1alpha1.Version,
					Served:  true,
					Storage: true,
					Schema: &extensionapiv1.CustomResourceValidation{
						OpenAPIV3Schema: &extensionapiv1.JSONSchemaProps{
							Type: jsonSchemePropsTypeAsObject,
							Properties: map[string]extensionapiv1.JSONSchemaProps{
								"apiVersion": {Type: jsonSchemePropsTypeAsString},
								"kind":       {Type: jsonSchemePropsTypeAsString},
								"metadata":   {Type: jsonSchemePropsTypeAsObject},
								"spec": {
									Type: jsonSchemePropsTypeAsObject,
									Properties: map[string]extensionapiv1.JSONSchemaProps{
										"replicas": {Type: jsonSchemePropsTypeAsInteger},
										"image":    {Type: jsonSchemePropsTypeAsString},
										"hostpath": {Type: jsonSchemePropsTypeAsString},
										"credential": {
											Type: jsonSchemePropsTypeAsObject,
											Properties: map[string]extensionapiv1.JSONSchemaProps{
												"access_key": {Type: jsonSchemePropsTypeAsString},
												"secret_key": {Type: jsonSchemePropsTypeAsString},
											},
										},
										"buckets": {
											Type: jsonSchemePropsTypeAsArray,
											Items: &extensionapiv1.JSONSchemaPropsOrArray{
												Schema: &extensionapiv1.JSONSchemaProps{
													Type: jsonSchemePropsTypeAsString,
												},
											},
										},
									},
								},
								"status": {
									Type: jsonSchemePropsTypeAsObject,
									Properties: map[string]extensionapiv1.JSONSchemaProps{
										"inited": {Type: jsonSchemePropsTypeAsString},
									},
								},
							},
							Required: []string{"apiVersion", "kind", "metadata", "spec"},
						},
					},
					Subresources: &extensionapiv1.CustomResourceSubresources{
						Status: &extensionapiv1.CustomResourceSubresourceStatus{},
					},
				},
			},
			PreserveUnknownFields: false,
		},
	}
	return crd
}
