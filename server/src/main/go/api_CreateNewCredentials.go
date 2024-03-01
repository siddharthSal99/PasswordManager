package main

import "github.com/gin-gonic/gin"

func (s *Server) CreateNewCredentials(c *gin.Context) {
	// get the email, site, new password and validation code from the request

	// retrieve the cached validation code and verify they match

	// encrypt the password

	// add a new creds db entry with email, site and new encrypted password

	// return success
}
