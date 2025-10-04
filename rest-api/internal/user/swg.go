package user

type Public struct {
	ID       uint64  `json:"id,string"`
	Name     *string `json:"name,omitempty"`
	Username string  `json:"username"`
	Email    string  `json:"email"`
}

func ToPublic(u *User) *Public {
	if u == nil {
		return nil
	}
	return &Public{
		ID:       u.ID,
		Name:     u.Name,
		Username: u.Username,
		Email:    u.Email,
	}
}
