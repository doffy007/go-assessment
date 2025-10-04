package role

import (
	"errors"
	"net/http"
	"rest-api/internal/api"

	"github.com/gin-gonic/gin"
)

// AssignRole godoc
// @Summary Assign a role to a user
// @Description Assign a specific role (e.g. Writter) to a user by ID
// @Tags Admin
// @Param userId path int true "User ID"
// @Security BearerAuth
// @Success 200 "Role assigned successfully"
// @Failure 400 {object} role.ErrorRoleResponse "Invalid role"
// @Failure 500 {object} role.ErrorRoleResponse "Failed to assign role"
// @Param request body role.AssignRoleRequest true "Role to assign"
// @Router /admin/users/{userId}/assign [post]
func AssignRole(ctx *gin.Context) {
	var req AssignRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		api.Abort(ctx, http.StatusBadRequest, err, nil)
		return
	}

	switch req.Role {
	case SuperAdmin, Writter:
	default:
		api.Abort(ctx, http.StatusBadRequest, errors.New("invalid role"), nil)
		return
	}

	userID, err := api.GetUint64Param(ctx, "userId", true)
	if err != nil {
		return
	}

	if err := Service.AssignRole(userID, req.Role); err != nil {
		api.Abort(ctx, http.StatusInternalServerError, err, errors.New("failed to assign role"))
		return
	}

	ctx.Status(http.StatusOK)
}

// UnassignRole godoc
// @Summary Remove role from a user
// @Description Remove any assigned role from a user by ID
// @Tags Admin
// @Param userId path int true "User ID"
// @Security BearerAuth
// @Success 200 "Role unassigned successfully"
// @Failure 500 {object} role.ErrorRoleResponse "Failed to unassign role"
// @Router /admin/users/{userId}/unassign [post]
func UnassignRole(ctx *gin.Context) {
	userID, err := api.GetUint64Param(ctx, "userId", true)
	if err != nil {
		return
	}

	if err := Service.UnassignRole(userID); err != nil {
		api.Abort(ctx, http.StatusInternalServerError, err, errors.New("failed to unassign role"))
		return
	}

	ctx.Status(http.StatusOK)
}
