package auth

import (
	"github.com/docker/distribution/registry/auth"
	"github.com/theonlyjohnny/rac/api/internal/storage"
)

type Authenticator interface {
	FilterAccessRequests(user *storage.User, requests []auth.Access) ([]auth.Access, error)
}

type authenticatorImpl struct{}

func NewAuthenticator() Authenticator {
	return &authenticatorImpl{}
}

func (au *authenticatorImpl) FilterAccessRequests(user *storage.User, requests []auth.Access) ([]auth.Access, error) {
	return requests, nil
}
