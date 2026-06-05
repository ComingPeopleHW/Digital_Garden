package postgres

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"sort"

	"github.com/ComingPeopleHW/Digital_Garden/backend/internal/garden"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	ErrConflict = errors.New("conflict")
	ErrNotFound = errors.New("not found")

	//go:embed migrations/*.sql
	migrationFiles embed.FS
)

type Store struct {
	db *sql.DB
}

func Open(ctx context.Context, databaseURL string) (*Store, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Migrate(ctx context.Context) error {
	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		body, err := migrationFiles.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if _, err := s.db.ExecContext(ctx, string(body)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
	}

	return nil
}

func (s *Store) ListUsers(ctx context.Context) ([]garden.User, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, username, name, bio, avatar_url, location
		FROM users
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]garden.User, 0)
	for rows.Next() {
		var user garden.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Name, &user.Bio, &user.AvatarURL, &user.Location); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}

	return users, nil
}

func (s *Store) UserByUsername(ctx context.Context, username string) (garden.User, error) {
	var user garden.User
	if err := s.db.QueryRowContext(ctx, `
		SELECT id, username, name, bio, avatar_url, location
		FROM users
		WHERE lower(username) = lower($1)
	`, username).Scan(
		&user.ID,
		&user.Username,
		&user.Name,
		&user.Bio,
		&user.AvatarURL,
		&user.Location,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return garden.User{}, ErrNotFound
		}
		return garden.User{}, fmt.Errorf("find user by username: %w", err)
	}

	return user, nil
}

func (s *Store) ListPosts(ctx context.Context) ([]garden.Post, error) {
	rows, err := s.db.QueryContext(ctx, postSelectSQL+`
		ORDER BY p.created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()

	posts := make([]garden.Post, 0)
	for rows.Next() {
		post, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate posts: %w", err)
	}

	return posts, nil
}

func (s *Store) ListPostsByUsername(ctx context.Context, username string) ([]garden.Post, error) {
	if _, err := s.UserByUsername(ctx, username); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, postSelectSQL+`
		HAVING lower(u.username) = lower($1)
		ORDER BY p.created_at DESC
	`, username)
	if err != nil {
		return nil, fmt.Errorf("list posts by username: %w", err)
	}
	defer rows.Close()

	posts := make([]garden.Post, 0)
	for rows.Next() {
		post, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user posts: %w", err)
	}

	return posts, nil
}

func (s *Store) CreateUser(ctx context.Context, input garden.CreateUserInput) (garden.User, error) {
	var user garden.User
	if err := s.db.QueryRowContext(ctx, `
		INSERT INTO users (id, username, email, name, password_hash, bio, avatar_url, location)
		VALUES ($1, $2, $3, $4, $5, '', '', '')
		RETURNING id, username, name, bio, avatar_url, location
	`, input.ID, input.Username, input.Email, input.Name, input.PasswordHash).Scan(
		&user.ID,
		&user.Username,
		&user.Name,
		&user.Bio,
		&user.AvatarURL,
		&user.Location,
	); err != nil {
		if isUniqueViolation(err) {
			return garden.User{}, ErrConflict
		}
		return garden.User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (s *Store) FindUserForLogin(ctx context.Context, login string) (garden.User, string, error) {
	var user garden.User
	var passwordHash string
	if err := s.db.QueryRowContext(ctx, `
		SELECT id, username, name, bio, avatar_url, location, password_hash
		FROM users
		WHERE (lower(username) = lower($1) OR lower(email) = lower($1))
			AND password_hash IS NOT NULL
		LIMIT 1
	`, login).Scan(
		&user.ID,
		&user.Username,
		&user.Name,
		&user.Bio,
		&user.AvatarURL,
		&user.Location,
		&passwordHash,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return garden.User{}, "", ErrNotFound
		}
		return garden.User{}, "", fmt.Errorf("find user for login: %w", err)
	}

	return user, passwordHash, nil
}

func (s *Store) CreateSession(ctx context.Context, token string, userID string) error {
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO sessions (token, user_id, expires_at)
		VALUES ($1, $2, now() + interval '14 days')
	`, token, userID); err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (s *Store) UserBySessionToken(ctx context.Context, token string) (garden.User, error) {
	var user garden.User
	if err := s.db.QueryRowContext(ctx, `
		SELECT u.id, u.username, u.name, u.bio, u.avatar_url, u.location
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token = $1 AND s.expires_at > now()
	`, token).Scan(
		&user.ID,
		&user.Username,
		&user.Name,
		&user.Bio,
		&user.AvatarURL,
		&user.Location,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return garden.User{}, ErrNotFound
		}
		return garden.User{}, fmt.Errorf("find user by session: %w", err)
	}

	return user, nil
}

