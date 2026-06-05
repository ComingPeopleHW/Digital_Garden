package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/ComingPeopleHW/Digital_Garden/backend/internal/garden"
	"github.com/ComingPeopleHW/Digital_Garden/backend/internal/store/postgres"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/crypto/bcrypt"
)

type ContentStore interface {
	ListUsers(ctx context.Context) ([]garden.User, error)
	ListPosts(ctx context.Context) ([]garden.Post, error)
	UserByUsername(ctx context.Context, username string) (garden.User, error)
	ListPostsByUsername(ctx context.Context, username string) ([]garden.Post, error)
	CreateUser(ctx context.Context, input garden.CreateUserInput) (garden.User, error)
	FindUserForLogin(ctx context.Context, login string) (garden.User, string, error)
	CreateSession(ctx context.Context, token string, userID string) error
	UserBySessionToken(ctx context.Context, token string) (garden.User, error)
	DeleteSession(ctx context.Context, token string) error
	UpdateUserProfile(ctx context.Context, input garden.UpdateUserProfileInput) (garden.User, error)
	CreatePost(ctx context.Context, input garden.CreatePostInput) (garden.Post, error)
	ReactToPost(ctx context.Context, postID string, userKey string, reaction garden.ReactionType) (garden.Post, error)
}

const sessionCookieName = "dg_session"

func NewRouter(frontendOrigin string, store ContentStore) http.Handler {
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
		r.Get("/me", me(store))
		r.Post("/auth/register", register(store))
		r.Post("/auth/login", login(store))
		r.Post("/auth/logout", logout(store))
		r.Patch("/me/profile", updateProfile(store))
		r.Get("/users", listUsers(store))
		r.Get("/users/{username}", getUser(store))
		r.Get("/users/{username}/posts", listUserPosts(store))
		r.Get("/posts", listPosts(store))
		r.Post("/posts", createPost(store))
		r.Post("/posts/{postID}/reactions", reactToPost(store))
	})

	return r
}

func me(store ContentStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := currentUser(r.Context())
		if ok {
			writeJSON(w, http.StatusOK, map[string]any{"user": user})
			return
		}

		token, ok := sessionToken(r)
		if !ok {
			writeJSON(w, http.StatusOK, map[string]any{"user": nil})
			return
		}

		user, err := store.UserBySessionToken(r.Context(), token)
		if err != nil {
			if errors.Is(err, postgres.ErrNotFound) {
				clearSessionCookie(w)
				writeJSON(w, http.StatusOK, map[string]any{"user": nil})
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to load session")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"user": user})
	}
}

func register(store ContentStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			Name     string `json:"name"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		payload.Username = strings.TrimSpace(payload.Username)
		payload.Email = strings.TrimSpace(payload.Email)
		payload.Name = strings.TrimSpace(payload.Name)
		if payload.Name == "" {
			payload.Name = payload.Username
		}
		if payload.Username == "" || payload.Email == "" || len(payload.Password) < 8 {
			writeError(w, http.StatusBadRequest, "username, email, and 8+ character password are required")
			return
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to secure password")
			return
		}

		user, err := store.CreateUser(r.Context(), garden.CreateUserInput{
			ID:           "user_" + randomHex(12),
			Username:     payload.Username,
			Email:        payload.Email,
			Name:         payload.Name,
			PasswordHash: string(passwordHash),
		})
		if err != nil {
			if errors.Is(err, postgres.ErrConflict) {
				writeError(w, http.StatusConflict, "username or email already exists")
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to create user")
			return
		}

		token := randomHex(32)
		if err := store.CreateSession(r.Context(), token, user.ID); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create session")
			return
		}
		setSessionCookie(w, token)
		writeJSON(w, http.StatusCreated, map[string]garden.User{"user": user})
	}
}

func login(store ContentStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Login    string `json:"login"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		user, passwordHash, err := store.FindUserForLogin(r.Context(), strings.TrimSpace(payload.Login))
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid login or password")
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(payload.Password)); err != nil {
			writeError(w, http.StatusUnauthorized, "invalid login or password")
			return
		}

		token := randomHex(32)
		if err := store.CreateSession(r.Context(), token, user.ID); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create session")
			return
		}
		setSessionCookie(w, token)
		writeJSON(w, http.StatusOK, map[string]garden.User{"user": user})
	}
}

func logout(store ContentStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if token, ok := sessionToken(r); ok {
			if err := store.DeleteSession(r.Context(), token); err != nil {
				writeError(w, http.StatusInternalServerError, "failed to log out")
				return
			}
		}
		clearSessionCookie(w)
		w.WriteHeader(http.StatusNoContent)
	}
}

