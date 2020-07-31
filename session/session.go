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
	"go-rgw/gc"
	"io"
	"sync"
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
	objectID := connection.MysqlMgr.MySQL.FindObject(bucketID + "-" + objectName).ObjectID
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
	err = connection.MysqlMgr.MySQL.SaveObjectTransaction(name, oid, string(metadata), acl, false)
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
	bucketID := allocator.AllocateUUID()
	connection.MysqlMgr.MySQL.CreateBucket(bucketName, bucketID)
}

func GetObject(bucketName, objectName string) (data []byte, err error) {
	bucketID := connection.MysqlMgr.MySQL.FindBucket(bucketName).BucketID
	name := bucketID + "-" + objectName
	object := connection.MysqlMgr.MySQL.FindObject(name)
	oid := object.ObjectID
	isMultipart := object.IsMultipart
	if !isMultipart {
		data, err = readOneObject(oid)
	} else {
		data, err = readMultipartObject(oid)
	}
	if err != nil {
		return nil, err
	}
	return data, err
}

// cache objectName->partObjectName
// var partsCache = make(map[string][]string, 100)

type PartsCache struct {
	partsCacheMap map[string][]string
	mutex         sync.Mutex
}

var partsCache = PartsCache{
	partsCacheMap: make(map[string][]string, 100),
}

type ObjectCache struct {
	objectCacheMap map[string]*Object
	mutex          sync.Mutex
}

type Object struct {
	objectID string
	metadata string
	acl      string
}

var objectCache = ObjectCache{
	objectCacheMap: make(map[string]*Object, 10),
}

// save objectName->objectID & metadata & acl
func CreateMultipartUpload(objectName, bucketName, metadata, acl string, isMultipart bool) error {
	clusterID, err := allocator.GetClusterID()
	if err != nil {
		return err
	}
	bucketID := connection.MysqlMgr.MySQL.FindBucket(bucketName).BucketID
	if bucketID == "" {
		return fmt.Errorf("bucket doesn't exist")
	}
	oid := allocator.AllocateObjectID(bucketID, clusterID)
	name := bucketID + "-" + objectName

	objectCache.mutex.Lock()
	objectCache.objectCacheMap[name] = &Object{
		objectID: oid,
		metadata: metadata,
		acl:      acl,
	}
	objectCache.mutex.Unlock()

	return nil
}

// save one object's part
func SaveObjectPart(objectName, bucketName, partID, uploadID, hash string, object io.ReadCloser, metadataM map[string][]string) (err error) {
	//rollback
	rollback := func(rollback func()) {
		if err != nil {
			rollback()
		}
	}

	var cache = make([]byte, 1024*1024)
	var data []byte
	// read the object
	for {
		n, err := object.Read(cache)
		if err != nil && err != io.EOF {
			return err
		}
		data = append(data, cache[:n]...)
		if err == io.EOF {
			break
		}
	}

	clusterID, err := allocator.GetClusterID()
	if err != nil {
		return err
	}
	bucketID := connection.MysqlMgr.MySQL.FindBucket(bucketName).BucketID
	if bucketID == "" {
		return fmt.Errorf("bucket doesn't exist")
	}
	name := bucketID + "-" + objectName
	objectTmp, ok := objectCache.objectCacheMap[name]
	if !ok {
		return fmt.Errorf("object cache error")
	}
	objectID := objectTmp.objectID
	partOid := allocator.AllocateObjectID(bucketID, clusterID)

	// write object's part
	err = connection.CephMgr.Ceph.WriteObject(connection.BucketData, partOid, data, 0)
	defer rollback(func() { rollbackSaveObject(partOid) })
	metadata, err := json.Marshal(&metadataM)
	if err != nil {
		return
	}
	partObjectName := uploadID + ":" + partID + ":" + objectID
	err = connection.MysqlMgr.MySQL.SavePartObjectTransaction(partObjectName, partOid, string(metadata))
	// concurrent upload
	partsCache.mutex.Lock()
	partsCache.partsCacheMap[name] = append(partsCache.partsCacheMap[name], partObjectName)
	partsCache.mutex.Unlock()
	return
}

