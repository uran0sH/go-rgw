# go-rgw
[English](README.md) | [中文](README_zh.md)  

## Table of contents
- [Introduction](#introduction)
- [Architecture](#architecture)

## Introduction
We need an object storage gateway which is more flexible, lightweight, and adaptable to a wider range of scenarios 
because the RGW of Ceph is difficult to be applicated on a large scale in some enterprise scenarios. Go-rgw is a 
lightweight gateway of ceph based on go-ceph. So far we have implemented some functions: upload, download and 
multipartupload. Also, we will add some media functions.  
We store the metadata and acl of the object in the database, like MySQL, ES and so on, and we only store the data of 
object in the ceph, which is convenient to increase the Ceph cluster when we need more space to store data. Also, 
decreasing the number of reads and writes to the Ceph could improve the system performance.

## Architecture
![architecture](docs/architecture.png)


## Installation Guide
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

## Example Usage
We offer a Go-SDK in [octopus-sdk]().

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