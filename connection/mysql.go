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

// table
type Object struct {
	Filename string `gorm:"primary_key"`
	ObjectID string
}

type User struct {
	Username string `gorm:"primary_key"`
	Password string
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
	fmt.Println("connect mysql successfully")
	db.AutoMigrate(&Object{})
	db.AutoMigrate(&User{})
	m.Database = db
	return nil
}

func (m *MySQL) Close() error {
	err := m.Database.Close()
	return err
}

func (m *MySQL) SaveObject(filename, oid string) {
	data := Object{Filename: filename, ObjectID: oid}
	m.Database.Create(&data)
}

func (m *MySQL) DeleteObject(filename, oid string) {
	m.Database.Delete(&Object{Filename: filename, ObjectID: oid})
}

func (m *MySQL) DeleteObjectByName(filename string) {
	m.Database.Where("filename = ?", filename).Delete(&Object{})
}

func (m *MySQL) FindObjectByName(filename string) (object Object) {
	m.Database.Where("filename = ?", filename).First(&object)
	return
}

func (m *MySQL) UpdateObject(filename, oid string) {
	m.Database.Model(&Object{}).Where("filename = ?", filename).Update("object_id", oid)
}

func (m *MySQL) SaveUser(username, password string) {
	user := User{Username: username, Password: password}
	m.Database.Create(&user)
}

func (m *MySQL) UpdateUsername(username, password string) {
	m.Database.Model(&User{}).Where("password = ?", password).Update("username", username)
}

func (m *MySQL) UpdatePassword(username, password string) {
	m.Database.Model(&User{}).Where("username = ?", username).Update("password", password)
}

func (m *MySQL) FindUser(username string) (u User) {
	m.Database.Where("username = ?", username).First(&u)
	return
}
