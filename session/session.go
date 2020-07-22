package session

import (
	"fmt"
	"go-rgw/allocator"
	"go-rgw/connection"
)

// SaveMap saves object to the ceph cluster
//func SaveMap(filename string, data []byte) error {
//	oid := connection.MysqlMgr.MySQL.FindMapByName(filename).ID
//	if oid == "" {
//		oid = allocator.AllocateID()
//		connection.MysqlMgr.MySQL.SaveMap(filename, oid)
//	}
//	err := connection.CephMgr.Ceph.WriteObject("test-pool", oid, data, 0)
//	fmt.Println(err)
//	return err
//}

// SaveDataMetadata saves data and metadata and ensures the data and metadata consistency.
// metadataName is {bucketName}.meta_{filename}
// dataName is {bucketName}.{filename}
func SaveDataMetadata(bucketName string, filename string, data []byte, metadata []byte) (err error) {
	// rollback
	rollback := func(rollback func()) {
		if err != nil {
			rollback()
		}
	}
	// join the bucketname and filenmae
	metadataName := bucketName + ".meta_" + filename
	dataName := bucketName + "." + filename

	// get UUID
	bucketID := connection.MysqlMgr.MySQL.FindMapByName(bucketName).ID
	if bucketID == "" {
		bucketID = allocator.AllocateBucketID()
		connection.MysqlMgr.MySQL.SaveMap(bucketName, bucketID)
	}
	clusterID, err := allocator.GetClusterID()
	if err != nil {
		return err
	}
	oid := allocator.AllocateObjectID(bucketID, clusterID)
	metaOid := allocator.AllocateMetadataID(oid)

	// save data and metadata to the Ceph cluster
	err = connection.CephMgr.Ceph.WriteObject(connection.BucketData, metaOid, metadata, 0)
	if err != nil {
		return err
	}
	defer rollback(func() { rollbackSaveObject(metaOid) })
	err = connection.CephMgr.Ceph.WriteObject(connection.BucketData, oid, data, 0)
	if err != nil {
		return err
	}
	defer rollback(func() { rollbackSaveObject(oid) })

	// save the map between filename and objectID to the Database
	dataMapID := connection.MysqlMgr.MySQL.FindMapByName(dataName).ID
	metaMapID := connection.MysqlMgr.MySQL.FindMapByName(metadataName).ID
	if dataMapID != "" && metaMapID != "" {
		err = connection.MysqlMgr.MySQL.UpdateMapTransaction(dataName, oid, metadataName, metaOid)
		if err != nil {
			return err
		}
	}
	err = connection.MysqlMgr.MySQL.SaveMapTransaction(dataName, oid, metadataName, metaOid)
	return err
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

func GetObject(filename string) ([]byte, error) {
	oid := connection.MysqlMgr.MySQL.FindMapByName(filename).ID
	if oid == "" {
		return nil, fmt.Errorf("the filename doesn't exist")
	}
	data := make([]byte, 100)
	n, err := connection.CephMgr.Ceph.ReadObject(connection.BucketData, oid, data, 0)
	return data[:n], err
}
