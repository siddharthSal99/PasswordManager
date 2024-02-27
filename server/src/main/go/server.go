package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Server struct {
	credsDbHost     string
	credsDbPort     string
	credsDbUsername string
	credsDbPassword string
	credsDbName     string

	authDbHost     string
	authDbPort     string
	authDbUsername string
	authDbPassword string
	authDbName     string

	redisHost     string
	redisPort     string
	redisPassword string

	encryptionKey string
}

func (s *Server) Init() {
	godotenv.Load()
	s.credsDbHost = os.Getenv("credsDbHost")
	s.credsDbPort = os.Getenv("credsDbPort")
	s.credsDbUsername = os.Getenv("credsDbUsername")
	s.credsDbPassword = os.Getenv("credsDbPassword")
	s.credsDbName = os.Getenv("credsDbName")

	s.authDbHost = os.Getenv("authDbHost")
	s.authDbPort = os.Getenv("authDbPort")
	s.authDbUsername = os.Getenv("authDbUsername")
	s.authDbPassword = os.Getenv("authDbPassword")
	s.authDbName = os.Getenv("authDbName")

	s.redisHost = os.Getenv("redisHost")
	s.redisPort = os.Getenv("redisPort")
	s.redisPassword = os.Getenv("redisPassword")

	s.encryptionKey = os.Getenv("encryptionKey")
}

func main() {

	router := gin.Default()
	s := Server{}
	s.Init()

	/*
		Test Endpoint - Delete later
	*/
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	/*
		API route definitions
	*/
	router.GET("/password", s.AuthorizeAndCacheCredentials)
	router.GET("/password/validate", s.ValidateCodeAndRetrievePassword)
	router.PUT("/password", s.UpdatePassword)
	router.POST("/password", s.CreatePasswordEntry)
	router.DELETE("/password", s.DeletePassword)

	router.RunTLS(":443", "./certs/sspassman.com+4.pem", "./certs/sspassman.com+4-key.pem")

}
