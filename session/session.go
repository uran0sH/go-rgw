package session

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go-rgw/allocator"
	"go-rgw/connection"
	"io"
)

func SaveObject(objectName, bucketName string, object io.ReadCloser, hash string, metadataM map[string][]string,
	acl string) (err error) {
	// rollback
	rollback := func(rollback func()) {
		if err != nil {
			rollback()
		}
	}

	// 1M
	var objectCache = make([]byte, 1024*1024)
	var data []byte
	// read the object
	for {
		n, err := object.Read(objectCache)
		if err != nil && err != io.EOF {
			return err
		}
		data = append(data, objectCache[:n]...)
		if err == io.EOF {
			break
		}
	}

	// allocate id
	clusterID, err := allocator.GetClusterID()
	if err != nil {
		return err
	}
	bucketID := connection.MysqlMgr.MySQL.FindBucket(bucketName).BucketID
	if bucketID == "" {
		return fmt.Errorf("bucket doesn't exist")
	}
	oid := allocator.AllocateObjectID(bucketID, clusterID)

	// check hash
	check := md5.New()
	hashcache := bufio.NewReader(bytes.NewReader(data))
	_, err = io.Copy(check, hashcache)
	if err != nil {
		return
	}
	hashC := base64.StdEncoding.EncodeToString(check.Sum(nil))
	if hashC != hash {
		return fmt.Errorf("hash inconsistency")
	}

	// save object
	err = connection.CephMgr.Ceph.WriteObject(connection.BucketData, oid, data, 0)
	if err != nil {
		return
	}
	defer rollback(func() { rollbackSaveObject(oid) })
	// remove the existed object
	objectID := connection.MysqlMgr.MySQL.FindObject(objectName).ObjectID
	if objectID != "" {
		err = connection.CephMgr.Ceph.DeleteObject(connection.BucketData, objectID)
		if err != nil {
			return
		}
	}
	// save metadata, acl and objectid to database
	metadata, err := json.Marshal(&metadataM)
	if err != nil {
		return
	}
	// object name
	name := bucketID + "-" + objectName
	err = connection.MysqlMgr.MySQL.SaveObjectTransaction(name, oid, string(metadata), acl)
	if err != nil {
		return
	}
	return nil
}

// rollback save object
func rollbackSaveObject(id string) {
	// async
	go func() {
		for {
			err := connection.CephMgr.Ceph.DeleteObject(connection.BucketData, id)
			if err == nil {
				break
			}
		}
	}()
}

func CreateBucket(bucketName string) {
	bucketID := allocator.AllocateBucketID()
	connection.MysqlMgr.MySQL.CreateBucket(bucketName, bucketID)
}

func GetObject(bucketName, objectName string) ([]byte, error) {
	bucketID := connection.MysqlMgr.MySQL.FindBucket(bucketName).BucketID
	name := bucketID + "-" + objectName
	oid := connection.MysqlMgr.MySQL.FindObject(name).ObjectID
	if oid == "" {
		return nil, fmt.Errorf("the objectName doesn't exist")
	}
	data := make([]byte, 1024*1024)
	n, err := connection.CephMgr.Ceph.ReadObject(connection.BucketData, oid, data, 0)
	return data[:n], err
}
