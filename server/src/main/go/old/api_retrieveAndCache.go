package main

import (
	"context"
	"fmt"
	mathrand "math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (s *Server) RetrieveAndCache(c *gin.Context) {
	var email string
	var site string
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

	// cookie, err := c.Cookie("sspassman-auth")
	// if err != nil {
	// 	c.JSON(200, gin.H{
	// 		"message": "Cookie not found",
	// 	})
	// 	return
	// }

	cookie, err := c.Cookie("sspassman-auth")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Server Error - database connection",
		})
		return
	}
	rds := s.connectToAuthTokenCache()
	defer rds.Close()

	kv, err := s.retrieveFromAuthTokenCache(rds, email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Server Error - database connection",
		})
		return
	}
	if token, ok := kv[email]; !ok || token != cookie {
		c.JSON(http.StatusUnauthorized, gin.H{
			"msg": "Unauthorized",
		})
		return
	}

	/*
		Get uid and password from creds db and make sure it exists.
	*/
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		s.credsDbUsername,
		s.credsDbPassword,
		s.credsDbHost,
		s.credsDbPort,
		s.credsDbName)

	// Connect using pgx
	credsdbConn, err := pgx.Connect(context.Background(), url)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Server Error - database connection",
		})
		return
	}
	defer credsdbConn.Close(context.Background())

	userid, pwdCipher, err := getCredentials(credsdbConn, email, site)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"msg": fmt.Sprintf("Credentials not found for email: %s and site: %s", email, site),
		})
		return
	}

	// c.JSON(http.StatusOK, gin.H{
	// 	"msg": fmt.Sprintf("Credentials: username: %s, password: %s ", userid, pwd),
	// })

	/*
		Validate the email with a code
	*/
	// Generate the code
	code := ""
	for i := 0; i < 6; i++ {
		code += strconv.Itoa(mathrand.Intn(10))
	}
	fmt.Println("code:", code)

	// fmt.Sprintf("%s, %s", password, userid)

	// Connect to redis
	credsRds := s.connectToCredsCache()
	defer credsRds.Close()

	// store the code, email, site, username, password in redis
	key := email + ":" + site
	err = s.storeInCredsCache(context.Background(), credsRds, key, s.hashAndSalt([]byte(code)), pwdCipher, userid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Server Error - validation code generation",
		})
		return
	}

	if err = s.setRedisTTL(context.Background(), credsRds, key, 2*time.Minute); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Server Error - validation code generation",
		})
		return
	}

	/*
		Respond with the retrieved password after validation
	*/
	c.JSON(http.StatusOK, gin.H{
		"msg": "Code generated",
	})
}
