package post

import (
	"time"
)

type CreatePostRequest struct {
	Type        string         `json:"type" example:"article"`
	Title       string         `json:"title" example:"Contoh Judul"`
	Slug        string         `json:"slug" example:"contoh-judul"`
	Excerpt     string         `json:"excerpt" example:"Ini excerpt post"`
	SEOTitle    string         `json:"seo_title"`
	SEOExcerpt  string         `json:"seo_excerpt"`
	Content     map[string]any `json:"content"`
	ContentHTML string         `json:"content_html"`
	ContentText string         `json:"content_text"`
	Tags        []string       `json:"tags" example:"[\"tech\",\"golang\"]"`
	Status      string         `json:"status" example:"draft"`
	Metadata    map[string]any `json:"metadata"`
	ScheduledAt *time.Time     `json:"scheduled_at,omitempty"`
	PublishedAt *time.Time     `json:"published_at,omitempty"`
}

type PostResponse struct {
	Post PostSwagger `json:"post"`
}

type PostsResponse struct {
	Total   int           `json:"total"`
	Results []PostSwagger `json:"results"`
}

type PostSwagger struct {
	Type        string            `json:"type" example:"article"`
	Title       string            `json:"title" example:"Contoh Judul"`
	Slug        string            `json:"slug" example:"contoh-judul"`
	Excerpt     string            `json:"excerpt" example:"Ini excerpt post"`
	SEOTitle    string            `json:"seo_title" example:"Contoh SEO Title"`
	SEOExcerpt  string            `json:"seo_excerpt" example:"Contoh SEO Excerpt"`
	Content     map[string]string `json:"content" example:"{\"section1\":\"Konten1\",\"section2\":\"Konten2\"}"`
	ContentHTML string            `json:"content_html" example:"<p>Konten HTML</p>"`
	ContentText string            `json:"content_text" example:"Konten text"`
	Tags        []string          `json:"tags" example:"[\"tech\",\"golang\"]"`
	Status      string            `json:"status" example:"draft"`
	CreatedAt   string            `json:"created_at" example:"2025-10-03T12:00:00Z"`
	UpdatedAt   string            `json:"updated_at" example:"2025-10-03T12:00:00Z"`
	ScheduledAt *string           `json:"scheduled_at,omitempty" example:"2025-10-04T08:00:00Z"`
	PublishedAt *string           `json:"published_at,omitempty" example:"2025-10-04T09:00:00Z"`
	TotalViews  int               `json:"total_views,omitempty" example:"10"`
	Metadata    map[string]string `json:"metadata,omitempty" example:"{\"key1\":\"value1\",\"key2\":\"value2\"}"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}

type UnauthorizedResponse struct {
	Error string `json:"error" example:"unauthorized"`
}
