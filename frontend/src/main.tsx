import React, { useEffect, useMemo, useState } from "react";
import { createRoot } from "react-dom/client";
import { ArrowDown, ArrowUp, Image, Leaf, MapPin, PenLine, Users } from "lucide-react";
import "./styles.css";

type User = {
  id: string;
  username: string;
  name: string;
  bio: string;
  avatarUrl: string;
  location: string;
};

type Post = {
  id: string;
  author: User;
  title: string;
  body: string;
  imageUrl?: string;
  upvotes: number;
  downvotes: number;
  createdAt: string;
};

const API_BASE = import.meta.env.VITE_API_BASE ?? "http://localhost:8080";

function App() {
  const [posts, setPosts] = useState<Post[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      fetch(`${API_BASE}/api/posts`).then((response) => response.json()),
      fetch(`${API_BASE}/api/users`).then((response) => response.json()),
    ])
      .then(([nextPosts, nextUsers]) => {
        setPosts(nextPosts);
        setUsers(nextUsers);
      })
      .finally(() => setIsLoading(false));
  }, []);

  const featuredUser = users[0];
  const totals = useMemo(() => {
    return posts.reduce(
      (acc, post) => {
        acc.upvotes += post.upvotes;
        acc.downvotes += post.downvotes;
        return acc;
      },
      { upvotes: 0, downvotes: 0 },
    );
  }, [posts]);

  async function reactToPost(postId: string, type: "upvote" | "downvote") {
    const response = await fetch(`${API_BASE}/api/posts/${postId}/reactions`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ type }),
    });
    const updatedPost = (await response.json()) as Post;
    setPosts((current) => current.map((post) => (post.id === updatedPost.id ? updatedPost : post)));
  }

  return (
    <main className="shell">
      <aside className="sidebar">
        <div className="brand">
          <span className="brandMark">
            <Leaf size={20} />
          </span>
          <span>Digital Garden</span>
        </div>

        <section className="composePanel">
          <div>
            <p className="eyebrow">多人主页</p>
            <h1>每个人都拥有一块可生长的公开空间</h1>
          </div>
          <div className="quickActions">
            <button type="button">
              <PenLine size={18} />
              发布文字
            </button>
            <button type="button">
              <Image size={18} />
              上传图片
            </button>
          </div>
        </section>

        {featuredUser && (
          <section className="profilePreview">
            <img src={featuredUser.avatarUrl} alt="" />
            <div>
              <strong>{featuredUser.name}</strong>
              <span>@{featuredUser.username}</span>
            </div>
            <p>{featuredUser.bio}</p>
          </section>
        )}
      </aside>

      <section className="content">
        <header className="topbar">
          <div>
            <p className="eyebrow">Public Feed</p>
            <h2>最新发布</h2>
          </div>
          <div className="stats">
            <span>
              <Users size={16} />
              {users.length} users
            </span>
            <span>
              <ArrowUp size={16} />
              {totals.upvotes}
            </span>
            <span>
              <ArrowDown size={16} />
              {totals.downvotes}
            </span>
          </div>
        </header>

        <div className="feed">
          {isLoading ? (
            <div className="emptyState">正在载入花园内容...</div>
          ) : (
            posts.map((post) => <PostCard key={post.id} post={post} onReact={reactToPost} />)
          )}
        </div>
      </section>
    </main>
  );
}

function PostCard({
  post,
  onReact,
}: {
  post: Post;
  onReact: (postId: string, type: "upvote" | "downvote") => void;
}) {
  return (
    <article className="postCard">
      <div className="postAuthor">
        <img src={post.author.avatarUrl} alt="" />
        <div>
          <strong>{post.author.name}</strong>
          <span>
            @{post.author.username} · {formatDate(post.createdAt)}
          </span>
        </div>
        <span className="location">
          <MapPin size={14} />
          {post.author.location}
        </span>
      </div>
      <h3>{post.title}</h3>
      <p>{post.body}</p>
      {post.imageUrl && <img className="postImage" src={post.imageUrl} alt="" />}
      <div className="reactionBar">
        <button type="button" onClick={() => onReact(post.id, "upvote")}>
          <ArrowUp size={18} />
          {post.upvotes}
        </button>
        <button type="button" onClick={() => onReact(post.id, "downvote")}>
          <ArrowDown size={18} />
          {post.downvotes}
        </button>
      </div>
    </article>
  );
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("zh-CN", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}

createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);

