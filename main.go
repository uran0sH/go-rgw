package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-rgw/auth"
	"go-rgw/auth/jwt"
	"go-rgw/connection"
	"go-rgw/gc"
	"go-rgw/log"
	"go-rgw/router"
)

var defaultConfig = "./application.yml"

func main() {
	configFile := flag.String("config", defaultConfig, "configuration filename")
	config, err := readConfig(*configFile)
	if err != nil {
		fmt.Println(err)
	}
	err = log.Init(config.Log.Filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	mysql := connection.NewMySQL(config.Database.Username, config.Database.Password, config.Database.Address, config.Database.Name,
		"utf8mb4")
	err = mysql.Init()
	if err != nil {
		fmt.Println(err)
		return
	}
	connection.InitMySQLManager(mysql)
	ceph, err := connection.NewCeph()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = ceph.InitDefault()
	if err != nil {
		fmt.Println(err)
		return
	}
	connection.InitCephManager(ceph)
	gc.Init()
	var r *gin.Engine
	if config.Authorization == "jwt" {
		r = router.SetupRouter(&jwt.JWT{})
	} else {
		r = router.SetupRouter(auth.DefaultAuth{})
	}
	if err := r.Run(":8080"); err != nil {
		fmt.Println(err)
		return
	}
}
