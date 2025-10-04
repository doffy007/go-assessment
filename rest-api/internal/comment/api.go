package comment

import (
	"encoding/json"
	"errors"
	"net/http"
	"rest-api/internal/api"
	"rest-api/internal/user"
	"rest-api/internal/util"

	apiparams "rest-api/internal/params"

	"github.com/gin-gonic/gin"
)

// CreateComment godoc
// @Summary Create a comment
// @Description Create a new comment for a post
// @Tags Comment
// @Accept json
// @Produce json
// @Param postID path int true "Post ID"
// @Param request body comment.CommentSwagger true "Comment payload"
// @Success 200 {object} comment.CommentResponse
// @Failure 400 {object} comment.ErrorResponse
// @Failure 401 {object} comment.ErrorResponse
// @Failure 500 {object} comment.ErrorResponse
// @Security BearerAuth
// @Router /comments/posts/{postID} [post]
func CreateComment(ctx *gin.Context) {
	var req Comment
	if err := ctx.BindJSON(&req); err != nil {
		api.Abort(ctx, http.StatusBadRequest, err, nil)
		return
	}

	userID, err := api.GetUserID(ctx, true)
	if err != nil {
		return
	}
	req.UserID = userID

	postID, err := api.GetUint64Param(ctx, "postID", true)
	if err != nil {
		api.Abort(ctx, http.StatusBadRequest, err, errors.New("invalid post ID"))
		return
	}
	req.PostID = postID

	if err := req.validate(); err != nil {
		api.Abort(ctx, http.StatusBadRequest, err, nil)
		return
	}

	created, err := Service.CreateComment(&req)
	if err != nil {
		api.Abort(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	ctx.JSON(http.StatusOK, created)
}

// GetComment godoc
// @Summary Get a comment by ID
// @Description Get comment details by commentID
// @Tags Comment
// @Accept json
// @Produce json
// @Param commentID path int true "Comment ID"
// @Success 200 {object} comment.CommentResponse
// @Failure 404 {object} comment.ErrorResponse
// @Failure 500 {object} comment.ErrorResponse
// @Router /comments/{commentID} [get]
func GetComment(ctx *gin.Context) {
	id, err := api.GetUint64Param(ctx, "commentID", true)
	if err != nil {
		return
	}

	req := Comment{ID: id}
	if err = Service.GetComment(&req); err != nil {
		if err == api.ErrNotFound {
			api.Abort(ctx, http.StatusNotFound, err, nil)
		} else {
			api.Abort(ctx, http.StatusInternalServerError, err, errors.New("failed to get comment"))
		}
		return
	}

	ctx.JSON(http.StatusOK, req)
}

// GetComments godoc
// @Summary List comments
// @Description Get list of comments with optional search, filters, and pagination
// @Tags Comment
// @Accept json
// @Produce json
// @Param search query string false "Search query"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} comment.CommentsResponse
// @Failure 500 {object} comment.ErrorResponse
// @Router /comments [get]
func GetComments(ctx *gin.Context) {
	resp, err := Service.GetComments(apiparams.GetParams(ctx))
	if err != nil {
		api.Abort(ctx, http.StatusInternalServerError, err, errors.New("failed to get comments"))
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// GetCommentsByPost godoc
// @Summary List comments for a post
// @Description Get list of comments for a specific post
// @Tags Comment
// @Accept json
// @Produce json
// @Param postID path int true "Post ID"
// @Success 200 {object} comment.CommentsResponse
// @Failure 500 {object} comment.ErrorResponse
// @Router /comments/posts/{postID} [get]
func GetCommentsByPost(ctx *gin.Context) {
	postID, err := api.GetUint64Param(ctx, "postID", true)
	if err != nil {
		return
	}

	comments, err := Service.GetCommentsByPostID(postID)
	if err != nil {
		api.Abort(ctx, http.StatusInternalServerError, err, errors.New("failed to get comments by post"))
		return
	}

	ctx.JSON(http.StatusOK, comments)
}

// UpdateComment godoc
// @Summary Update a comment
// @Description Update an existing comment by ID
// @Tags Comment
// @Accept json
// @Produce json
// @Param commentID path int true "Comment ID"
// @Param request body comment.CommentSwagger true "Comment payload"
// @Success 200 {object} comment.CommentResponse
// @Failure 400 {object} comment.ErrorResponse
// @Failure 500 {object} comment.ErrorResponse
// @Security BearerAuth
// @Router /comments/{commentID} [put]
func UpdateComment(ctx *gin.Context) {
	b, err := ctx.GetRawData()
	if err != nil {
		api.Abort(ctx, http.StatusBadRequest, err, api.ErrPayload)
		return
	}
	updatedFields := util.GetUpdatedJSONFields(b)

	req := &Comment{Author: &user.Public{}}
	if err = json.Unmarshal(b, req); err != nil {
		api.Abort(ctx, http.StatusBadRequest, err, api.ErrPayload)
		return
	}

	if req.ID, err = api.GetUint64Param(ctx, "commentID", true); err != nil {
		return
	}

	if err := Service.UpdateComment(req, updatedFields, nil); err != nil {
		api.Abort(ctx, http.StatusInternalServerError, err, errors.New("failed to update comment"))
		return
	}

	ctx.JSON(http.StatusOK, req)
}

// DeleteComment godoc
// @Summary Delete a comment
// @Description Soft delete a comment by ID
// @Tags Comment
// @Accept json
// @Produce json
// @Param commentID path int true "Comment ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} comment.ErrorResponse
// @Failure 404 {object} comment.ErrorResponse
// @Failure 500 {object} comment.ErrorResponse
// @Security BearerAuth
// @Router /comments/{commentID} [delete]
func DeleteComment(ctx *gin.Context) {
	id, err := api.GetUint64Param(ctx, "commentID", true)
	if err != nil {
		api.Abort(ctx, http.StatusBadRequest, err, errors.New("invalid comment ID"))
		return
	}

	deleted, err := Service.DeleteComment(id)
	if err != nil {
		api.Abort(ctx, http.StatusInternalServerError, err, errors.New("failed to delete comment"))
		return
	}

	if !deleted {
		api.Abort(ctx, http.StatusNotFound, nil, errors.New("comment not found"))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted": deleted})
}
