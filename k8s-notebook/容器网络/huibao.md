https://www.lixueduan.com/posts/docker/10-bridge-network/
https://github.com/mz1999/blog/blob/master/docs/docker-network-bridge.md
《自己动手写Docker》
# network namespace

# 网络虚拟化技术
## veth 
Veth 是成对出现的虚拟网络设备 发送 Veth 一端虚 设备 请求会 另一端
备中发出。在容器的虚拟 场景中，经常会使 Veth 连接不同的网络Namespace

## Linux bridge
Bridge 虚拟设备是用来桥接的网络设备，它相当于现实世界中的交换机 可以连接不同的网络设备，当请求到达 Bridge 设备时，可以通过报文中的 Mac 地址进行广播或转发。例如，创建 Bridge 设备，来连接 Namespace 中的网络设备和宿主机上的网络

## 实验
![alt text](bridge-vethpair.png)
通过 Linux 提供的各种虚拟设备以及 iptables 模拟出了 Docker bridge 网络模型，并测试了几种场景的网络互通
仅仅创建出Namespace网络隔离环境来模拟容器行为：
https://www.lixueduan.com/posts/docker/10-bridge-network/

1. 创建“容器”
```bash
sudo ip netns add docker0
sudo ip netns add docker1
```
查看创建出的网络Namesapce：
![alt text](image-1.png)
![alt text](image-2.png)

2. 创建Veth pairs
```bash
sudo ip link add veth0 type veth peer name veth1
sudo ip link add veth2 type veth peer name veth3
```
查看创建出的Veth pairs：
![alt text](image-3.png)
将Veth一端放入容器：
```bash
sudo ip link set Tveth0 netns Tdocker0
sudo ip link set Tveth2 netns Tdocker1
```
![alt text](image-4.png)
进入容器Tdocker0查看网卡：发现Tveth0已经放入了“容器”Tdocker0内，并且可以看出与index为8的是一组veth pair
![alt text](image-5.png)
在宿主机上查看网卡ip addr，发现veth0和veth2已经消失，确实是放入“容器”内了。

3. 创建bridge
安装bridge管理工具brctl `sudo apt-get install bridge-utils`
创建网桥`sudo brctl addbr Tbr0`
将Veth的另一端接入bridge 
```bash
sudo brctl addif Tbr0 Tveth1
sudo brctl addif Tbr0 Tveth3
```
查看效果`sudo brctl show`
![alt text](image-6.png)
两个网卡Tveth1和Tveth3已经“插”在bridge上。

4. 为"容器“内的网卡分配IP地址，并激活上线【意味着网络接口卡（网卡）已经被启用，并且已经连接到网络，可以开始发送和接收数据。】
新创建的netns里只有lo网卡：
![alt text](image-7.png)
```bash
 sudo ip netns exec Tdocker0 ip a add 172.18.0.2/24 dev Tveth0
 sudo ip netns exec Tdocker0 ip link set Tveth0 up
 ```
![alt text](image-8.png)
同样在Tdocker1中执行：
```bash
sudo ip netns exec Tdocker1 ip addr add 172.18.0.3/24 dev Tveth2
sudo ip netns exec Tdocker1 ip link set Tveth2 up
```
![alt text](image-9.png)

5. Veth另一端的网卡激活上线
`sudo ip link set Tveth1 up`
`sudo ip link set Tveth3 up`

6. 为bridge分配IP地址，激活上线
```bash
sudo ip addr add 172.18.0.1/24 dev Tbr0
sudo ip link set Tbr0 up
```
![alt text](image-10.png)

7. 容器”间的互通测试

# iptables

# docker容器网络
docker的4种网络模式
bridge模式
container模式
host模式
none模式
# 跨主机容器网络