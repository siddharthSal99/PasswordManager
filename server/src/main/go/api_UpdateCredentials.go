package main

import "github.com/gin-gonic/gin"

func (s *Server) UpdateCredentials(c *gin.Context) {
	// get the email, site, new password and validation code from the request

	// retrieve the cached validation code and verify they match

	// encrypt the password

	// update the creds db entry with email and site to new encrypted password

	// return success
}