func updateProfile(store ContentStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := requireUser(r.Context(), r, store)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "login required")
			return
		}

		var payload struct {
			Name      string `json:"name"`
			Bio       string `json:"bio"`
			AvatarURL string `json:"avatarUrl"`
			Location  string `json:"location"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		payload.Name = strings.TrimSpace(payload.Name)
		payload.Bio = strings.TrimSpace(payload.Bio)
		payload.AvatarURL = strings.TrimSpace(payload.AvatarURL)
		payload.Location = strings.TrimSpace(payload.Location)
		if payload.Name == "" {
			writeError(w, http.StatusBadRequest, "name is required")
			return
		}

		updatedUser, err := store.UpdateUserProfile(r.Context(), garden.UpdateUserProfileInput{
			UserID:    user.ID,
			Name:      payload.Name,
			Bio:       payload.Bio,
			AvatarURL: payload.AvatarURL,
			Location:  payload.Location,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to update profile")
			return
		}

		writeJSON(w, http.StatusOK, map[string]garden.User{"user": updatedUser})
	}
}

func listUsers(store ContentStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := store.ListUsers(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list users")
			return
		}
		writeJSON(w, http.StatusOK, users)
	}
}

func getUser(store ContentStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := store.UserByUsername(r.Context(), chi.URLParam(r, "username"))
		if err != nil {
			if errors.Is(err, postgres.ErrNotFound) {
				writeError(w, http.StatusNotFound, "user not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to load user")
			return
		}
		writeJSON(w, http.StatusOK, user)
	}
}

func listPosts(store ContentStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		posts, err := store.ListPosts(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list posts")
			return
		}
		writeJSON(w, http.StatusOK, posts)
	}
}

func listUserPosts(store ContentStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		posts, err := store.ListPostsByUsername(r.Context(), chi.URLParam(r, "username"))
		if err != nil {
			if errors.Is(err, postgres.ErrNotFound) {
				writeError(w, http.StatusNotFound, "user not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to list user posts")
			return
		}
		writeJSON(w, http.StatusOK, posts)
	}
}

func createPost(store ContentStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := requireUser(r.Context(), r, store)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "login required")
			return
		}

		var payload struct {
			Title    string `json:"title"`
			Body     string `json:"body"`
			ImageURL string `json:"imageUrl"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		payload.Title = strings.TrimSpace(payload.Title)
		payload.Body = strings.TrimSpace(payload.Body)
		payload.ImageURL = strings.TrimSpace(payload.ImageURL)
		if payload.Title == "" || payload.Body == "" {
			writeError(w, http.StatusBadRequest, "title and body are required")
			return
		}

		post, err := store.CreatePost(r.Context(), garden.CreatePostInput{
			ID:       "post_" + randomHex(12),
			AuthorID: user.ID,
			Title:    payload.Title,
			Body:     payload.Body,
			ImageURL: payload.ImageURL,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create post")
			return
		}

		writeJSON(w, http.StatusCreated, post)
	}
}

func reactToPost(store ContentStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Type string `json:"type"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		reaction := garden.ReactionType(payload.Type)
		if !reaction.Valid() {
			writeError(w, http.StatusBadRequest, "reaction type must be upvote or downvote")
			return
		}

		postID := chi.URLParam(r, "postID")
		userKey := reactionUserKey(r)
		post, err := store.ReactToPost(r.Context(), postID, userKey, reaction)
		if err != nil {
			if errors.Is(err, postgres.ErrNotFound) {
				writeError(w, http.StatusNotFound, "post not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to react to post")
			return
		}

		writeJSON(w, http.StatusOK, post)
	}
}

func reactionUserKey(r *http.Request) string {
	if token, ok := sessionToken(r); ok {
		return "session:" + token
	}
	userKey := strings.TrimSpace(r.Header.Get("X-Demo-User-Key"))
	if userKey == "" {
		return "demo-viewer"
	}
	return userKey
}

func cors(frontendOrigin string) func(http.Handler) http.Handler {
	allowedOrigins := map[string]struct{}{
		frontendOrigin:          {},
		"http://localhost:5173": {},
		"http://127.0.0.1:5173": {},
		"http://localhost:4173": {},
		"http://127.0.0.1:4173": {},
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if origin := r.Header.Get("Origin"); origin != "" {
				if _, ok := allowedOrigins[origin]; ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					w.Header().Set("Vary", "Origin")
				}
			}
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Demo-User-Key")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type contextKey string

const userContextKey contextKey = "user"

func currentUser(ctx context.Context) (garden.User, bool) {
	user, ok := ctx.Value(userContextKey).(garden.User)
	return user, ok
}

func requireUser(ctx context.Context, r *http.Request, store ContentStore) (garden.User, error) {
	if user, ok := currentUser(ctx); ok {
		return user, nil
	}
	token, ok := sessionToken(r)
	if !ok {
		return garden.User{}, postgres.ErrNotFound
	}
	return store.UserBySessionToken(ctx, token)
}

func sessionToken(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		return "", false
	}
	return cookie.Value, true
}

func setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int((14 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func randomHex(bytesLen int) string {
	bytes := make([]byte, bytesLen)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
