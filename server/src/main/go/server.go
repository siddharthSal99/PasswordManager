package main

import (
	"net/http"
	"os"
	"time"

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

	validationCacheHost     string
	validationCachePort     string
	validationCachePassword string

	authTokenCacheHost     string
	authTokenCachePort     string
	authTokenCachePassword string

	sourceEmailUsername string
	sourceEmailHost     string
	sourceEmailPassword string

	encryptionKey          string
	validationCodeDuration time.Duration
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

	s.validationCacheHost = os.Getenv("validationCacheHost")
	s.validationCachePort = os.Getenv("validationCachePort")
	s.validationCachePassword = os.Getenv("validationCachePassword")

	s.authTokenCacheHost = os.Getenv("authTokenCacheHost")
	s.authTokenCachePort = os.Getenv("authTokenCachePort")
	s.authTokenCachePassword = os.Getenv("authTokenCachePassword")

	s.sourceEmailUsername = os.Getenv("sourceEmailUsername")
	s.sourceEmailPassword = os.Getenv("sourceEmailPassword")
	s.sourceEmailHost = os.Getenv("sourceEmailHost")

	s.encryptionKey = os.Getenv("encryptionKey")
	s.validationCodeDuration = 2 * time.Minute
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
	// router.GET("/password", s.Authorize)
	// router.GET("/password/validate", s.ValidateCodeAndRetrievePassword)
	// router.PUT("/password", s.UpdatePassword)
	// router.POST("/password", s.CreatePasswordEntry)
	// router.POST("/code", s.SendCode)
	// router.DELETE("/password", s.DeletePassword)

	/*
		These CRUD endpoints start by checking if the auth cookie from the browser matches the one in auth token cache
		if not, then check if the code sent along with the request is valid,
		then perform the action if the user is authenticated

	*/
	router.GET("/credentials", s.GetCredentials)
	router.POST("/credentials", s.CreateNewCredentials)
	router.PUT("/credentials", s.UpdateCredentials)
	router.DELETE("/credentials", s.DeleteCredentials)

	/*
		This endpoint checks if the provided email is in the 'authorized' database,
		then generates a validation code
	*/
	router.GET("/validationCode", s.AuthorizeAndSendCode)

	router.RunTLS(":443", "./certs/sspassman.com+4.pem", "./certs/sspassman.com+4-key.pem")

}
