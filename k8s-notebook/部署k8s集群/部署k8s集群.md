# 部署及升级k8s集群



## 部署k8s集群

### 步骤：

​					https://k8s.easydoc.net/docs/dRiQjyTY/28366845/6GiNOzyZ/nd7yOvdY

​					https://blog.csdn.net/wzy_168/article/details/100582322



kubeadm init --image-repository=registry.aliyuncs.com/google_containers



清除之前的

sudo kubeadm reset

sudo rm -r /etc/kubernetes/



sudo kubeadm token create --print-join-command

kubeadm join --token i50com.f72crkiist1a3sm6 192.168.20.235:6443 --discovery-token-ca-cert-hash sha256:8be061ec9959f6e5a2af39bac274c3e903cc51de40a96ae0633cacf110e861d9



### 部署遇到的问题：

https://blog.csdn.net/wzy_168/article/details/100582322

问题：flannel一直在CrashLoopBackOff\，日志中Erro 

![image-20230718151331434](image-20230718151331434.png)

![image-20230718151445113](image-20230718151445113.png)

![image-20230718151513527](image-20230718151513527.png)

解决：

修改 
/etc/kubernetes/manifests/kube-controller-manager.yaml
增加参数
--allocate-node-cidrs=true
--cluster-cidr=10.244.0.0/16
然后重启kubelet

![image-20230718151553157](image-20230718151553157.png)





2.schedule

![image-20230718162319119](image-20230718162319119.png)

https://blog.csdn.net/yujia_666/article/details/121082427

出现这种情况，是/etc/kubernetes/manifests/下的kube-controller-manager.yaml和kube-scheduler.yaml设置的默认端口是0导致的，解决方式是注释掉对应的port即可，

![image-20230718162342328](image-20230718162342328.png)



kubeadm join提示unable to fetch the kubeadm-config ConfigMap

![image-20230718154921593](image-20230718154921593.png)



3.join

3.1 udpIdleTimeout

kubectl -n kube-system edit cm kube-proxy

![image-20230718201710492](image-20230718201710492.png)





报错[kubelet-check] It seems like the kubelet isn‘t running or healthy.

[kubelet-check] The HTTP call equal to 'curl -sSL http://localhost:10248/healthz' failed with error: Get "http://localhost:10248/healthz": dial tcp 127.0.0.1:10248: connect: connection refused.

https://blog.csdn.net/qq_39111199/article/details/128410231?spm=1001.2101.3001.6650.3&utm_medium=distribute.pc_relevant.none-task-blog-2%7Edefault%7EBlogCommendFromBaidu%7ERate-3-128410231-blog-91361477.235%5Ev38%5Epc_relevant_default_base3&depth_1-utm_source=distribute.pc_relevant.none-task-blog-2%7Edefault%7EBlogCommendFromBaidu%7ERate-3-128410231-blog-91361477.235%5Ev38%5Epc_relevant_default_base3&utm_relevant_index=4



## 升级 Kubernetes 集群

https://www.chenshaowen.com/blog/how-upgrade-kubernetes-cluster.html

https://tinychen.com/20221224-k8s-12-kubeadm-upgrade-cluster/

升级思路：驱逐负载、摘除流量之后，先升级master节点，然后再一次升级每台node节点。注： kubeadm 跨版本升级容易出问题

当前集群：

![image-20230808161307698](image-20230808161307698.png)

### 步骤：

添加k8s安装源

```shell
apt-get update && apt-get install -y apt-transport-https
curl https://mirrors.aliyun.com/kubernetes/apt/doc/apt-key.gpg | apt-key add - 
cat <<EOF >/etc/apt/sources.list.d/kubernetes.list
deb https://mirrors.aliyun.com/kubernetes/apt/ kubernetes-xenial main
EOF
apt-get update
```

 查看并升级

```shell
# 查看所有可用的kubeadm的版本
$ sudo apt list --all-versions kubeadm
Listing... Done
kubeadm/kubernetes-xenial 1.27.4-00 amd64 [upgradable from: 1.22.1-00]
kubeadm/kubernetes-xenial 1.27.3-00 amd64
...
kubeadm/kubernetes-xenial,now 1.22.1-00 amd64 [installed,upgradable to: 1.27.4-00]

# 升级 Kubeadm
$ sudo apt-get install -y kubeadm=1.23.17-00
```

