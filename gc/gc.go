package gc

import (
	"go-rgw/connection"
	"time"
)

var deleteObjectChan = make(chan string, 10)
var deleteObjectDBChan = make(chan string, 10)
var deleteMetadataChan = make(chan string, 10)
var deleteAclChan = make(chan string, 10)

func Init() {
	for i := 0; i < 3; i++ {
		go deleteObjectAsync()
		go deleteObjectDBAsync()
		go deleteMetadataAsync()
		go deleteAclAsync()
	}
}

func WriteDeleteObjectChan(objectID string) {
	for {
		select {
		case deleteObjectChan <- objectID:
			return
		default:
			time.Sleep(time.Second)
		}
	}
}

func deleteObjectAsync() {
	for {
		select {
		case objectID := <-deleteObjectChan:
			deleteObjectCeph(objectID)
		default:
			time.Sleep(time.Second)
		}
	}
}

func deleteObjectCeph(oid string) {
	for {
		err := connection.CephMgr.Ceph.DeleteObject(connection.BucketData, oid)
		if err == nil {
			break
		}
	}
}

func WriteDeleteObjectDBChan(objectName string) {
	for {
		select {
		case deleteObjectDBChan <- objectName:
			return
		default:
			time.Sleep(time.Second)
		}
	}
}

func deleteObjectDBAsync() {
	for {
		select {
		case objectName := <-deleteObjectDBChan:
			deleteObjectDB(objectName)
		default:
			time.Sleep(time.Second)
		}
	}
}

func deleteObjectDB(name string) {
	for {
		err := connection.MysqlMgr.MySQL.DeleteObject(name)
		if err == nil {
			break
		}
	}
}

func WriteDeleteMetadataChan(metadataID string) {
	for {
		select {
		case deleteMetadataChan <- metadataID:
			return
		default:
			time.Sleep(time.Second)
		}
	}
}

func deleteMetadataAsync() {
	for {
		select {
		case metadataID := <-deleteMetadataChan:
			deleteMetadata(metadataID)
		default:
			time.Sleep(time.Second)
		}
	}
}

func deleteMetadata(metadataID string) {
	for {
		err := connection.MysqlMgr.MySQL.DeleteObjectMetadata(metadataID)
		if err == nil {
			break
		}
	}
}

func WriteDeleteAclChan(aclID string) {
	for {
		select {
		case deleteAclChan <- aclID:
			return
		default:
			time.Sleep(time.Second)
		}
	}
}

func deleteAclAsync() {
	for {
		select {
		case aclID := <-deleteAclChan:
			deleteAcl(aclID)
		default:
			time.Sleep(time.Second)
		}
	}
}

func deleteAcl(aclID string) {
	for {
		err := connection.MysqlMgr.MySQL.DeleteObjectAcl(aclID)
		if err == nil {
			break
		}
	}
}
