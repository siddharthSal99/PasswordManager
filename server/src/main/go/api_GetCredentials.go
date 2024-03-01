package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) GetCredentials(c *gin.Context) {
	// get the email, site, and validation code from the request
	/*
		Extract the query parameters for email and site
	*/
	var email string
	var site string
	var userCode string
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

	if val, ok := c.GetQuery("code"); ok {
		userCode = val
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "No validation code provided",
		})
		return
	}

	// retrieve the cached validation code and verify they match
	rds := s.connectToValidationCodeCache()
	kv, err := s.retrieveValidationCode(rds, email, site)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Server error",
		})
		return
	}

	var validationCodeHash string
	if val, ok := kv["validationCode"]; ok {
		validationCodeHash = val
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "No code found - please try again",
		})
		return
	}

	if !s.checkPassword(validationCodeHash, userCode) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"msg": "Invalid code",
		})
		return
	}
	// retrieve the credentials from the creds db

	userid, pwdCipherText, err := s.getCredentialsFromDatabase(email, site)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Server error",
		})
		return
	}
	// decrypt the password
	passwordPlainText, err := s.decryptString(pwdCipherText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Server error",
		})
		return
	}

	// return the credentials
	c.JSON(http.StatusOK, gin.H{
		"msg": gin.H{
			"userid":   userid,
			"password": passwordPlainText,
		},
	})
}
