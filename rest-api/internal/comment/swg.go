package comment

import "time"

type CommentSwagger struct {
	Content string `json:"content" example:"This is my comment"`
	Status  string `json:"status" example:"active"`
}

type CommentResponse struct {
	ID        uint64     `json:"id" example:"123"`
	PostID    uint64     `json:"post_id" example:"1"`
	UserID    uint64     `json:"user_id" example:"99"`
	Author    UserPublic `json:"author"`
	Content   string     `json:"content" example:"This is a comment"`
	Status    string     `json:"status" example:"active"`
	CreatedAt time.Time  `json:"created_at" example:"2025-10-04T12:00:00Z"`
	UpdatedAt time.Time  `json:"updated_at" example:"2025-10-04T12:00:00Z"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type CommentsResponse struct {
	Total   int               `json:"total" example:"100"`
	Results []CommentResponse `json:"results"`
}

type ErrorResponse struct {
	Message string `json:"message" example:"error message"`
}

type UserPublic struct {
	ID       uint64 `json:"id" example:"99"`
	Username string `json:"username" example:"john_doe"`
	Name     string `json:"name" example:"John Doe"`
}
