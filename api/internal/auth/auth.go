package auth

import (
	"fmt"

	"github.com/docker/distribution/registry/auth"
	uuid "github.com/satori/go.uuid"
	"github.com/theonlyjohnny/rac/api/internal/storage"
)

type ProvidedCredentials struct {
	Username string
	Password string

	RefreshToken string
}

type Authenticator interface {
	FilterAccessRequests(user *storage.User, requests []auth.Access) ([]auth.Access, error)
	AuthenticateCredentials(providedCreds ProvidedCredentials) (*storage.User, error)

	CreateRefreshToken(user *storage.User) (string, error)
}

type authenticatorImpl struct {
	dao storage.DAO

	refreshCache map[string]*storage.User
}

func NewAuthenticator(dao storage.DAO) Authenticator {
	return &authenticatorImpl{
		dao:          dao,
		refreshCache: make(map[string]*storage.User, 0),
	}
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

func (au *authenticatorImpl) AuthenticateCredentials(providedCreds ProvidedCredentials) (*storage.User, error) {

	if providedCreds.Username != "" && providedCreds.Password != "" {
		if providedCreds.Username == "admin" && providedCreds.Password == "password" {
			return &storage.User{
					UserID: "god",
				},
				nil
		}
	} else if providedCreds.RefreshToken != "" {
		return au.getUserByRefresh(providedCreds.RefreshToken)
	}
	return nil, fmt.Errorf("account not found")
}

func (au *authenticatorImpl) CreateRefreshToken(user *storage.User) (string, error) {
	if user == nil {
		return "", fmt.Errorf("nil user to CreateRefreshToken")
	}
	refresh := uuid.NewV4().String()
	err := au.rememberRefresh(refresh, user)
	return refresh, err
}
func (au *authenticatorImpl) getUserByRefresh(refresh string) (*storage.User, error) {
	val, ok := au.refreshCache[refresh]
	if !ok {
		return nil, fmt.Errorf("refresh token not found")
	}
	return val, nil
}

func (au *authenticatorImpl) rememberRefresh(refresh string, user *storage.User) error {

	_, ok := au.refreshCache[refresh]
	if ok {
		return fmt.Errorf("refresh already claimed")
	}

	au.refreshCache[refresh] = user
	return nil
}
