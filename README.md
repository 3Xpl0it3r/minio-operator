#### 介绍
&emsp;`minio-operator`在k8s集群中部署`minio`, `minio`集群最小为4个实例

#### Example
&emsp;杨例文件如下:
```yaml
# apiVersion 固定
apiVersion:  miniooperator.3xpl0it3r.cn/v1alpha1 
# Kind 固定
kind: Minio
metadata:
  # minio集群名称
  name: minio
spec:
  # 副本数
  replicas: 4
  image: "minio/minio"
  hostpath: "/data/minio"
```

&emsp;`replicas`为副本数,`replicas`个数要么为1,要么`>4`,当k8s只有一个节点的时候,operator会固定的将replicas设置为1(无论用户设置多少,单节点运行多实例没啥意,服务器磁盘基本都做了raid)


#### 使用
&emsp;
```bash
$ cat fake.yaml
apiVersion:  miniooperator.3xpl0it3r.cn/v1alpha1
kind: Minio
metadata:
  name: minio
spec:
  replicas: 4
  image: "minio/minio"
  hostpath: "/data/minio"


$ kubectl apply -f fake.yaml
```
