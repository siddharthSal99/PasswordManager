package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) CreateNewCredentials(c *gin.Context) {
	// get the email, site, new password and validation code from the request
	// Define a map to hold the request body
	var jsonMap map[string]string

	// Bind the JSON to the map
	if err := c.BindJSON(&jsonMap); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract the fields from the post body
	var email string
	var site string
	var userid string
	var password string
	var userCode string
	var ok bool

	if email, ok = jsonMap["email"]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no email provided in post body"})
	}

	if site, ok = jsonMap["site"]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no site provided in post body"})
	}

	if userid, ok = jsonMap["userid"]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no userid provided in post body"})
	}

	if password, ok = jsonMap["password"]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no password provided in post body"})
	}

	if userCode, ok = jsonMap["validationCode"]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no validation code provided in post body"})
	}

	// retrieve the cached validation code and verify they match
	rds := s.connectToValidationCodeCache()
	defer rds.Close()
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

	// encrypt the password

	pwdCipherText, err := s.encryptString(password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Server error",
		})
		return
	}

	// add a new creds db entry with email, site and new encrypted password
	err = s.insertCredentialsIntoCredsDb(email, site, userid, pwdCipherText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Server error",
		})
		return
	}
	// return success
	c.JSON(http.StatusOK, gin.H{
		"msg": "successfully added credentials for " + email + ", " + site,
	})
}
