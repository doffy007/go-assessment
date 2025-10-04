package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"rest-api/internal/api"
	"rest-api/internal/util"

	apiparams "rest-api/internal/params"

	"github.com/gin-gonic/gin"
)

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user in the system
// @Tags User
// @Accept json
// @Produce json
// @Param request body user.User true "User data"
// @Success 200 {object} user.User
// @Failure 400 {object} map[string]string "Invalid request payload"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users [post]
func CreateUser(ctx *gin.Context) {
	req := &User{}

	if err := ctx.ShouldBindJSON(req); err != nil {
		api.Abort(ctx, http.StatusBadRequest, err, api.ErrPayload)
		return
	}

	user, err := Service.CreateUser(req)
	if err != nil {
		api.Abort(ctx, http.StatusInternalServerError, err, errors.New("failed to create user"))
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get user details by ID, or the logged-in user if ID not provided
// @Tags Admin
// @Produce json
// @Param userId path int false "User ID"
// @Success 200 {object} user.User
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /admin/users/{userId} [get]
func GetUser(ctx *gin.Context) {
	var err error

	userID, _ := api.GetUint64Param(ctx, "userId", false)
	if userID == 0 {
		userID, err = api.GetUserID(ctx, true)
		if err != nil {
			return
		}
	}

	resp, err := Service.GetUser(userID)
	if err != nil {
		api.Abort(ctx, http.StatusInternalServerError, err, errors.New("failed to get user"))
		return
	}

	if resp == nil {
		api.Abort(ctx, http.StatusNotFound, api.ErrNotFound, nil)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetUsers godoc
// @Summary Get list of users
// @Description Get list of users with optional filtering
// @Tags Admin
// @Produce json
// @Success 200 {array} user.User
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /admin/users [get]
func GetUsers(ctx *gin.Context) {
	resp, err := Service.GetUsers(apiparams.GetParams(ctx))
	if err != nil {
		api.Abort(ctx, http.StatusInternalServerError, err, errors.New("failed to get users"))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateUser godoc
// @Summary Update user
// @Description Update fields of a user
// @Tags User
// @Accept json
// @Produce json
// @Param userId path int false "User ID"
// @Param request body user.User true "Updated user fields"
// @Success 200 {object} user.User
// @Failure 400 {object} map[string]string "Invalid request payload"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /users/{userId} [put]
func UpdateUser(ctx *gin.Context) {
	b, err := ctx.GetRawData()
	if err != nil {
		api.Abort(ctx, http.StatusBadRequest, err, api.ErrPayload)
		return
	}
	updatedFields := util.GetUpdatedJSONFields(b)

	req := &User{}
	if err = json.Unmarshal(b, req); err != nil {
		api.Abort(ctx, http.StatusBadRequest, err, api.ErrPayload)
		return
	}

	req.ID, _ = api.GetUint64Param(ctx, "userId", false)

	if req.ID == 0 {
		req.ID, err = api.GetUserID(ctx, true)
		if err != nil {
			return
		}
	}

	resp, err := Service.UpdateUser(req, updatedFields)
	if err != nil {
		api.Abort(ctx, http.StatusInternalServerError, err, errors.New("failed to update user"))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete a user by ID
// @Tags User
// @Param userId path int true "User ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string "Invalid user ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /users/{userId} [delete]
func DeleteUser(ctx *gin.Context) {
	userID, err := api.GetUint64Param(ctx, "userId", true)
	if err != nil {
		api.Abort(ctx, http.StatusBadRequest, err, errors.New("invalid user ID"))
		return
	}

	if success, err := Service.DeleteUser(userID); err != nil {
		if err == api.ErrNotFound {
			api.Abort(ctx, http.StatusNotFound, err, nil)
		} else {
			api.Abort(ctx, http.StatusInternalServerError, err, errors.New("failed to delete user"))
		}
		return
	} else if !success {
		api.Abort(ctx, http.StatusInternalServerError, nil, errors.New("failed to delete user"))
		return
	}

	ctx.Status(http.StatusNoContent)
}

// GetMe godoc
// @Summary Get current user
// @Description Get logged-in user info based on JWT token
// @Tags User
// @Produce json
// @Success 200 {object} user.Public "Current user info"
// @Failure 401 {object} post.UnauthorizedResponse "Unauthorized"
// @Failure 500 {object} post.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users/me [get]
func GetMe(ctx *gin.Context) {
	token, exists := ctx.Get(api.TokenKey)
	if !exists {
		api.Abort(ctx, http.StatusUnauthorized, errors.New("missing token"), nil)
		return
	}

	user, err := Service.GetUserByToken(token.(string))
	if err != nil {
		api.Abort(ctx, http.StatusInternalServerError, err, errors.New("failed to get user"))
		return
	}

	if user == nil {
		api.Abort(ctx, http.StatusUnauthorized, errors.New("invalid token or user not found"), nil)
		return
	}

	ctx.JSON(http.StatusOK, user)
}
