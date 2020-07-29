package connection

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"sync"
)

type MySQL struct {
	User     string
	Password string
	IPAddr   string
	Name     string
	Charset  string
	Database *gorm.DB
	mutex    sync.RWMutex
}

type User struct {
	UserID   string `gorm:"primary_key"`
	Username string
	Password string
}

// TODO add foreign key
type Bucket struct {
	BucketID   string `gorm:"primary_key"`
	BucketName string
}

// TODO add foreign key
// ObjectName = bucketID-object
// ObjectID = clustID.bucketID.ObjectUUID
// IsMultipart whether the object is a multipart upload
type Object struct {
	ObjectName  string `gorm:"primary_key"`
	ObjectID    string
	IsMultipart bool
}

// MetadataID is the objectID-acl
type ObjectACL struct {
	ACLID string `gorm:"primary_key"`
	ACL   string
}

type BucketACL struct {
	BucketID string `gorm:"primary_key"`
	ACL      string
}

// MetadataID is the objectID-metadata
type ObjectMetadata struct {
	MetadataID string `gorm:"primary_key"`
	Metadata   string
}

type ObjectPart struct {
	ObjectID string `gorm:"primary_key"`
	PartsID  string
}

func NewMySQL(user, password, ipAddr, name, charset string) *MySQL {
	mysql := &MySQL{
		User:     user,
		Password: password,
		IPAddr:   ipAddr,
		Name:     name,
		Charset:  charset,
		Database: nil,
	}
	return mysql
}

func (m *MySQL) Init() error {
	configStr := fmt.Sprintf("%s:%s@(%s)/%s?charset=%s&parseTime=True&loc=Local", m.User, m.Password, m.IPAddr,
		m.Name, m.Charset)
	db, err := gorm.Open("mysql", configStr)
	if err != nil {
		return err
	}
	// TODO use the log
	fmt.Println("connect mysql successfully")

	// Automigrate the database if these following tables are not in the database
	db.AutoMigrate(&Bucket{})
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Object{})
	db.AutoMigrate(&ObjectACL{})
	db.AutoMigrate(&ObjectMetadata{})
	db.AutoMigrate(&BucketACL{})

	m.Database = db
	return nil
}

func (m *MySQL) Close() error {
	err := m.Database.Close()
	return err
}

func (m *MySQL) CreateBucket(name, id string) {
	bucket := Bucket{id, name}
	m.Database.Create(&bucket)
}

func (m *MySQL) DeleteBucket(name string) {
	m.Database.Where("bucket_name = ?", name).Delete(&Bucket{})
}

func (m *MySQL) FindBucket(name string) (bucket Bucket) {
	m.Database.Where("bucket_name = ?", name).First(&bucket)
	return
}

func (m *MySQL) ListBuckets(uid string) []Bucket {
	var buckets []Bucket
	m.Database.Where("user_id = ?", uid).Find(&buckets)
	return buckets
}

func (m *MySQL) CreateUser(username, password, uid string) {
	user := User{UserID: uid, Username: username, Password: password}
	m.Database.Create(&user)
}

func (m *MySQL) UpdateUsername(uid, username string) {
	m.Database.Model(&User{}).Where("user_id = ?", uid).Update("username", username)
}

func (m *MySQL) UpdatePassword(username, password string) {
	m.Database.Model(&User{}).Where("username = ?", username).Update("password", password)
}

func (m *MySQL) FindUser(username string) User {
	var u User
	m.Database.Where("username = ?", username).First(&u)
	return u
}

func (m *MySQL) CreateObject(objectName string, oid string, isMultipart bool) error {
	object := Object{ObjectName: objectName, ObjectID: oid, IsMultipart: isMultipart}
	return m.Database.Create(&object).Error
}

func (m *MySQL) DeleteObject(objectName string) {
	m.Database.Where("object_name = ?", objectName).Delete(Object{})
}

func (m *MySQL) FindObject(objectName string) Object {
	var object Object
	m.Database.Where("object_name = ?", objectName).First(&object)
	return object
}

func (m *MySQL) UpdateObject(objectName, oid string) {
	m.Database.Model(&Object{}).Where("object_name = ?", objectName).Update("object_id", oid)
}

func (m *MySQL) DeleteObjectMetadata(metaID string) {
	m.Database.Where("metadata_id = ?", metaID).Delete(&ObjectMetadata{})
}

func (m *MySQL) DeleteObjectAcl(aclID string) {
	m.Database.Where("acl_id = ?", aclID).Delete(&ObjectACL{})
}

// save the acl, metadata and oid
func (m *MySQL) SaveObjectTransaction(objectName string, oid string, metadata string, acl string) (err error) {
	tx := m.Database.Begin()

	defer func() {
		if err != nil && tx != nil {
			tx.Rollback()
		}
	}()

	metadataID := oid + "-metadata"
	aclID := oid + "-acl"

	var tempObj Object
	if tx.Where("object_name = ?", objectName).First(&tempObj); tempObj == (Object{}) {
		object := Object{ObjectName: objectName, ObjectID: oid, IsMultipart: false}
		if err = tx.Create(&object).Error; err != nil {
			return
		}
	} else {
		// delete metadata's and acl's old version
		tempMetadata := tempObj.ObjectID + "-metadata"
		if err = tx.Where("metadata_id = ?", tempMetadata).Delete(&ObjectMetadata{}).Error; err != nil {
			return
		}
		tempACL := tempObj.ObjectID + "-acl"
		if err = tx.Where("acl_id = ?", tempACL).Delete(&ObjectACL{}).Error; err != nil {
			return
		}
		tempObj.ObjectID = oid
		if err = tx.Save(&tempObj).Error; err != nil {
			return
		}
	}

	objectMetadata := ObjectMetadata{MetadataID: metadataID, Metadata: metadata}
	if err = tx.Create(&objectMetadata).Error; err != nil {
		return
	}

	objectACL := ObjectACL{ACLID: aclID, ACL: acl}
	if err = tx.Create(&objectACL).Error; err != nil {
		return
	}

	tx.Commit()
	return nil
}

// save multipartObject && metadata
func (m *MySQL) SavePartObjectTransaction(partObjectName, partObjectID, metadata string) (err error) {
	tx := m.Database.Begin()

	defer func() {
		if err != nil && tx != nil {
			tx.Rollback()
		}
	}()

	partObject := Object{ObjectName: partObjectName, ObjectID: partObjectID, IsMultipart: false}
	if err = tx.Create(&partObject).Error; err != nil {
		return
	}
	metadataID := partObjectID + "-metadata"
	objectMetadata := ObjectMetadata{MetadataID: metadataID, Metadata: metadata}
	if err = tx.Create(&objectMetadata).Error; err != nil {
		return
	}

	tx.Commit()
	return nil
}

func (m *MySQL) SaveObjectPart(objectID string, partsID string) {
	objectPart := ObjectPart{ObjectID: objectID, PartsID: partsID}
	m.Database.Create(&objectPart)
}
