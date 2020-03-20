package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/theonlyjohnny/rac/api/internal/storage"
)

func (a *API) postClaim(c *gin.Context) {
	var body struct {
		RepoName string `json:"repo_name"`
		UserID   string `json:"user_id"` //TODO auth
	}
	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := a.dao.ClaimRepo(body.RepoName, &storage.User{body.UserID}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{})
}
