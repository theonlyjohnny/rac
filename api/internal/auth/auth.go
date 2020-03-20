package auth

import (
	"fmt"

	"github.com/docker/distribution/registry/auth"
	"github.com/theonlyjohnny/rac/api/internal/storage"
)

type Authenticator interface {
	FilterAccessRequests(user *storage.User, requests []auth.Access) ([]auth.Access, error)
}

type authenticatorImpl struct {
	dao storage.DAO
}

func NewAuthenticator(dao storage.DAO) Authenticator {
	return &authenticatorImpl{dao}
}

func (au *authenticatorImpl) FilterAccessRequests(user *storage.User, requests []auth.Access) ([]auth.Access, error) {
	if user == nil {
		return au.filterAnonymousAccessRequests(requests)
	}
	out := []auth.Access{}

	for _, e := range requests {
		if e.Action == "push" {
			name := e.Resource.Name
			repo, err := au.dao.GetRepo(name)
			if err != nil {
				fmt.Printf("Unable to get repo %s so not allowing push access -- %s\n", name, err.Error())
			} else if repo == nil || repo.Owner == nil || user.Equals(repo.Owner) {
				//if repo is new or
				//if repo is unclaimed or
				//if repo is claimed, and authenticated user matches:
				//allowed to push
				var msg string
				if repo == nil {
					msg = "repo is new"
				} else if repo.Owner == nil {
					msg = "repo is unclaimed"
				} else if user.Equals(repo.Owner) {
					msg = "user is owner"
				}

				fmt.Printf("allowing %s push on %s because %s \n", user, name, msg)
				out = append(out, e)
			}
		} else {
			out = append(out, e)
		}
	}

	fmt.Printf("filtered %s to %s \n", requests, out)

	return out, nil
}

func (au *authenticatorImpl) filterAnonymousAccessRequests(requests []auth.Access) ([]auth.Access, error) {
	//for now unauthenticated users can pull only, but can pull anything
	out := []auth.Access{}

	for _, e := range requests {
		if e.Action != "push" {
			out = append(out, e)
		}
	}

	fmt.Printf("filtered %s to %s \n", requests, out)
	return out, nil
}
