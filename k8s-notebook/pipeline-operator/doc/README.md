# 部署
## 一些步骤
```
kubebuilder init --domain ndsl.cn
kubebuilder create api --group distri-infer --version v1 --kind Pipeline
```

> crd `pipelines.distri-infer.ndsl.cn`

修改types.go
`make manifests`

写template 写解析渲染

todo:
写reconcil逻辑

测试
```bash
make manifests
make install
kubectl get crd

kubectl apply -f ./config/samples/namespace.yaml
kubectl get ns

！注意：需要保证集群nfs-csi部署成功


go run ./cmd/main.go
kubectl apply -f ./config/samples/distri-infer_v1_pipeline.yaml

```
删除
```bash

```

检查pod 内部 模型文件、env 
```bash

```

## 封装镜像测试
$ make docker-build docker-push IMG=xiaox1958141/pipelineOperator:v1

根据 `IMG` 指定的镜像将控制器部署到集群中:
$ make deploy IMG=xiaox1958141/pipelineOperator:v1