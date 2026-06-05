package garden

import "time"

type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	Bio       string `json:"bio"`
	AvatarURL string `json:"avatarUrl"`
	Location  string `json:"location"`
}

type Post struct {
	ID        string    `json:"id"`
	Author    User      `json:"author"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	ImageURL  string    `json:"imageUrl,omitempty"`
	Upvotes   int       `json:"upvotes"`
	Downvotes int       `json:"downvotes"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreateUserInput struct {
	ID           string
	Username     string
	Email        string
	Name         string
	PasswordHash string
}

type UpdateUserProfileInput struct {
	UserID    string
	Name      string
	Bio       string
	AvatarURL string
	Location  string
}

type CreatePostInput struct {
	ID       string
	AuthorID string
	Title    string
	Body     string
	ImageURL string
}

type UpdatePostInput struct {
	PostID   string
	AuthorID string
	Title    string
	Body     string
	ImageURL string
}

type ReactionType string

const (
	ReactionUpvote   ReactionType = "upvote"
	ReactionDownvote ReactionType = "downvote"
)

func (r ReactionType) Valid() bool {
	return r == ReactionUpvote || r == ReactionDownvote
}
