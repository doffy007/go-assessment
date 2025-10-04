package auth

import "rest-api/internal/user"

type DtoSignIn struct {
	Username string `json:"username" example:"johndoe"`
	Password string `json:"password" example:"secret123"`
}

type DtoRegister struct {
	Name     *string `json:"name" example:"John Doe"`
	Username string  `json:"username" example:"johndoe"`
	Email    string  `json:"email" example:"johndoe@mail.com"`
	Password string  `json:"password" example:"secret123"`
}

type DtoSignInResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6..."`
}

type DtoRegisterResponse struct {
	User user.Public `json:"user"`
}
