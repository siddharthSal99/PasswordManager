package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) DeleteCredentials(c *gin.Context) {
	// get the email, site, and validation code from the request
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

	// delete the entry from creds db with that email and site
	err = s.deleteCredentialsFromDatabase(email, site)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Server error",
		})
		return
	}

	// return success
	c.JSON(http.StatusOK, gin.H{
		"msg": "successfully deleted credentials for " + email + ", " + site,
	})
}
