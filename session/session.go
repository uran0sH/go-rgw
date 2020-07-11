package session

import (
	"fmt"
	"go-rgw/allocator"
	"io"
	"math/rand"
)


func SaveObject(filename string, data []byte) {

}

func Save(filename string, src io.Reader) {
	oid := allocator.AllocateID()
	data := make([]byte, 100)
	var object []byte
	for {
		n, err := src.Read(data)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(n)
		object = append(object, data[:n]...)
		if n == 0 || err == nil || err == io.EOF{
			break
		}
	}
	mysqlIndex := rand.Intn(mysqlManager.num)
	isNew := mysqlManager.MySQLs[mysqlIndex].Save(filename, oid)
	if isNew == false {
		oid = mysqlManager.MySQLs[mysqlIndex].FindByName(filename).ObjectID
	}
	cephIndex := rand.Intn(cephManager.num)
	err := cephManager.Cephs[cephIndex].WriteObject("test-pool", oid, object, 0)
	fmt.Println(err)
}
