package v1alpha1

func MinioDefaulter(minio *Minio) {
    if minio.Spec.Replicas == 0 {
        minio.Spec.Replicas = 1
    }
    if minio.Spec.Image == "" {
        minio.Spec.Image = "registry.bizsaas.net/quay.io/minio"
    }
    if minio.Spec.HostPath == ""{
        minio.Spec.HostPath = "/data/minio"
    }
    if minio.Spec.Credential.AccessKey == "" {
        minio.Spec.Credential.AccessKey = "root123"
    }
    if minio.Spec.Credential.SecretKey == "" {
        minio.Spec.Credential.SecretKey = "adminadmin"
    }

    if minio.Spec.Port.ApiPort == 0 {
        minio.Spec.Port.ApiPort = 9001
    }
    if minio.Spec.Port.HttpPort == 0 {
        minio.Spec.Port.HttpPort = 9000
    }
    // minio.Spec.Port.NodePort is not set default for k8s will allocate a new one for it
}
