package main

import (
	"fmt"
	"go-rgw/connection"
	"go-rgw/gc"
	"go-rgw/router"
)

func main() {
	mysql := connection.NewMySQL("root", "root", "118.31.64.83:3306", "ceph",
		"utf8mb4")
	err := mysql.Init()
	if err != nil {
		fmt.Println(err)
	}
	connection.InitMySQLManager(mysql)
	ceph, err := connection.NewCeph()
	if err != nil {
		fmt.Println(err)
	}
	err = ceph.InitDefault()
	if err != nil {
		fmt.Println(err)
	}
	connection.InitCephManager(ceph)
	gc.Init()
	r := router.SetupRouter()
	if err := r.Run(":8080"); err != nil {
		fmt.Println(err)
	}
}
