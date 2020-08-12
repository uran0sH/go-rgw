package session

import (
	"encoding/json"
	"errors"
	"go-rgw/connection"
)

const (
	Read        = "READ"
	Write       = "WRITE"
	FullControl = "FULL_CONTROL"
	// default acl
	Private         = "PRIVATE"
	PublicRead      = "PUBLIC_READ"
	PublicReadWrite = "PUBLIC_READ_WRITE"
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

func NewAcl(userId, defaultAcl string, accessControlList []Grant) *Acl {
	acl := &Acl{
		UserId:            userId,
		DefaultAcl:        defaultAcl,
		AccessControlList: accessControlList,
	}
	return acl
}

func NewAccessControlList(grantee map[string][]string) []Grant {
	var accessControList []Grant
	for key, value := range grantee {
		for _, id := range value {
			accessControList = append(accessControList, Grant{UserId: id, Permission: key})
		}
	}
	return accessControList
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
	if acl.DefaultAcl == PublicReadWrite {
		return true, nil
	}
	for _, v := range acl.AccessControlList {
		if userId == v.UserId {
			if v.Permission == Write || v.Permission == FullControl {
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
	if acl.DefaultAcl == PublicReadWrite || acl.DefaultAcl == PublicRead {
		return true, nil
	}
	for _, v := range acl.AccessControlList {
		if userId == v.UserId {
			if v.Permission == Read || v.Permission == FullControl {
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
	if acl.DefaultAcl == PublicRead || acl.DefaultAcl == PublicReadWrite {
		return true, nil
	}
	for _, v := range acl.AccessControlList {
		if userId == v.UserId {
			if v.Permission == Read || v.Permission == FullControl {
				return true, nil
			} else {
				return false, nil
			}
		}
	}
	return false, nil
}
