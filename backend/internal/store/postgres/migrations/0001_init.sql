CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  username TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  bio TEXT NOT NULL DEFAULT '',
  avatar_url TEXT NOT NULL DEFAULT '',
  location TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS posts (
  id TEXT PRIMARY KEY,
  author_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  body TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS post_media (
  id BIGSERIAL PRIMARY KEY,
  post_id TEXT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  url TEXT NOT NULL,
  sort_order INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (post_id, url)
);

CREATE TABLE IF NOT EXISTS post_reactions (
  post_id TEXT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  user_key TEXT NOT NULL,
  reaction_type TEXT NOT NULL CHECK (reaction_type IN ('upvote', 'downvote')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (post_id, user_key)
);

CREATE INDEX IF NOT EXISTS idx_posts_author_id ON posts(author_id);
CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_post_media_post_id ON post_media(post_id);
CREATE INDEX IF NOT EXISTS idx_post_reactions_post_id ON post_reactions(post_id);

INSERT INTO users (id, username, name, bio, avatar_url, location, created_at)
VALUES
  (
    'user_aurora',
    'aurora',
    '林夏',
    '记录产品灵感、城市散步和一些正在变好的小习惯。',
    'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&w=256&q=80',
    'Shanghai',
    now() - interval '3 days'
  ),
  (
    'user_river',
    'river',
    '周屿',
    'Golang engineer. Building calm software and useful personal tools.',
    'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?auto=format&fit=crop&w=256&q=80',
    'Hangzhou',
    now() - interval '2 days'
  ),
  (
    'user_mira',
    'mira',
    'Mira Chen',
    'Photography, essays, notebooks, and tiny public experiments.',
    'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&w=256&q=80',
    'Singapore',
    now() - interval '1 day'
  )
ON CONFLICT (id) DO NOTHING;

INSERT INTO posts (id, author_id, title, body, created_at)
VALUES
  (
    'post_001',
    'user_aurora',
    '把主页做成一座花园',
    '我希望这里不是简历，也不是传统博客，而是一个可以慢慢生长的空间。文字、图片、链接和生活痕迹都能自然地放进来。',
    now() - interval '3 hours'
  ),
  (
    'post_002',
    'user_river',
    '后端边界先保持朴素',
    '用户、内容、媒体、互动先拆成清晰的模块。第一阶段不急着复杂化，把 API 合同和数据流跑顺更重要。',
    now() - interval '8 hours'
  ),
  (
    'post_003',
    'user_mira',
    '今天的照片墙',
    '光线好的时候，连临时拍下来的角落都像是在提醒我：页面也应该留一点呼吸感。',
    now() - interval '24 hours'
  )
ON CONFLICT (id) DO NOTHING;

INSERT INTO post_media (post_id, url, sort_order)
VALUES
  ('post_001', 'https://images.unsplash.com/photo-1497215728101-856f4ea42174?auto=format&fit=crop&w=1200&q=80', 0),
  ('post_003', 'https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=1200&q=80', 0)
ON CONFLICT DO NOTHING;

INSERT INTO post_reactions (post_id, user_key, reaction_type)
SELECT 'post_001', 'seed_post_001_up_' || n, 'upvote' FROM generate_series(1, 42) AS n
ON CONFLICT DO NOTHING;

INSERT INTO post_reactions (post_id, user_key, reaction_type)
SELECT 'post_001', 'seed_post_001_down_' || n, 'downvote' FROM generate_series(1, 2) AS n
ON CONFLICT DO NOTHING;

INSERT INTO post_reactions (post_id, user_key, reaction_type)
SELECT 'post_002', 'seed_post_002_up_' || n, 'upvote' FROM generate_series(1, 31) AS n
ON CONFLICT DO NOTHING;

INSERT INTO post_reactions (post_id, user_key, reaction_type)
SELECT 'post_002', 'seed_post_002_down_' || n, 'downvote' FROM generate_series(1, 1) AS n
ON CONFLICT DO NOTHING;

INSERT INTO post_reactions (post_id, user_key, reaction_type)
SELECT 'post_003', 'seed_post_003_up_' || n, 'upvote' FROM generate_series(1, 58) AS n
ON CONFLICT DO NOTHING;

INSERT INTO post_reactions (post_id, user_key, reaction_type)
SELECT 'post_003', 'seed_post_003_down_' || n, 'downvote' FROM generate_series(1, 4) AS n
ON CONFLICT DO NOTHING;
