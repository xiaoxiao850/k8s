# 容器运行时概述
container runtime是一个管理容器所有运行进程的工具，包括创建、删除容器、打包和共享容器。
低级容器运行时：和底层操作系统打交道，管理容器的生命周期，常见的容器运行时如runc，
高级容器运行时：是对低级容器运行时的进一步封装，专注于多个容器管理及容器镜像管理，常见的容器运行时：containerd，Docker，cri-o,podman

# containerd
是一个守护进程，在单个主机上管理主机完整的容器生命周期，包括创建、启动、停止容器以及存储镜像、配置挂载、配置网络等。
containerd 从docker Engine剥离出来
可以`ctr version`查看版本
```bash
aiedge@master-test-251:~$ kubectl get no -Aowide
NAME              STATUS     ROLES           AGE   VERSION   INTERNAL-IP      EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION      CONTAINER-RUNTIME
master-test-251   Ready      control-plane   77d   v1.25.3   192.168.20.251   <none>        Ubuntu 20.04.6 LTS   5.4.0-174-generic   containerd://1.6.20
worker-test-252   Ready   <none>          76d   v1.25.3   192.168.20.252   <none>        Ubuntu 20.04.6 LTS   5.4.0-170-generic   containerd://1.6.20
worker-test-253   Ready   <none>          76d   v1.25.3   192.168.20.253   <none>        Ubuntu 20.04.6 LTS   5.4.0-170-generic   containerd://1.6.20
```
## namespace
containerd 相比Docker多了`namespace`的概念，常见的namespace：`default`，`moby`，`k8s.io`
`default`：不指定时默认
`moby`：Docker使用的namespace
`k8s.io`：kubelet与crictl使用的namespace
`ctr -n k8s.io image ls`
## 容器操作
`ctr -n k8s.io container ls`查看容器列表
`ctr -n k8s.io task ls`查看运行的容器列表
`ctr run`等效于`ctr container create`+`ctr task start`
`ctr c info nginx_1`查看容器详细配置
`ctr  -n k8s.io t metrics nginx_1`查看容器使用指标
