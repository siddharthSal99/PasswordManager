package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (s *Server) ValidateAndCommitNewCredentials(c *gin.Context) {

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
	rds := s.connectToCredsCache()
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
			url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
				s.credsDbUsername,
				s.credsDbPassword,
				s.credsDbHost,
				s.credsDbPort,
				s.credsDbName)

			credsdbConn, err := pgx.Connect(context.Background(), url)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": "Server Error - database connection",
				})
				return
			}
			defer credsdbConn.Close(context.Background())
			pwdCipher, err := encryptString(password, s.encryptionKey)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": "Server Error - encryption error",
				})
				return
			}
			err = insertCredentialsIntoCredsDb(credsdbConn, email, site, userid, pwdCipher)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": "Server Error - database error",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"msg": "committed credentials",
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
