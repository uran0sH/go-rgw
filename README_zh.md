# go-rgw
[English](README.md) | [中文](README_zh.md)  

## 介绍
因为Ceph 原生的RGW在一些企业场景下难以得到大规模应用，所以我们需要一个更加灵活轻便，适配更广泛的场景的对象存储网关。 go-rgw 是一个
基于 Ceph 的轻量级对象存储网关。目前已经实现上传，下载，分块上传功能。  
与Ceph的RGW将 metadata, acl 等保存在omap中不同的是， go-rgw 将这些数据单独保存在数据库（目前只支持MySQL，在未来会支持更多种）中，在Ceph集群中只存储数据，这样可以减少使用
Librados API 对Ceph中的数据进行读取，而且方便在数据量增大的时候增加Ceph集群，而不会影响已经运行的集群，减少数据的抖动。

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
We provide a client example in `test/client_test.go`. 

* Step1: write configuration file, like  
```yaml
database:
  dialect: mysql
  username: root
  password: root
  address: 127.0.0.1:3306
  name: ceph
authorization: jwt
Ceph:
  user: " "
  monitors: " "
  keyring: " "
log:
  filename: ./rgw.log
```
All the above fields are required. The default path of the configuration file is `./application.yml`
* Step2: start server:  
`./octopus`  
or if you don't use the default path of the configuration file:  
`./octopus -config={path}`

## 系统架构
![architecture](docs/architecture.png)

### RESTful API
这个项目使用 RESTful 风格的 API ，详细的可以看 [API Reference](#api-reference)

### Session Controller
这个模块主要负责会话管理，包括上传下载大小文件，类事物化控制——保持数据和元数据的一致性。

#### 上传对象
`func SaveObject(objectName, bucketName string, object io.ReadCloser, hash string, metadataM map[string][]string, 
acl string) (err error)`  
这个函数将对象保存到 Ceph 集群中。  
`objectName` 是对象的名字  
`bucketName` 保存到的Bucket的名字  
`object` 是对象的数据  
`hash` 对象的Hash值（主要是 MD5 ）  
`metadataM` 对象的元数据  
`acl` 对象的 acl ，以字符串的形式保存到数据库中  
这个函数会确保上传对象数据和元数据的一致性，如果其中一个上传失败会进行回滚。与Ceph自带的对象存储网关不同的是，我们只将对象的数据保存到
 Ceph 集群中，而对象的元数据和 acl ，我们将其保存到数据库中（目前是保存到 MySQL 中）。保存对象的时候，我们会为其生成一个独一无二的 id 
 ，这个 id 由 clusterID.bucketID.objectUUID 组成，通过 clusterID 我们可以确定保存到的集群，这样就算新增集群也不会影响原有的数据。 
 bucketID 是 bucketName 对应的 ID ， objectUUID 通过 UUID 生成器生成。对象的 id 和对象的名字形成一组关系保存到数据库中。 

#### 分块上传对象
1. 初始化分块上传：上传对象名字，元数据，网关会暂时将这些数据进行缓存不写入数据库中。ACL。返回给客户端一个UploadID。  
2. 上传分块：客户端上传分块的时候携带uploadID+partNum（partNum可以是一个有序离散的整数），存储网关根据uploadID进行Object的保存并且
返回每个分块的hash值。  
3. 完成或放弃上传。  

#### 下载对象
`func GetObject(bucketName, objectName string) (data []byte, err error)`  
这个函数将根据 objectID 从 Ceph 集群获取对象数据。
`objectName` 是对象的名字  
`bucketName` 保存到的 Bucket 的名字  
`data` 对象的数据  
`GetObject()`先根据 bucketName 和 objectName 从数据库中取出对象 id 并且判断这个对象是否是分块上传的对象（数据库中isMultipart字段
标识是否是分块上传）。如果是分块上传会调用`readMultipartObject()`读取对象，否则调用`readOneObject()`。
 
##### 读取一个对象
`func readOneObject(oid string) ([]byte, error)`  
这个函数调用 connection 层封装好的 `ReadObject()` 从 Ceph 集群中读出一个对象。 `ReadObject()` 使用 `go-ceph` 的封装好的函数读取。

##### 读取分块上传对象
`func readMultipartObject(oid string) ([]byte, error)`  
先从数据库中取得所有分块的 objectID ，然后按照 partID 从小到大的顺序读取出来，然后拼接成一个完整的对象返回给上一层。
 
## API REFERENCE
* create a bucket  
`/createbucket/:bucket`  
`bucket` 创建的bucket名字
* upload an object  
`/upload/:bucket/:object`  
`bucket` 上传对象需要保存在哪个bucket  
`object` 上传对象的名字  
请求头：  
`Content-MD5` 值为对象的 MD5 值
* download an object  
`/download/:bucket/:object`  
`bucket` 下载对象保存在哪个bucket  
`object` 下载的对象的名字
* create multipartupload  
`/uploads/create/:bucket/:object`  
`bucket` 上传对象需要保存在哪个bucket  
`object` 上传对象的名字  
返回:  
 `uploadId`
* upload a part of an object  
`/uploads/upload/:bucket/:object`  
`bucket` 上传对象需要保存在哪个bucket  
`object` 上传对象的名字  
参数：  
`partNumber` 分块的编号，从小到大
请求头：  
`Content-MD5` 值为分块的 MD5 值
* complete multipartupload  
`/uploads/complete/:bucket/:object`  
`bucket` 上传对象需要保存在哪个bucket  
`object` 上传对象的名字  
请求参数：  
`UploadId` 上传对象的uploadId  
* abort multipartupload  
`/uploads/abort/:bucket/:object`  
`bucket` 上传对象需要保存在哪个bucket  
`object` 上传对象的名字  
请求参数：  
`UploadId` 上传对象的uploadId  
* blur image  
`/image/blur/:bucket/:object`  
`bucket` 上传对象需要保存在哪个bucket  
`object` 上传对象的名字  
请求参数：  
`sigma`  模糊参数  
`suffix` 文件后缀
* resize image  
`/image/resize/:bucket/:object`  
`bucket` 上传对象需要保存在哪个bucket  
`object` 上传对象的名字  
`width`  调整后图像宽  
`height` 调整后图像高  
`suffix` 文件后缀
* crop image  
`/image/cropAnchor/:bucket/:object`  
`bucket` 上传对象需要保存在哪个bucket  
`object` 上传对象的名字  
`width`  调整后图像宽  
`height` 调整后图像高  
`anchor`  以哪个点进行裁剪，有 "center", "topleft", "top", "topright", "left", "right", "bottomleft", "bottom", "bottomRight"  
`suffix` 文件后缀

### User
* register  
`/register`
* login  
`/login`