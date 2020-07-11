package session

import "go-rgw/connection"

type MySQLManager struct {
	num int
	MySQLs map[int]*connection.MySQL
}

type CephManager struct {
	num int
	Cephs map[int]*connection.Ceph
}

var mysqlManager *MySQLManager
var cephManager *CephManager

func InitManager() {
	mysqlManager = &MySQLManager{
		num: 0,
		MySQLs: make(map[int]*connection.MySQL, 10),
	}
	cephManager = &CephManager{
		num: 0,
		Cephs: make(map[int]*connection.Ceph, 10),
	}
}
func AddMySQL(sql *connection.MySQL) {
	mysqlManager.MySQLs[mysqlManager.num] = sql
	mysqlManager.num++
}

func AddCeph(ceph *connection.Ceph) {
	cephManager.Cephs[cephManager.num] = ceph
	cephManager.num++
}
