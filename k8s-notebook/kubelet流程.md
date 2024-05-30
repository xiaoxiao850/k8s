# kubelet 
## 概述
在 k8s 中，kubelet 是每个节点上运行的最重要组件之一。**是节点上的代理，负责容器的生命周期，并与控制平面api-server交互**。
- **容器管理**：kubelet负责管理节点上运行的容器。通过与**CRI**(ex: Docker、containerd等)进行交互(**启停监控容器状态**)。
- **Pod生命周期管理**：kubelet负责Pod的生命周期管理。通过与**api-server**通信，并接受Pod 源请求事件，进行相关处理(**创建、更新或删除**)，还需要监控 Pod 的健康状态，并在需要时重启失败的容器。
- 资源管理：kubelet 负责监控节点的资源使用情况，并确保容器的资源请求与限制得到满足。
- api-server 通信：kubelet 与 api-server 需要进行通信。它会定期向 api-server 发送节点状态、接收来自 api-server 的通知，并进行相应的处理
- 健康检查和自愈能力：kubelet 负责执行容器的健康检查(probe)，并在容器不健康或故障时，采取相应的操作。(重启、替换或上报等)

## Kubelet 上报 Node 状态
`nodeStatusUpdateFrequency` 是 kubelet 计算节点状态的频率。 如果没有启用节点租约功能，这也是 kubelet 将节点状态发布给 master 的频率。

## Kubelet 进程退出会发生什么
```bash
停止 kubelet 进程：
systemctl stop kubelet

查看 kubelet 状态：
systemctl status kubelet
```
- 当 `Node` 上 `kubelet` 进程退出 <= 5min，`Node` 立即变为 `NotReady`，pod 都为正常，不会有 `restart`
- 当 `Node` 上 `kubelet` 进程退出 > 5min，pod 变为 `Terminating`，新创建 pod 满足期望的副本数；

## Node 重启 (~10s)
- 当 `Node` 重启 (~10s)、关机时长 <= 5min，表现与上面第一点 `kubelet` 进程退出 <= 5min 基本一致；
- 当 `Node` 关机时长 > 5min，表现与上面第二点 `kubelet` 进程退出 > 5min 基本一致；

上述四种情况下，`kubectl exec/logs` 都会异常提示 `connection refused` 或 `i/o timeout`；