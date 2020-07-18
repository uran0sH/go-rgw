package session

import (
	"fmt"
	"go-rgw/allocator"
	"go-rgw/connection"
)

// SaveObject saves object to the ceph cluster
func SaveObject(filename string, data []byte) error {
	oid := mysqlManager.MySQL.FindByName(filename).ObjectID
	if oid == "" {
		oid = allocator.AllocateID()
		mysqlManager.MySQL.Save(filename, oid)
	}
	err := cephManager.Ceph.WriteObject("test-pool", oid, data, 0)
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
	err := cephManager.Ceph.WriteObject(connection.BucketData, metaOid, metadata, 0)
	if err != nil {
		return err
	}
	err = cephManager.Ceph.WriteObject(connection.BucketData, oid, data, 0)
	if err != nil {
		// rollback
		// this goroutine is used to delete object
		go func() {
			for {
				err := cephManager.Ceph.DeleteObject(connection.BucketData, metaOid)
				if err == nil {
					break
				}
			}
		}()
		return err
	}

	// save the map between filename and objectID to the Database
	rel := mysqlManager.MySQL.FindByName(dataName)
	if rel != (connection.FilenameToID{}) {
		mysqlManager.MySQL.Update(dataName, oid)
	} else {
		mysqlManager.MySQL.Save(dataName, oid)
	}
	rel = mysqlManager.MySQL.FindByName(metadataName)
	if rel != (connection.FilenameToID{}) {
		mysqlManager.MySQL.Update(metadataName, metaOid)
	} else {
		mysqlManager.MySQL.Save(metadataName, metaOid)
	}
	return nil
}

func GetObject(filename string) ([]byte, error) {
	oid := mysqlManager.MySQL.FindByName(filename).ObjectID
	if oid == "" {
		return nil, fmt.Errorf("the filename doesn't exist")
	}
	data := make([]byte, 100)
	n, err := cephManager.Ceph.ReadObject("test-pool", oid, data, 0)
	return data[:n], err
}
