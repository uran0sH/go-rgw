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

type FilenameToID struct {
	Filename string `gorm:"primary_key"`
	ObjectID string
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
	m.Database = db
	return nil
}

func (m *MySQL) Close() error {
	err := m.Database.Close()
	return err
}

func (m *MySQL) Save(filename, oid string) {
	data := FilenameToID{Filename: filename, ObjectID: oid}
	m.Database.Create(&data)
}

func (m *MySQL) Delete(filename, oid string) {
	m.Database.Delete(&FilenameToID{Filename: filename, ObjectID: oid})
}

func (m *MySQL) DeleteByName(filename string) {
	m.Database.Where("filename = ?", filename).Delete(&FilenameToID{})
}

func (m *MySQL) FindByName(filename string) (f FilenameToID) {
	m.Database.Where("filename = ?", filename).First(&f)
	return
}

func (m *MySQL) Update(filename, oid string) {
	m.Database.Model(&FilenameToID{}).Where("filename = ?", filename).Update("object_id", oid)
}
