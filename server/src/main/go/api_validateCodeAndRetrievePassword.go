package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func (s *Server) validate(rds *redis.Client, code string, key string) (bool, error) {

	kv, err := s.retrieveRedis(context.Background(), rds, key)
	hashedCode := kv["validationCode"]
	if err != nil {
		return false, err
	}

	return s.checkPassword(hashedCode, code), nil

}

func (s *Server) ValidateCodeAndRetrievePassword(c *gin.Context) {

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

	key := email + ":" + site
	rds := s.connectToRedis()
	defer rds.Close()

	if validated, err := s.validate(rds, code, key); err != nil {
		if validated {
			userid, password, err := s.retrieveCredentialsFromRedis(rds, key)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": "Server Error - retrieving credentials",
				})
				return
			}
			c.JSON(http.StatusUnauthorized, gin.H{
				"msg": gin.H{
					"userid":   userid,
					"password": password,
				},
			})
			return
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"msg": "Incorrect code provided",
			})
			return
		}
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Error retrieving credentials",
		})
		return
	}

}
