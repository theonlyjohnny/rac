package api

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/docker/distribution/registry/auth"
	"github.com/gin-gonic/gin"
	"github.com/theonlyjohnny/rac/api/internal/storage"
)

var typeRegexp = regexp.MustCompile(`^([a-z0-9]+)(\([a-z0-9]+\))?$`)

func (a *API) getAuth(c *gin.Context) {
	user := &storage.User{"py_owner"}

	q := c.Request.URL.Query()
	accessRequests := resolveScopeSpecifiers(q["scope"])

	permitted, err := a.auth.FilterAccessRequests(user, accessRequests)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	fmt.Printf("user %s got access filtered: %s -> %s \n", user, accessRequests, permitted)
	token, err := a.token.CreateTokenForAcess(permitted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

//Stolen from https://github.com/docker/distribution/contrib/token-server/token.go
func resolveScopeSpecifiers(scopeSpecs []string) []auth.Access {
	requestedAccessSet := make(map[auth.Access]struct{}, 2*len(scopeSpecs))

	for _, scopeSpecifier := range scopeSpecs {
		// There should be 3 parts, separated by a `:` character.
		parts := strings.SplitN(scopeSpecifier, ":", 3)

		if len(parts) != 3 {
			fmt.Printf("ignoring unsupported scope format %s \n", scopeSpecifier)
			continue
		}

		resourceType, resourceName, actions := parts[0], parts[1], parts[2]

		resourceType, resourceClass := splitResourceClass(resourceType)
		if resourceType == "" {
			continue
		}

		// Actions should be a comma-separated list of actions.
		for _, action := range strings.Split(actions, ",") {
			requestedAccess := auth.Access{
				Resource: auth.Resource{
					Type:  resourceType,
					Class: resourceClass,
					Name:  resourceName,
				},
				Action: action,
			}

			// Add this access to the requested access set.
			requestedAccessSet[requestedAccess] = struct{}{}
		}
	}

	requestedAccessList := make([]auth.Access, 0, len(requestedAccessSet))
	for requestedAccess := range requestedAccessSet {
		requestedAccessList = append(requestedAccessList, requestedAccess)
	}

	return requestedAccessList
}

func splitResourceClass(t string) (string, string) {
	matches := typeRegexp.FindStringSubmatch(t)
	if len(matches) < 2 {
		return "", ""
	}
	if len(matches) == 2 || len(matches[2]) < 2 {
		return matches[1], ""
	}
	return matches[1], matches[2][1 : len(matches[2])-1]
}
