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
type NameToID struct {
	Name string `gorm:"primary_key"`
	ID   string
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
	db.AutoMigrate(&NameToID{})
	db.AutoMigrate(&User{})
	m.Database = db
	return nil
}

func (m *MySQL) Close() error {
	err := m.Database.Close()
	return err
}

func (m *MySQL) SaveMap(name, id string) {
	data := NameToID{Name: name, ID: id}
	m.Database.Create(&data)
}

func (m *MySQL) DeleteMap(name, id string) {
	m.Database.Delete(&NameToID{Name: name, ID: id})
}

func (m *MySQL) DeleteMapByName(name string) {
	m.Database.Where("name = ?", name).Delete(&NameToID{})
}

func (m *MySQL) FindMapByName(name string) (result NameToID) {
	m.Database.Where("name = ?", name).First(&result)
	return
}

func (m *MySQL) UpdateMap(name, id string) {
	m.Database.Model(&NameToID{}).Where("name = ?", name).Update("id", id)
}

// SaveMapTransaction if you want to save both the mapping from data name to data ID and the mapping from metadata name
// to metadata ID, you should use this method.
func (m *MySQL) SaveMapTransaction(dataName, dataID, metaName, metaID string) error {
	tx := m.Database.Begin()
	data := NameToID{Name: dataName, ID: dataID}
	if err := tx.Create(&data).Error; err != nil {
		tx.Rollback()
		return err
	}
	metadata := NameToID{Name: metaName, ID: metaID}
	if err := tx.Create(&metadata).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (m *MySQL) UpdateMapTransaction(dataName, dataID, metaName, metaID string) error {
	tx := m.Database.Begin()
	if err := tx.Model(&NameToID{}).Where("name = ?", dataName).Update("id", dataID).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&NameToID{}).Where("name = ?", metaName).Update("id", metaID).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
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
