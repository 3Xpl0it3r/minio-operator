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
}
