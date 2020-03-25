package api

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/docker/distribution/registry/auth"
	"github.com/gin-gonic/gin"
	racAuth "github.com/theonlyjohnny/rac/api/internal/auth"
)

var typeRegexp = regexp.MustCompile(`^([a-z0-9]+)(\([a-z0-9]+\))?$`)

type doAuthRequest struct {
	Account string ` form:"account"`
	// ClientID     string `form:"client_id"`
	OfflineToken bool `form:"offline_token"`
	// Service      string
	Scope []string `form:"scope"`
}

func (a *API) doAuth(c *gin.Context) {

	body := doAuthRequest{}
	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("doAuth %#v \n", body)

	providedCredsI, _ := c.Get(authedUserCtxKey)
	providedCreds, ok := providedCredsI.(racAuth.ProvidedCredentials)
	fmt.Printf("ProvidedCredentials:%#v\n", providedCreds)
	if !ok {
		c.Header("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\"", "http://rac.api:8090/auth"))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "must provide Basic credentials or refresh_token"})
		return
	}

	if providedCreds.Username != body.Account {
		c.Header("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\"", "http://rac.api:8090/auth"))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "username and account don't match"})
		return
	}

	user, err := a.auth.AuthenticateCredentials(providedCreds)
	fmt.Printf("authenticated to user: %s, err :%s \n", user, err)
	if err != nil || user == nil {
		//TODO make docker client show proper "invalid credentials" message here
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	var (
		refresh   string
		permitted []auth.Access
	)

	if len(body.Scope) > 0 {
		accessRequests := resolveScopeSpecifiers(body.Scope)
		permitted, err = a.auth.FilterAccessRequests(user, accessRequests)
		fmt.Printf("user %s got access filtered: %s -> %s \n", user, accessRequests, permitted)
	} else {
		fmt.Printf("user %s is logging in \n", user)
		//docker login
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	if body.OfflineToken {
		refresh, err = a.auth.CreateRefreshToken(user)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	token, err := a.token.CreateTokenForAcess(permitted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	resp := gin.H{
		// support older clients by specifying both token and access_token
		"token":        token,
		"access_token": token,
	}

	if refresh != "" {
		resp["refresh_token"] = refresh
	}

	fmt.Printf("resp: %#v \n", resp)
	c.JSON(http.StatusOK, resp)
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
