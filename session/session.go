package session

import (
	"fmt"
	"go-rgw/allocator"
	"go-rgw/connection"
)

// SaveObject saves object to the ceph cluster
func SaveObject(filename string, data []byte) error {
	oid := connection.MysqlMgr.MySQL.FindObjectByName(filename).ObjectID
	if oid == "" {
		oid = allocator.AllocateID()
		connection.MysqlMgr.MySQL.SaveObject(filename, oid)
	}
	err := connection.CephMgr.Ceph.WriteObject("test-pool", oid, data, 0)
	fmt.Println(err)
	return err
}

// SaveDataMetadata saves data and metadata and ensures the data and metadata consistency.
// metadataName is {bucketName}.meta_{filename}
// dataName is {bucketName}.{filename}
func SaveDataMetadata(bucketName string, filename string, data []byte, metadata []byte) error {
	metadataName := bucketName + ".meta_" + filename
	dataName := bucketName + "." + filename
	// get UUID
	metaOid := allocator.AllocateID()
	oid := allocator.AllocateID()

	// save data and metadata to the Ceph cluster
	err := connection.CephMgr.Ceph.WriteObject(connection.BucketData, metaOid, metadata, 0)
	if err != nil {
		return err
	}
	err = connection.CephMgr.Ceph.WriteObject(connection.BucketData, oid, data, 0)
	if err != nil {
		// rollback
		// this goroutine is used to delete object
		go func() {
			for {
				err := connection.CephMgr.Ceph.DeleteObject(connection.BucketData, metaOid)
				if err == nil {
					break
				}
			}
		}()
		return err
	}

	// save the map between filename and objectID to the Database
	rel := connection.MysqlMgr.MySQL.FindObjectByName(dataName)
	if rel != (connection.Object{}) {
		connection.MysqlMgr.MySQL.UpdateObject(dataName, oid)
	} else {
		connection.MysqlMgr.MySQL.SaveObject(dataName, oid)
	}
	rel = connection.MysqlMgr.MySQL.FindObjectByName(metadataName)
	if rel != (connection.Object{}) {
		connection.MysqlMgr.MySQL.UpdateObject(metadataName, metaOid)
	} else {
		connection.MysqlMgr.MySQL.SaveObject(metadataName, metaOid)
	}
	return nil
}

func GetObject(filename string) ([]byte, error) {
	oid := connection.MysqlMgr.MySQL.FindObjectByName(filename).ObjectID
	if oid == "" {
		return nil, fmt.Errorf("the filename doesn't exist")
	}
	data := make([]byte, 100)
	n, err := connection.CephMgr.Ceph.ReadObject(connection.BucketData, oid, data, 0)
	return data[:n], err
}