```shell
$ kubeadm upgrade plan --ignore-preflight-errors=ControlPlaneNodesReady,CoreDNSUnsupportedPlugins,CoreDNSMigration
```

查看升级计划

```shell
$ kubeadm upgrade plan --ignore-preflight-errors=ControlPlaneNodesReady,CoreDNSUnsupportedPlugins,CoreDNSMigration
...
[upgrade/versions] Target version: v1.23.17
[upgrade/versions] Latest version in the v1.22 series: v1.22.17

Components that must be upgraded manually after you have upgraded the control plane with 'kubeadm upgrade apply':
COMPONENT   CURRENT       TARGET
kubelet     3 x v1.21.4   v1.23.17

Upgrade to the latest stable version:

COMPONENT                 CURRENT    TARGET
kube-apiserver            v1.22.17   v1.23.17
kube-controller-manager   v1.22.17   v1.23.17
kube-scheduler            v1.22.17   v1.23.17
kube-proxy                v1.22.17   v1.23.17
CoreDNS                   v1.8.4     v1.8.6
etcd                      3.5.0-0    3.5.1-0

You can now apply the upgrade by executing the following command:

        kubeadm upgrade apply v1.23.17

Note: Before you can perform this upgrade, you have to update kubeadm to v1.23.17.

_____________________________________________________________________


The table below shows the current state of component configs as understood by this version of kubeadm.
Configs that have a "yes" mark in the "MANUAL UPGRADE REQUIRED" column require manual config upgrade or
resetting to kubeadm defaults before a successful upgrade can be performed. The version to manually
upgrade to is denoted in the "PREFERRED VERSION" column.

API GROUP                 CURRENT VERSION   PREFERRED VERSION   MANUAL UPGRADE REQUIRED
kubeproxy.config.k8s.io   v1alpha1          v1alpha1            no
kubelet.config.k8s.io     v1beta1           v1beta1             no

```

拉取依赖的镜像

```shell
$ docker pull hubimage/kube-apiserver:v1.23.17
$ docker tag hubimage/kube-apiserver:v1.23.17 k8s.gcr.io/kube-apiserver:v1.23.17
```

开始升级 Kubernetes 集群

```bash
$ kubeadm upgrade apply v1.23.4 --ignore-preflight-errors=ControlPlaneNodesReady,CoreDNSUnsupportedPlugins,CoreDNSMigration
```

![image-20230808193204449](image-20230808193204449.png)

对其余的控制面节点更新命令也对应变成了

```shell
$ kubeadm upgrade node （在node上执行不是master）

[upgrade] Reading configuration from the cluster...
[upgrade] FYI: You can look at this config file with 'kubectl -n kube-system get cm kubeadm-config -o yaml'
[preflight] Running pre-flight checks
[preflight] Skipping prepull. Not a control plane node.
[upgrade] Skipping phase. Not a control plane node.
[kubelet-start] Writing kubelet configuration to file "/var/lib/kubelet/config.yaml"
[upgrade] The configuration for this node was successfully updated!
[upgrade] Now you should go ahead and upgrade the kubelet package using your package manager.

```



升级 kubectl 和 kubelet

- 上面的更新操作完成之后我们只是把K8S集群中的相关pod都升级了一遍，但是kubelet并没有升级，因此这里看到的版本信息还是

  ![image-20230808194023098](image-20230808194023098.png)

- 升级kubelet之前我们要对节点进行驱逐操作，将上面除了daemonset之外的全部工作负载都驱逐掉

  ```shell
  $ kubectl drain xx-test-master235 --ignore-daemonsets
  ```

- 更新kubelet和kubectl

  ```shell
  $ sudo apt install kubelet=1.23.17-00 kubectl=1.23.17-00 
  
  $ systemctl daemon-reload
  $ systemctl restart kubelet
  
  # 查看日志检查相关服务是否正常
  $ systemctl status kubelet -l
  $ journalctl -xeu kubelet
  
  ```

  这时候再查看相关的状态就能看到节点已经升级成功：

  ![image-20230808200458021](image-20230808200458021.png)

  确定该节点正常后，就能恢复它的调度

  ```shell
  $ kubectl uncordon xx-test-master235
  ```

  ![image-20230808201037089](image-20230808201037089.png)

