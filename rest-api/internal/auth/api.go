package auth

import (
	"errors"
	"net/http"
	"rest-api/internal/api"
	"rest-api/internal/user"

	"github.com/gin-gonic/gin"
)

// SignIn godoc
// @Summary User login
// @Description Authenticate user with username and password, returns JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body auth.DtoSignIn true "Login request payload"
// @Success 200 {object} auth.DtoSignInResponse "JWT token returned"
// @Failure 400 {object} map[string]string "Invalid request payload"
// @Failure 401 {object} map[string]string "Invalid password"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /signin [post]
func SignIn(ctx *gin.Context) {
	var req DtoSignIn

	if err := ctx.ShouldBindJSON(&req); err != nil {
		api.Abort(ctx, http.StatusBadRequest, err, api.ErrPayload)
		return
	}

	token, err := Service.SignIn(req.Username, req.Password)
	if err != nil {
		switch err.Error() {
		case "user not found":
			api.Abort(ctx, http.StatusNotFound, err, errors.New("user not found"))
		case "invalid password":
			api.Abort(ctx, http.StatusUnauthorized, err, errors.New("invalid password"))
		default:
			api.Abort(ctx, http.StatusInternalServerError, err, errors.New("failed to sign in"))
		}
		return
	}

	ctx.JSON(http.StatusOK, DtoSignInResponse{Token: token})
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user in the system
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body auth.DtoRegister true "Register request payload"
// @Success 201 {object} auth.DtoRegisterResponse "User created successfully"
// @Failure 400 {object} map[string]string "Invalid request payload"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /register [post]
func Register(ctx *gin.Context) {
	var req DtoRegister

	if err := ctx.ShouldBindJSON(&req); err != nil {
		api.Abort(ctx, http.StatusBadRequest, err, api.ErrPayload)
		return
	}

	usr := &user.User{
		Name:     req.Name,
		Username: req.Username,
		Email:    req.Email,
		Password: &req.Password,
	}

	newUser, err := Service.Register(usr)
	if err != nil {
		api.Abort(ctx, http.StatusInternalServerError, err, errors.New("failed to register user"))
		return
	}

	ctx.JSON(http.StatusCreated, DtoRegisterResponse{
		User: *user.ToPublic(newUser),
	})
}
