package session

import (
	"encoding/json"
	"fmt"
	"go-rgw/allocator"
	"go-rgw/connection"
	"io"
)

func SaveObject(objectName, bucketName string, object io.ReadCloser, metadataM map[string][]string,
	acl string) (err error) {
	// rollback
	rollback := func(rollback func()) {
		if err != nil {
			rollback()
		}
	}

	// 5M
	var objectCache = make([]byte, 5*1024*1024)
	var data []byte
	// read the object
	for {
		n, err := object.Read(objectCache)
		if err != nil && err != io.EOF {
			return
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

	// save object
	err = connection.CephMgr.Ceph.WriteObject(connection.BucketData, oid, data, 0)
	if err != nil {
		return
	}
	defer rollback(func() { rollbackSaveObject(oid) })

	// save metadata, acl and objectid to database
	metadata, err := json.Marshal(&metadataM)
	if err != nil {
		return
	}
	err = connection.MysqlMgr.MySQL.SaveObjectTransaction(objectName, oid, string(metadata), acl)
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

//func GetObject(filename string) ([]byte, error) {
//	oid := connection.MysqlMgr.MySQL.FindMapByName(filename).ID
//	if oid == "" {
//		return nil, fmt.Errorf("the filename doesn't exist")
//	}
//	data := make([]byte, 100)
//	n, err := connection.CephMgr.Ceph.ReadObject(connection.BucketData, oid, data, 0)
//	return data[:n], err
//}
