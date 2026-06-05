package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

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

type Store struct {
	mu    sync.RWMutex
	users []User
	posts []Post
}

func NewRouter(frontendOrigin string) http.Handler {
	store := newSeedStore()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors(frontendOrigin))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api", func(r chi.Router) {
		r.Get("/users", store.listUsers)
		r.Get("/posts", store.listPosts)
		r.Post("/posts/{postID}/reactions", store.reactToPost)
	})

	return r
}

func (s *Store) listUsers(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	writeJSON(w, http.StatusOK, s.users)
}

func (s *Store) listPosts(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	writeJSON(w, http.StatusOK, s.posts)
}

func (s *Store) reactToPost(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Type string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if payload.Type != "upvote" && payload.Type != "downvote" {
		writeError(w, http.StatusBadRequest, "reaction type must be upvote or downvote")
		return
	}

	postID := chi.URLParam(r, "postID")

	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.posts {
		if s.posts[i].ID != postID {
			continue
		}
		if payload.Type == "upvote" {
			s.posts[i].Upvotes++
		} else {
			s.posts[i].Downvotes++
		}
		writeJSON(w, http.StatusOK, s.posts[i])
		return
	}

	writeError(w, http.StatusNotFound, "post not found")
}

func cors(frontendOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", frontendOrigin)
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func newSeedStore() *Store {
	users := []User{
		{
			ID:        "user_aurora",
			Username:  "aurora",
			Name:      "林夏",
			Bio:       "记录产品灵感、城市散步和一些正在变好的小习惯。",
			AvatarURL: "https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&w=256&q=80",
			Location:  "Shanghai",
		},
		{
			ID:        "user_river",
			Username:  "river",
			Name:      "周屿",
			Bio:       "Golang engineer. Building calm software and useful personal tools.",
			AvatarURL: "https://images.unsplash.com/photo-1500648767791-00dcc994a43e?auto=format&fit=crop&w=256&q=80",
			Location:  "Hangzhou",
		},
		{
			ID:        "user_mira",
			Username:  "mira",
			Name:      "Mira Chen",
			Bio:       "Photography, essays, notebooks, and tiny public experiments.",
			AvatarURL: "https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&w=256&q=80",
			Location:  "Singapore",
		},
	}

	return &Store{
		users: users,
		posts: []Post{
			{
				ID:        "post_001",
				Author:    users[0],
				Title:     "把主页做成一座花园",
				Body:      "我希望这里不是简历，也不是传统博客，而是一个可以慢慢生长的空间。文字、图片、链接和生活痕迹都能自然地放进来。",
				ImageURL:  "https://images.unsplash.com/photo-1497215728101-856f4ea42174?auto=format&fit=crop&w=1200&q=80",
				Upvotes:   42,
				Downvotes: 2,
				CreatedAt: time.Now().Add(-3 * time.Hour),
			},
			{
				ID:        "post_002",
				Author:    users[1],
				Title:     "后端边界先保持朴素",
				Body:      "用户、内容、媒体、互动先拆成清晰的模块。第一阶段不急着复杂化，把 API 合同和数据流跑顺更重要。",
				Upvotes:   31,
				Downvotes: 1,
				CreatedAt: time.Now().Add(-8 * time.Hour),
			},
			{
				ID:        "post_003",
				Author:    users[2],
				Title:     "今天的照片墙",
				Body:      "光线好的时候，连临时拍下来的角落都像是在提醒我：页面也应该留一点呼吸感。",
				ImageURL:  "https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=1200&q=80",
				Upvotes:   58,
				Downvotes: 4,
				CreatedAt: time.Now().Add(-24 * time.Hour),
			},
		},
	}
}
