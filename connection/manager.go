package connection

type MySQLManager struct {
	MySQL *MySQL
}

type CephManager struct {
	Ceph *Ceph
}

var MysqlMgr *MySQLManager
var CephMgr *CephManager

func InitMySQLManager(sql *MySQL) {
	MysqlMgr = &MySQLManager{
		MySQL: sql,
	}
}

func InitCephManager(ceph *Ceph) {
	CephMgr = &CephManager{
		Ceph: ceph,
	}
}