​		

​		接着对剩下的两个节点进行同样的操作，都更新完成之后我们可以看到整个集群的控制面就完成升级了。

​		





### 升级遇到的问题：

1. 升级过程中node notready了，查看节点描述：cni有问题。解决：重新部署flannel，再次查看node,都处于ready

2. upgrade卡住，解决：先提前拉镜像。 kubeadm config images pull  --v=5

3. 找不到对应的kubelet kubectl 版本。解决：在节点上，添加同样的k8s源。

   ```bash
   $ apt-get update && apt-get install -y apt-transport-https
   curl https://mirrors.aliyun.com/kubernetes/apt/doc/apt-key.gpg | apt-key add - 
   cat <<EOF >/etc/apt/sources.list.d/kubernetes.list
   deb https://mirrors.aliyun.com/kubernetes/apt/ kubernetes-xenial main
   EOF
   apt-get update
   ```

   

4. 升级成功但是notready

   ![image-20230808205636161](image-20230808205636161.png)

   

5. 1



## kubeEdge部署

使用Keadm进行部署。

https://blog.csdn.net/weixin_38159695/article/details/118486461

![image-20230809172539308](image-20230809172539308.png)

### 步骤：

#### 1. 安装Cloudcore

##### 1.1下载keadm

```shell
$ wget https://github.com/kubeedge/kubeedge/releases/download/v1.13.1/keadm-v1.13.1-linux-amd64.tar.gz
#首先解压keadm压缩包
$ tar -zxvf keadm-v1.10.1- linux-amd64.tar.gz

#进入解压文件
$ cd keadm-v1.10.1- linux-amd64/ keadm

#将keadm的二进制文件拷贝到系统环境变量中
$ cp keadm /usr/local/bin

#查看版本信息
$ keadm version
```



##### 1.2 使用keadm安装cloudcore

```shell
$ sudo keadm init --advertise-address=192.168.20.235 --kubeedge-version=1.13.1
#检查
$ ps -ef|grep cloudcore
```



#### 2.  边缘节点安装keadm

#### 3.  边缘节点加入集群

```shell
$ sudo keadm join --cloudcore-ipport=192.168.20.235:32736 --certport 30821 --runtimetype=docker --token=bbbc5e39a9968c5ccfdd5f0e029bc2670dd00567ad453437838ffb34a13a4a7f.eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTE2NjExMjB9.OCF4VhFCkPxGxYA6imAZsvFpfjU0v3Z4q7GnRj1vnVw

```



### 部署遇到的问题：

问题：execute keadm command failed:  edge node join failed: copy resources failed: rpc error: code = DeadlineExceeded desc = context deadline exceeded...

解决：添加一个参数信息：`--runtimetype=docker`  。`--kubeedge-version=1.13.0`