func CompleteMultipartUpload(bucketName, objectName, uploadID string, partIDs []string) (err error) {
	bucketID := connection.MysqlMgr.MySQL.FindBucket(bucketName).BucketID
	name := bucketID + "-" + objectName

	objectCache.mutex.Lock()
	object, ok := objectCache.objectCacheMap[name]
	if !ok {
		return fmt.Errorf("objectID doesn't exist")
	}
	objectID := object.objectID
	metadata := object.metadata
	acl := object.acl
	delete(objectCache.objectCacheMap, name)
	objectCache.mutex.Unlock()
	err = connection.MysqlMgr.MySQL.SaveObjectTransaction(name, objectID, metadata, acl, true)
	if err != nil {
		return
	}

	// key partID
	// value partObjectID
	var parts = make(map[string]string)
	// check
	for _, id := range partIDs {
		part := uploadID + ":" + id + ":" + objectID
		tempPart, ok := partsCache.partsCacheMap[name]
		if !ok {
			return fmt.Errorf("object name error")
		}
		isExist := false
		for i, v := range tempPart {
			if v == part {
				isExist = true
				partsCache.partsCacheMap[name] = append(partsCache.partsCacheMap[name][:i],
					partsCache.partsCacheMap[name][i+1:]...)
				break
			}
		}
		if !isExist {
			return fmt.Errorf("part doesn't exist")
		}
		partObjectID := connection.MysqlMgr.MySQL.FindObject(part).ObjectID
		parts[id] = partObjectID
	}
	err = connection.MysqlMgr.MySQL.SaveObjectPartBatch(objectID, parts)
	if err != nil {
		return
	}
	partsCache.mutex.Lock()
	delete(partsCache.partsCacheMap, name)
	partsCache.mutex.Unlock()
	return nil
}

func AbortMultipartUpload(bucketName, objectName, uploadID string) error {
	bucketID := connection.MysqlMgr.MySQL.FindBucket(bucketName).BucketID
	name := bucketID + "-" + objectName

	objectCache.mutex.Lock()
	_, ok := objectCache.objectCacheMap[name]
	if !ok {
		return fmt.Errorf("objectID doesn't exist")
	}
	delete(objectCache.objectCacheMap, name)
	objectCache.mutex.Unlock()

	for _, partName := range partsCache.partsCacheMap[name] {
		partID := connection.MysqlMgr.MySQL.FindObject(partName).ObjectID
		go gc.WriteDeleteObjectChan(partID)
		go gc.WriteDeleteObjectDBChan(partName)
		go gc.WriteDeleteMetadataChan(partID + "-metadata")
	}

	partsCache.mutex.Lock()
	delete(partsCache.partsCacheMap, name)
	partsCache.mutex.Unlock()
	return nil
}

// read one object from ceph
func readOneObject(oid string) ([]byte, error) {
	var data []byte
	datacache := make([]byte, 1024*1024)
	for {
		n, err := connection.CephMgr.Ceph.ReadObject(connection.BucketData, oid, datacache, 0)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			break
		}
		data = append(data, datacache[:n]...)
	}
	return data, nil
}

func readMultipartObject(oid string) ([]byte, error) {
	objectParts := connection.MysqlMgr.MySQL.FindObjectPart(oid)
	var data []byte
	datacache := make([]byte, 1024*1024)
	for _, o := range objectParts {
		var partData []byte
		for {
			n, err := connection.CephMgr.Ceph.ReadObject(connection.BucketData, o.PartObjectID, datacache, 0)
			if err != nil {
				return nil, err
			}
			if n == 0 {
				break
			}
			partData = append(partData, datacache[:n]...)
		}
		data = append(data, partData...)
	}
	return data, nil
}
