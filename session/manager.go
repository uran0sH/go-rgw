package session

import "go-rgw/connection"

type MySQLManager struct {
	MySQL *connection.MySQL
}

type CephManager struct {
	Ceph *connection.Ceph
}

var mysqlManager *MySQLManager
var cephManager *CephManager

func InitMySQLManager(sql *connection.MySQL) {
	mysqlManager = &MySQLManager{
		MySQL: sql,
	}
}

func InitCephManager(ceph *connection.Ceph) {
	cephManager = &CephManager{
		Ceph: ceph,
	}
}