问题：[Join edge-node: Faild to get CA certificat](https://github.com/kubeedge/kubeedge/issues/3851)

CloudCore Service启动如问题图所示，只开放内部端口，不开放外部端口。



```shell
root@xx-test-edge239:/home/aiedge# journalctl -u edgecore.service -xe
...
Aug 09 21:19:19 xx-test-edge239 edgecore[82823]: I0809 21:19:19.142119   82823 common.go:97] start connect to mqtt server with client id: hub-client-sub-1691587159
Aug 09 21:19:19 xx-test-edge239 edgecore[82823]: I0809 21:19:19.142223   82823 common.go:99] client hub-client-sub-1691587159 isconnected: false
Aug 09 21:19:19 xx-test-edge239 edgecore[82823]: F0809 21:19:19.142477   82823 certmanager.go:96] Error: failed to get CA certificate, err: Get "https://192.168.20.235:10002/ca.crt": dial tcp 192.168.20.23>
Aug 09 21:19:19 xx-test-edge239 systemd[1]: edgecore.service: Main process exited, code=exited, status=1/FAILURE

root@xx-test-master235:/home/aiedge# kubectl logs cloudcore-77b5dfdd57-tvdmt -n kubeedge |tail -n 20
I0809 17:51:58.660803       1 core.go:46] starting module edgecontroller
I0809 17:51:58.660971       1 upstream.go:140] start upstream controller
I0809 17:51:58.661179       1 core.go:46] starting module devicecontroller
I0809 17:51:58.661458       1 core.go:46] starting module synccontroller
I0809 17:51:58.661521       1 downstream.go:334] start downstream controller
I0809 17:51:58.690540       1 downstream.go:959] Start downstream devicecontroller
W0809 17:51:59.442325       1 iptables.go:111] failed to delete all rules in tunnel port iptables chain: error flushing chain "TUNNEL-PORT": exit status 1: iptables: No chain/target/match by that name.
I0809 17:51:59.543972       1 server.go:247] Ca and CaKey don't exist in local directory, and will read from the secret
I0809 17:51:59.644593       1 server.go:251] Ca and CaKey don't exist in the secret, and will be created by CloudCore
I0809 17:52:00.241985       1 server.go:285] CloudCoreCert and key don't exist in local directory, and will read from the secret
I0809 17:52:00.252508       1 server.go:290] CloudCoreCert and key don't exist in the secret, and will be signed by CA
I0809 17:52:00.356648       1 tunnelserver.go:157] Succeed in loading TunnelCA from CloudHub
I0809 17:52:00.357261       1 tunnelserver.go:170] Succeed in loading TunnelCert and Key from CloudHub
...
```

解决：

192.168.20.235:32736 --certport 30821

问题：

![image-20230810120158251](image-20230810120158251.png)

解决 停edge节点kubelet

sudo systemctl stop kubelet



问题：Failed to check the running environment: Kube-proxy should not running on edge node when running edgecore。如果安装有kubeproxy，在启动edgecore时日志会出现报错

![image-20230810150200987](image-20230810150200987.png)

解决：

https://segmentfault.com/a/1190000040225049

```bash
kubectl edit daemonset -n kube-system kube-proxy
```

新增亲和性配置(affinity):

```yaml
  spec:
 affinity:
   nodeAffinity:
     requiredDuringSchedulingIgnoredDuringExecution:
       nodeSelectorTerms:
         - matchExpressions:
           - key: node-role.kubernetes.io/edge
             operator: DoesNotExist
```

重启edge节点上docker服务

```bash
 systemctl restart docker
```

重启edgecore



问题：flannel



解决;

https://github.com/kubeedge/kubeedge/issues/2677

add a new flannel in edge node，边端的flannel要重新部署

![image-20230810205935643](image-20230810205935643.png)

http://192.168.20.2/aiedge/platform-install/-/blob/update-1.13.0/vendor/flannel/edge-flannel.yml

http://192.168.20.2/aiedge/platform-install/-/blob/update-1.13.0/vendor/flannel/kube-flannel.yml

```
 kubectl edit ds -n kube-system kube-flannel-ds
add a new flannel in edge node using edge list-watch feature.
...
spec:
  selector:
    matchLabels:
      k8s-app: flannel-edge
  template:
    metadata:
      labels:
        k8s-app: flannel-edge
        tier: node
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: node-role.kubernetes.io/agent
                operator: Exists
              - key: kubernetes.io/arch
                operator: In
                values:
                - amd64
      containers:
      - args:
        - --ip-masq
        - --kube-subnet-mgr
        - --kube-api-url=http://127.0.0.1:10550
        name: kube-flannel-edge
        ...
```

Enable dyanmic controller

```
# vim /etc/kubeedge/config/cloudcore.yaml
modules:
  ..
  dynamicController:
    enable: true
..
```

restart cloudcore service



Update configuration of edgecore

```
# vim /etc/kubeedge/config/edgecore.yaml
modules:
  ..
  edgeMesh:
    enable: false
  ..
  edged:
    ..
    networkPluginName: cni
    clusterDNS: "10.96.0.10"    //<----- If you not run kube-proxy in edge node, change it to kube-dns pod ip
    clusterDomain: "cluster.local"
  ..
  metaManager:
    metaServer:
      enable: true
  ..
```

restart edgecore

```
service edgecore restart
```





https://kubeedge.io/zh/docs/advanced/debug/
