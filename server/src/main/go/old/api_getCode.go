package main

import (
	"context"
	"fmt"
	mathrand "math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *Server) SendCode(c *gin.Context) {
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
	err := s.storeInCredsCache(context.Background(), credsRds, key, s.hashAndSalt([]byte(code)), pwdCipher, userid)
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
