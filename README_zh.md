# go-rgw
[English](README.md) | [中文](README_zh.md)  

## 介绍
因为Ceph 原生的RGW在一些企业场景下难以得到大规模应用，所以我们需要一个更加灵活轻便，适配更广泛的场景的对象存储网关。 go-rgw 是一个
基于 Ceph 的轻量级对象存储网关。目前已经实现上传，下载，分块上传功能。  
与Ceph的RGW将 metadata, acl 等保存在omap中不同的是， go-rgw 将这些数据单独保存在数据库（目前只支持MySQL，在未来会支持更多种）中，在Ceph集群中只存储数据，这样可以减少使用
Librados API 对Ceph中的数据进行读取，而且方便在数据量增大的时候增加Ceph集群，而不会影响已经运行的集群，减少数据的抖动。

## 系统架构图
![architecture](docs/architecture.png)
