package post

import (
	"encoding/json"
	"errors"
	"net/http"
	"rest-api/internal/api"
	"rest-api/internal/util"
	"time"

	apiparams "rest-api/internal/params"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

// CreatePost godoc
// @Summary Create a post
// @Description Create a new post
// @Tags Post
// @Accept json
// @Produce json
// @Param request body post.PostSwagger true "Post payload"
// @Success 200 {object} post.PostResponse
// @Failure 400 {object} post.ErrorResponse
// @Failure 500 {object} post.ErrorResponse
// @Security BearerAuth
// @Router /writter/posts [post]
func CreatePost(ctx *gin.Context) {
	var reqBody CreatePostRequest
	if err := ctx.BindJSON(&reqBody); err != nil {
		api.Abort(ctx, http.StatusBadRequest, err, nil)
		return
	}

	userID, exists := ctx.Get(api.UserIDKey)
	if !exists {
		api.Abort(ctx, http.StatusUnauthorized, errors.New("missing user id in context"), nil)
		return
	}

	req := &Post{
		Type:        &reqBody.Type,
		Title:       &reqBody.Title,
		Slug:        &reqBody.Slug,
		Excerpt:     &reqBody.Excerpt,
		SEOTitle:    &reqBody.SEOTitle,
		SEOExcerpt:  &reqBody.SEOExcerpt,
		Content:     reqBody.Content,
		ContentHTML: &reqBody.ContentHTML,
		ContentText: &reqBody.ContentText,
		Tags:        reqBody.Tags,
		Status:      reqBody.Status,
		UserID:      userID.(uint64),
		Metadata:    reqBody.Metadata,
	}

	now := time.Now()
	req.CreatedAt = pgtype.Timestamptz{Time: now, Valid: true}
	req.UpdatedAt = pgtype.Timestamptz{Time: now, Valid: true}

	if reqBody.ScheduledAt != nil {
		req.ScheduledAt = pgtype.Timestamptz{Time: *reqBody.ScheduledAt, Valid: true}
	}

	if reqBody.PublishedAt != nil {
		req.PublishedAt = pgtype.Timestamptz{Time: *reqBody.PublishedAt, Valid: true}
	}

	if err := req.validate(); err != nil {
		api.Abort(ctx, http.StatusBadRequest, err, nil)
		return
	}

	createdPost, err := Service.CreatePost(req)
	if err != nil {
		api.Abort(ctx, http.StatusInternalServerError, err, nil)
		return
	}

	ctx.JSON(http.StatusOK, createdPost)
}

// GetPost godoc
// @Summary Get a post by ID
// @Description Get post details by postID
// @Tags Post
// @Accept json
// @Produce json
// @Param postID path int true "Post ID"
// @Success 200 {object} post.PostResponse
// @Failure 404 {object} post.ErrorResponse
// @Failure 500 {object} post.ErrorResponse
// @Router /posts/{postID} [get]
func GetPost(ctx *gin.Context) {
	postID, err := api.GetUint64Param(ctx, "postID", true)
	if err != nil {
		return
	}

	var req Post
	req.ID = postID

	if err = Service.GetPost(&req); err != nil {
		if err == api.ErrNotFound {
			api.Abort(ctx, http.StatusNotFound, err, nil)
		} else {
			api.Abort(ctx, http.StatusInternalServerError, err, errors.New("failed to get post"))
		}
		return
	}

	ctx.JSON(http.StatusOK, req)
}

// GetPosts godoc
// @Summary List posts
// @Description Get list of posts with optional search, filters, and pagination
// @Tags Post
// @Accept json
// @Produce json
// @Param search query string false "Search query"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} post.PostsResponse
// @Failure 500 {object} post.ErrorResponse
// @Router /posts [get]
func GetPosts(ctx *gin.Context) {
	resp, err := Service.GetPosts(apiparams.GetParams(ctx))
	if err != nil {
		api.Abort(ctx, http.StatusInternalServerError, err, errors.New("failed to get posts"))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdatePost godoc
// @Summary Update a post
// @Description Update an existing post by ID
// @Tags Post
// @Accept json
// @Produce json
// @Param postID path int true "Post ID"
// @Param request body post.PostSwagger true "Post payload"
// @Success 200 {object} post.PostResponse
// @Failure 400 {object} post.ErrorResponse
// @Failure 500 {object} post.ErrorResponse
// @Security BearerAuth
// @Router /writter/posts/{postID} [put]
func UpdatePost(ctx *gin.Context) {
	b, err := ctx.GetRawData()
	if err != nil {
		api.Abort(ctx, http.StatusBadRequest, err, api.ErrPayload)
		return
	}
	updatedFields := util.GetUpdatedJSONFields(b)

	req := &Post{}
	if err = json.Unmarshal(b, req); err != nil {
		api.Abort(ctx, http.StatusBadRequest, err, api.ErrPayload)
		return
	}

	userID, err := api.GetUserID(ctx, true)
	if err != nil {
		return
	}
	req.UserID = userID

	if req.ID, err = api.GetUint64Param(ctx, "postID", true); err != nil {
		return
	}

	if err := Service.UpdatePost(req, updatedFields, nil); err != nil {
		api.Abort(ctx, http.StatusInternalServerError, err, errors.New("failed to update post"))
		return
	}

	ctx.JSON(http.StatusOK, req)
}

// DeletePost godoc
// @Summary Delete a post
// @Description Soft delete a post by ID
// @Tags Post
// @Accept json
// @Produce json
// @Param postID path int true "Post ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} post.ErrorResponse
// @Failure 404 {object} post.ErrorResponse
// @Failure 500 {object} post.ErrorResponse
// @Security BearerAuth
// @Router /writter/posts/{postID} [delete]
func DeletePost(ctx *gin.Context) {
	id, err := api.GetUint64Param(ctx, "postID", true)
	if err != nil {
		api.Abort(ctx, http.StatusBadRequest, err, errors.New("invalid post ID"))
		return
	}

	deleted, err := Service.DeletePost(id)
	if err != nil {
		api.Abort(ctx, http.StatusInternalServerError, err, errors.New("failed to delete post"))
		return
	}

	if !deleted {
		api.Abort(ctx, http.StatusNotFound, nil, errors.New("post not found"))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted": deleted})
}
