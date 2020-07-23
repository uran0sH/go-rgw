package auth

import "github.com/gin-gonic/gin"

type Auth interface {
	// if the authentication needs users to login, it must implement this funciton.
	Login(c *gin.Context)
	// create the user in the database
	CreateUser(c *gin.Context)
	// auth
	Auth() func(*gin.Context)
}

type DefaultAuth struct{}

func (d DefaultAuth) Login(c *gin.Context) {
}

func (d DefaultAuth) CreateUser(c *gin.Context) {
}

func (d DefaultAuth) Auth() func(*gin.Context) {
	return func(c *gin.Context) {
	}
}
