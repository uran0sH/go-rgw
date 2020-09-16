# go-rgw
[English](README.md) | [中文](README_zh.md)  

## 介绍
因为Ceph 原生的RGW在一些企业场景下难以得到大规模应用，所以我们需要一个更加灵活轻便，适配更广泛的场景的对象存储网关。 go-rgw 是一个
基于 Ceph 的轻量级对象存储网关。目前已经实现上传，下载，分块上传功能。  
与Ceph的RGW将 metadata, acl 等保存在omap中不同的是， go-rgw 将这些数据单独保存在数据库（目前只支持MySQL，在未来会支持更多种）中，在Ceph集群中只存储数据，这样可以减少使用
Librados API 对Ceph中的数据进行读取，而且方便在数据量增大的时候增加Ceph集群，而不会影响已经运行的集群，减少数据的抖动。

## 系统架构图
![architecture](docs/architecture.png)

## 安装
* Step 1: Install Go on your machine and set up the environment by the following instruction at:  
[https://golang.org/doc/install](https://golang.org/doc/install)  
make sure you set up your `$GOPATH`   
* Step 2: Install the native RADOS library and development headers:  
On debian system(apt):  
`libcephfs-dev librbd-dev librados-dev`  
On rpm based systems (dnf, yum, etc):  
`libcephfs-devel librbd-devel librados-devel`  
* Step 3: Install `gcc` if you do not install.  
* Step 4: Modified `CEPH_VERSION` in the `Makefile`.
* Step 5: Make sure the Ceph's configuration file is in the `/etc/ceph/ceph.conf`.
* Step 6: Now you could start a local build by calling `make build` under the root path of this project.

## 使用
We offer a Go-SDK in []().  

* Step1: write configuration file, like  
```yaml
database:
  dialect: mysql
  username: root
  password: root
  address: 127.0.0.1:3306
  name: ceph
authorization: jwt
log:
  filename: ./rgw.log
```
All the above fields are required. The default path of the configuration file is `./application.yml`
* Step2: start server:  
`./octopus`  
or if you don't use the default path of the configuration file:  
`./octopus -config={path}`

## API
* create a bucket  
`/createbucket/:bucket`
* upload an object  
`/upload/:bucket/:object`
* download an object  
`/download/:bucket/:object`
* create multipartupload  
`/uploads/create/:bucket/:object`
* upload a part of an object  
`/uploads/upload/:bucket/:object`
* complete multipartupload  
`/uploads/complete/:bucket/:object`
* abort multipartupload
`/uploads/abort/:bucket/:object`

### User
* register  
`/register`
* login  
`/login`