func (s *Store) DeleteSession(ctx context.Context, token string) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE token = $1`, token); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (s *Store) UpdateUserProfile(ctx context.Context, input garden.UpdateUserProfileInput) (garden.User, error) {
	var user garden.User
	if err := s.db.QueryRowContext(ctx, `
		UPDATE users
		SET name = $2,
			bio = $3,
			avatar_url = $4,
			location = $5
		WHERE id = $1
		RETURNING id, username, name, bio, avatar_url, location
	`, input.UserID, input.Name, input.Bio, input.AvatarURL, input.Location).Scan(
		&user.ID,
		&user.Username,
		&user.Name,
		&user.Bio,
		&user.AvatarURL,
		&user.Location,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return garden.User{}, ErrNotFound
		}
		return garden.User{}, fmt.Errorf("update user profile: %w", err)
	}

	return user, nil
}

func (s *Store) CreatePost(ctx context.Context, input garden.CreatePostInput) (garden.Post, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return garden.Post{}, fmt.Errorf("begin create post transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO posts (id, author_id, title, body)
		VALUES ($1, $2, $3, $4)
	`, input.ID, input.AuthorID, input.Title, input.Body); err != nil {
		if isForeignKeyError(err) {
			return garden.Post{}, ErrNotFound
		}
		return garden.Post{}, fmt.Errorf("insert post: %w", err)
	}

	if input.ImageURL != "" {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO post_media (post_id, url, sort_order)
			VALUES ($1, $2, 0)
		`, input.ID, input.ImageURL); err != nil {
			return garden.Post{}, fmt.Errorf("insert post media: %w", err)
		}
	}

	post, err := getPostTx(ctx, tx, input.ID)
	if err != nil {
		return garden.Post{}, err
	}

	if err := tx.Commit(); err != nil {
		return garden.Post{}, fmt.Errorf("commit create post transaction: %w", err)
	}

	return post, nil
}

func (s *Store) ReactToPost(ctx context.Context, postID string, userKey string, reaction garden.ReactionType) (garden.Post, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return garden.Post{}, fmt.Errorf("begin reaction transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO post_reactions (post_id, user_key, reaction_type)
		VALUES ($1, $2, $3)
		ON CONFLICT (post_id, user_key)
		DO UPDATE SET reaction_type = EXCLUDED.reaction_type, updated_at = now()
	`, postID, userKey, string(reaction)); err != nil {
		if isForeignKeyError(err) {
			return garden.Post{}, ErrNotFound
		}
		return garden.Post{}, fmt.Errorf("upsert reaction: %w", err)
	}

	post, err := getPostTx(ctx, tx, postID)
	if err != nil {
		return garden.Post{}, err
	}

	if err := tx.Commit(); err != nil {
		return garden.Post{}, fmt.Errorf("commit reaction transaction: %w", err)
	}

	return post, nil
}

const postSelectSQL = `
	SELECT
		p.id,
		p.title,
		p.body,
		COALESCE(pm.url, '') AS image_url,
		p.created_at,
		u.id,
		u.username,
		u.name,
		u.bio,
		u.avatar_url,
		u.location,
		COUNT(*) FILTER (WHERE pr.reaction_type = 'upvote') AS upvotes,
		COUNT(*) FILTER (WHERE pr.reaction_type = 'downvote') AS downvotes
	FROM posts p
	JOIN users u ON u.id = p.author_id
	LEFT JOIN LATERAL (
		SELECT url
		FROM post_media
		WHERE post_id = p.id
		ORDER BY sort_order ASC, id ASC
		LIMIT 1
	) pm ON true
	LEFT JOIN post_reactions pr ON pr.post_id = p.id
	GROUP BY p.id, p.title, p.body, pm.url, p.created_at, u.id, u.username, u.name, u.bio, u.avatar_url, u.location
`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPost(row rowScanner) (garden.Post, error) {
	var post garden.Post
	if err := row.Scan(
		&post.ID,
		&post.Title,
		&post.Body,
		&post.ImageURL,
		&post.CreatedAt,
		&post.Author.ID,
		&post.Author.Username,
		&post.Author.Name,
		&post.Author.Bio,
		&post.Author.AvatarURL,
		&post.Author.Location,
		&post.Upvotes,
		&post.Downvotes,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return garden.Post{}, ErrNotFound
		}
		return garden.Post{}, fmt.Errorf("scan post: %w", err)
	}

	return post, nil
}

func getPostTx(ctx context.Context, tx *sql.Tx, postID string) (garden.Post, error) {
	row := tx.QueryRowContext(ctx, postSelectSQL+`
		HAVING p.id = $1
	`, postID)
	return scanPost(row)
}

func isForeignKeyError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
