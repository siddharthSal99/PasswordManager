package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) ValidateCode(c *gin.Context) {

	var code string
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

	if val, ok := c.GetQuery("code"); ok {
		code = val
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "No code provided",
		})
		return
	}

	rds := connectToRedis(fmt.Sprintf("%s:%s", s.redisHost, s.redisPort), s.redisPassword, 0)
	defer rds.Close()

	key := email + ":" + site
	kv, err := retrieveRedis(context.Background(), rds, key)
	hashedCode := kv["validationCode"]
	storedPwdCipher := kv["pwdCipher"]
	userid := kv["userid"]

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Server Error - code validation",
		})
		return
	}

	if checkPassword(hashedCode, code) {
		password, err := decryptString(storedPwdCipher, s.encryptionKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg": "Server Error - decryption",
			})
			return
		}
		// TODO: in the retrieveRedis function, return the map itself and then get the fields you need from the map.
		// Then you can store the userid in there too
		c.JSON(http.StatusOK, gin.H{
			"msg": fmt.Sprintf("Credentials: username: %s, password: %s, code: %s", userid, password, hashedCode),
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"msg": "Invalid code",
		})
	}

}
