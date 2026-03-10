package forum

// Response wrapper
type APIResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Message    string      `json:"message,omitempty"`
	Error      string      `json:"error,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// Models
type Topic struct {
	ID          string `json:"_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	ThreadCount int    `json:"thread_count"`
	Icon        string `json:"icon"`
	Gradient    string `json:"gradient"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type Thread struct {
	ID             string   `json:"_id"`
	Slug           string   `json:"slug"`
	TopicSlug      string   `json:"topic_slug"`
	MovieID        *int     `json:"movie_id"`
	MovieTitle     *string  `json:"movie_title"`
	Title          string   `json:"title"`
	Content        string   `json:"content"`
	CreatedBy      Author   `json:"created_by"`
	Tags           []string `json:"tags"`
	Stats          Stats    `json:"stats"`
	UpvotedBy      []string `json:"upvoted_by"`
	IsPinned       bool     `json:"is_pinned"`
	IsLocked       bool     `json:"is_locked"`
	IsDeleted      bool     `json:"is_deleted"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
	LastActivityAt string   `json:"last_activity_at"`
}

type Reply struct {
	ID          string     `json:"_id"`
	ThreadID    string     `json:"thread_id"`
	ParentID    *string    `json:"parent_id"`
	Depth       int        `json:"depth"`
	Content     string     `json:"content"`
	CreatedBy   Author     `json:"created_by"`
	Stats       ReplyStats `json:"stats"`
	UpvotedBy   []string   `json:"upvoted_by"`
	DownvotedBy []string   `json:"downvoted_by"`
	IsEdited    bool       `json:"is_edited"`
	EditedAt    *string    `json:"edited_at"`
	IsDeleted   bool       `json:"is_deleted"`
	CreatedAt   string     `json:"created_at"`
	UpdatedAt   string     `json:"updated_at"`
	Children    []Reply    `json:"children"`
}

type Author struct {
	UserID    string  `json:"user_id"`
	Username  string  `json:"username"`
	AvatarURL *string `json:"avatar_url"`
}

type Stats struct {
	ReplyCount int `json:"reply_count"`
	Upvotes    int `json:"upvotes"`
	Views      int `json:"views"`
}

type ReplyStats struct {
	Upvotes   int `json:"upvotes"`
	Downvotes int `json:"downvotes"`
}
