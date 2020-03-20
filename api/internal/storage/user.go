package storage

type User struct {
	UserID string
	//idk
}

func (u *User) Equals(u2 *User) bool {
	return u2 != nil && u.UserID == u2.UserID
}
