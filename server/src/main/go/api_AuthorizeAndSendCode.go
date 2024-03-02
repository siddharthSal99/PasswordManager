package main

import (
	"context"
	mathrand "math/rand"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (s *Server) createAndStoreCode(email string, site string) (string, error) {
	// Generate the code
	code := ""
	for i := 0; i < 6; i++ {
		code += strconv.Itoa(mathrand.Intn(10))
	}

	// Connect to redis
	credsRds := s.connectToValidationCodeCache()
	defer credsRds.Close()

	// store the code, email, site in redis
	key := email + ":" + site
	saltedCode, err := s.hashAndSalt([]byte(code))
	if err != nil {
		return "", err
	}
	err = s.storeValidationCode(context.Background(), credsRds, key, saltedCode)
	if err != nil {
		return "", err
	}
	return code, nil
}

func (s *Server) AuthorizeAndSendCode(c *gin.Context) {
	var email string
	var site string
	/*
		Extract the query parameters for email and site
	*/
	if val, ok := c.GetQuery("email"); ok {
		email = val
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "No email provided",
		})
		return
	}

	if val, ok := c.GetQuery("site"); ok {
		site = val
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "No site provided",
		})
		return
	}

	// Check if email is in auth db
	if isAuth, err := s.isAuthorized(email); !isAuth || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"msg": "this email is not authorized to access this resource",
		})
	} else {
		code, err := s.createAndStoreCode(email, site)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg": "Server Error:" + err.Error(),
			})
			return
		}
		err = s.sendCode(code, email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg": "Server Error:" + err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"msg": "code sent",
		})
	}
}
