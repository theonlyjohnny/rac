package storage

import "fmt"

type DAO interface {
	GetRepo(name string) (*Repo, error)
	ClaimRepo(name string, owner *User) error
	SaveRepo(repo *Repo) error
}

type daoImpl struct {
	repos map[string]*Repo
}

func NewDAO() DAO {
	return &daoImpl{
		repos: make(map[string]*Repo),
	}
}

func (d *daoImpl) GetRepo(name string) (*Repo, error) {
	r, _ := d.repos[name]
	return r, nil
}

func (d *daoImpl) ClaimRepo(name string, owner *User) error {
	r, ok := d.repos[name]
	if !ok {
		return fmt.Errorf("unknown repo %s", name)
	}

	if r.Owner != nil {
		return fmt.Errorf("repo %s is already claimed", name)
	}

	r.Owner = owner
	d.repos[name] = r
	return nil
}

func (d *daoImpl) SaveRepo(repo *Repo) error {
	d.repos[repo.Name] = repo
	return nil
}
