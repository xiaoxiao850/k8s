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
https://www.lixueduan.com/posts/docker/10-bridge-network/

# iptables

# docker容器网络
docker的4种网络模式
bridge模式
container模式
host模式
none模式
# 跨主机容器网络