package auth

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/docker/distribution/registry/auth"
)

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

var typeRegexp = regexp.MustCompile(`^([a-z0-9]+)(\([a-z0-9]+\))?$`)

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
