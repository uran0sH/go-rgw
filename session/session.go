package session

import (
	"fmt"
	"go-rgw/allocator"
)


func SaveObject(filename string, data []byte) error {
	oid := mysqlManager.MySQL.FindByName(filename).ObjectID
	if oid == "" {
		oid = allocator.AllocateID()
		mysqlManager.MySQL.Save(filename, oid)
	}
	err := cephManager.Ceph.WriteObject("test-pool", oid, data, 0)
	fmt.Println(err)
	return err
}

//func Save(filename string, src io.Reader) {
//	oid := allocator.AllocateID()
//	data := make([]byte, 100)
//	var object []byte
//	for {
//		n, err := src.Read(data)
//		if err != nil {
//			fmt.Println(err)
//		}
//		fmt.Println(n)
//		object = append(object, data[:n]...)
//		if n == 0 || err == nil || err == io.EOF{
//			break
//		}
//	}
//	isNew := mysqlManager.MySQL.Save(filename, oid)
//	if isNew == false {
//		oid = mysqlManager.MySQL.FindByName(filename).ObjectID
//	}
//	err := cephManager.Ceph.WriteObject("test-pool", oid, object, 0)
//	fmt.Println(err)
//}

func GetObject(filename string) ([]byte, error){
	oid := mysqlManager.MySQL.FindByName(filename).ObjectID
	if oid == "" {
		return nil, fmt.Errorf("the filename doesn't exist")
	}
	data := make([]byte, 100)
	n, err := cephManager.Ceph.ReadObject("test-pool", oid, data, 0)
	return data[:n], err
}