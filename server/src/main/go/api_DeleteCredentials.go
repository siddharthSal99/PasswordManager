package main

import "github.com/gin-gonic/gin"

func (s *Server) DeleteCredentials(c *gin.Context) {
	// get the email, site, and validation code from the request

	// retrieve the cached validation code and verify they match

	// delete the entry from creds db with that email and site

	// return success
}
