package session

import (
	"encoding/json"
	"errors"
	"go-rgw/connection"
)

const (
	read        = "READ"
	write       = "WRITE"
	fullControl = "FULL_CONTROL"
	// default acl
	private         = "PRIVATE"
	publicRead      = "PUBLIC_READ"
	publicReadWrite = "PUBLIC_READ_WRITE"
)

type Acl struct {
	UserId            string  `json:"owner"`
	DefaultAcl        string  `json:"default"`
	AccessControlList []Grant `json:"accessControlList"`
}

type Grant struct {
	UserId     string `json:"userId"`
	Permission string `json:"permission"`
}

func newAcl(userId, defaultAcl string) *Acl {
	acl := &Acl{
		UserId:     userId,
		DefaultAcl: defaultAcl,
	}
	return acl
}

func CouldPut(userId, bucketName string) (bool, error) {
	bucketId := connection.MysqlMgr.MySQL.FindBucket(bucketName).BucketID
	if bucketId == "" {
		return false, errors.New("the bucket doesn't exist")
	}
	bucketAcl := connection.MysqlMgr.MySQL.FindBukcetAcl(bucketId + "-acl").Acl
	if bucketAcl == "" {
		return false, errors.New("the bucket's acl doesn't exist")
	}

	var acl Acl
	err := json.Unmarshal([]byte(bucketAcl), &acl)
	if err != nil {
		return false, err
	}

	if userId == acl.UserId || userId == "root" {
		return true, nil
	}
	if acl.DefaultAcl == private || acl.DefaultAcl == publicRead {
		return false, nil
	}
	for _, v := range acl.AccessControlList {
		if userId == v.UserId {
			if v.Permission == write || v.Permission == fullControl {
				return true, nil
			} else {
				return false, nil
			}
		}
	}
	return false, nil
}

func CouldGet(userId, bucketName, objectName string) (bool, error) {
	bucketId := connection.MysqlMgr.MySQL.FindBucket(bucketName).BucketID
	if bucketId == "" {
		return false, errors.New("the bucket doesn't exist")
	}
	bucketAcl := connection.MysqlMgr.MySQL.FindBukcetAcl(bucketId + "-acl").Acl
	if bucketAcl == "" {
		return false, errors.New("the bucket's acl doesn't exist")
	}

	var acl Acl
	err := json.Unmarshal([]byte(bucketAcl), &acl)
	if err != nil {
		return false, err
	}
	if userId == acl.UserId || userId == "root" {
		return true, nil
	}
	if acl.DefaultAcl == private {
		return false, nil
	}
	for _, v := range acl.AccessControlList {
		if userId == v.UserId {
			if v.Permission == read || v.Permission == fullControl {
				ok, err := couldGetObject(userId, bucketId+"-"+objectName)
				if err != nil {
					return false, err
				}
				if !ok {
					return false, nil
				}
				return true, nil
			} else {
				return false, nil
			}
		}
	}
	return false, nil
}

func couldGetObject(userId, objectName string) (bool, error) {
	objectId := connection.MysqlMgr.MySQL.FindObject(objectName).ObjectID
	if objectId == "" {
		return false, errors.New("object doesn't exist")
	}
	objectAcl := connection.MysqlMgr.MySQL.FindObjectAcl(objectId + "-acl").Acl
	if objectAcl == "" {
		return false, errors.New("the object's acl doesn't exist")
	}

	var acl Acl
	err := json.Unmarshal([]byte(objectAcl), &acl)
	if err != nil {
		return false, err
	}

	if userId == acl.UserId || userId == "root" {
		return true, nil
	}
	if acl.DefaultAcl == private {
		return false, nil
	}
	for _, v := range acl.AccessControlList {
		if userId == v.UserId {
			if v.Permission == read || v.Permission == fullControl {
				return true, nil
			} else {
				return false, nil
			}
		}
	}
	return false, nil
}
