import React, { useEffect, useMemo, useState } from "react";
import { createRoot } from "react-dom/client";
import { ArrowDown, ArrowUp, Leaf, LogOut, MapPin, PenLine, Send, Users } from "lucide-react";
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

const API_BASE = import.meta.env.VITE_API_BASE ?? "http://127.0.0.1:8080";
const DEMO_USER_KEY_STORAGE = "digital-garden-demo-user-key";

type AuthMode = "login" | "register";

function App() {
  const [posts, setPosts] = useState<Post[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [me, setMe] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [authMode, setAuthMode] = useState<AuthMode>("login");
  const [authForm, setAuthForm] = useState({
    username: "",
    email: "",
    name: "",
    login: "",
    password: "",
  });
  const [postForm, setPostForm] = useState({ title: "", body: "", imageUrl: "" });
  const [notice, setNotice] = useState("");

  useEffect(() => {
    Promise.all([
      apiFetch("/api/posts").then((response) => response.json()),
      apiFetch("/api/users").then((response) => response.json()),
      apiFetch("/api/me").then((response) => response.json()),
    ])
      .then(([nextPosts, nextUsers, session]) => {
        setPosts(nextPosts);
        setUsers(nextUsers);
        setMe(session.user);
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
    const response = await apiFetch(`/api/posts/${postId}/reactions`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Demo-User-Key": getDemoUserKey(),
      },
      body: JSON.stringify({ type }),
    });
    if (!response.ok) {
      return;
    }
    const updatedPost = (await response.json()) as Post;
    setPosts((current) => current.map((post) => (post.id === updatedPost.id ? updatedPost : post)));
  }

  async function submitAuth(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setNotice("");
    const path = authMode === "login" ? "/api/auth/login" : "/api/auth/register";
    const payload =
      authMode === "login"
        ? { login: authForm.login, password: authForm.password }
        : {
            username: authForm.username,
            email: authForm.email,
            name: authForm.name,
            password: authForm.password,
          };

    const response = await apiFetch(path, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (!response.ok) {
      setNotice(authMode === "login" ? "登录失败" : "注册失败");
      return;
    }

    const session = (await response.json()) as { user: User };
    setMe(session.user);
    setUsers((current) =>
      current.some((user) => user.id === session.user.id) ? current : [...current, session.user],
    );
    setAuthForm({ username: "", email: "", name: "", login: "", password: "" });
  }

  async function logout() {
    await apiFetch("/api/auth/logout", { method: "POST" });
    setMe(null);
  }

  async function submitPost(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setNotice("");
    const response = await apiFetch("/api/posts", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(postForm),
    });
    if (!response.ok) {
      setNotice("发布失败");
      return;
    }

    const post = (await response.json()) as Post;
    setPosts((current) => [post, ...current]);
    setPostForm({ title: "", body: "", imageUrl: "" });
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
          {me ? (
            <form className="composeForm" onSubmit={submitPost}>
              <div className="accountLine">
                <span>@{me.username}</span>
                <button type="button" onClick={logout} aria-label="退出登录">
                  <LogOut size={17} />
                </button>
              </div>
              <input
                value={postForm.title}
                onChange={(event) => setPostForm((current) => ({ ...current, title: event.target.value }))}
                placeholder="标题"
                required
              />
              <textarea
                value={postForm.body}
                onChange={(event) => setPostForm((current) => ({ ...current, body: event.target.value }))}
                placeholder="正文"
                required
              />
              <input
                value={postForm.imageUrl}
                onChange={(event) => setPostForm((current) => ({ ...current, imageUrl: event.target.value }))}
                placeholder="图片 URL"
              />
              <button type="submit">
                <Send size={18} />
                发布
              </button>
            </form>
          ) : (
            <form className="authPanel" onSubmit={submitAuth}>
              <div className="authTabs">
                <button
                  type="button"
                  className={authMode === "login" ? "active" : ""}
                  onClick={() => setAuthMode("login")}
                >
                  登录
                </button>
                <button
                  type="button"
                  className={authMode === "register" ? "active" : ""}
                  onClick={() => setAuthMode("register")}
                >
                  注册
                </button>
              </div>
              {authMode === "register" ? (
                <>
                  <input
                    value={authForm.username}
                    onChange={(event) => setAuthForm((current) => ({ ...current, username: event.target.value }))}
                    placeholder="用户名"
                    required
                  />
                  <input
                    value={authForm.email}
                    onChange={(event) => setAuthForm((current) => ({ ...current, email: event.target.value }))}
                    placeholder="邮箱"
                    type="email"
                    required
                  />
                  <input
                    value={authForm.name}
                    onChange={(event) => setAuthForm((current) => ({ ...current, name: event.target.value }))}
                    placeholder="昵称"
                  />
                </>
              ) : (
                <input
                  value={authForm.login}
                  onChange={(event) => setAuthForm((current) => ({ ...current, login: event.target.value }))}
                  placeholder="用户名或邮箱"
                  required
                />
              )}
              <input
                value={authForm.password}
                onChange={(event) => setAuthForm((current) => ({ ...current, password: event.target.value }))}
                placeholder="密码"
                type="password"
                minLength={8}
                required
              />
              <button type="submit">
                <PenLine size={18} />
                {authMode === "login" ? "登录" : "创建账号"}
              </button>
            </form>
          )}
          {notice && <p className="notice">{notice}</p>}
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

function getDemoUserKey() {
  const current = window.localStorage.getItem(DEMO_USER_KEY_STORAGE);
  if (current) {
    return current;
  }

  const next = crypto.randomUUID();
  window.localStorage.setItem(DEMO_USER_KEY_STORAGE, next);
  return next;
}

function apiFetch(path: string, init?: RequestInit) {
  return fetch(`${API_BASE}${path}`, {
    ...init,
    credentials: "include",
  });
}

createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
