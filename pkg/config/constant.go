package config

import (
	crgroup "github.com/3Xpl0it3r/minio-operator/pkg/apis/miniooperator.3xpl0it3r.cn"
	crapiv1alpha1 "github.com/3Xpl0it3r/minio-operator/pkg/apis/miniooperator.3xpl0it3r.cn/v1alpha1"
)

const (
	MinioLabelAnnotationPrefix = crgroup.GroupName + "/" + crapiv1alpha1.Version + "__"
	MinioAppNameLabel          = MinioLabelAnnotationPrefix + "app-name"

    MinioAppLocation = MinioLabelAnnotationPrefix + "nodeName"
)
