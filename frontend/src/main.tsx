import React, { useEffect, useMemo, useState } from "react";
import { createRoot } from "react-dom/client";
import {
  ArrowDown,
  ArrowUp,
  Check,
  Edit3,
  Home,
  Leaf,
  LogOut,
  MapPin,
  PenLine,
  Save,
  Send,
  Settings,
  Trash2,
  X,
  Users,
} from "lucide-react";
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
  const [profileUser, setProfileUser] = useState<User | null>(null);
  const [profileUsername, setProfileUsername] = useState(getProfileUsername());
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
  const [profileForm, setProfileForm] = useState({ name: "", bio: "", avatarUrl: "", location: "" });
  const [isEditingProfile, setIsEditingProfile] = useState(false);
  const [notice, setNotice] = useState("");

  useEffect(() => {
    function handlePopState() {
      setProfileUsername(getProfileUsername());
    }

    window.addEventListener("popstate", handlePopState);
    return () => window.removeEventListener("popstate", handlePopState);
  }, []);

  useEffect(() => {
    setIsLoading(true);
    setNotice("");
    const postsPath = profileUsername ? `/api/users/${encodeURIComponent(profileUsername)}/posts` : "/api/posts";
    const profileRequest = profileUsername
      ? apiFetch(`/api/users/${encodeURIComponent(profileUsername)}`).then((response) =>
          response.ok ? response.json() : null,
        )
      : Promise.resolve(null);

    Promise.all([
      apiFetch(postsPath).then((response) => (response.ok ? response.json() : [])),
      apiFetch("/api/users").then((response) => response.json()),
      apiFetch("/api/me").then((response) => response.json()),
      profileRequest,
    ])
      .then(([nextPosts, nextUsers, session, nextProfileUser]) => {
        setPosts(nextPosts);
        setUsers(nextUsers);
        setMe(session.user);
        setProfileUser(nextProfileUser);
      })
      .finally(() => setIsLoading(false));
  }, [profileUsername]);

  useEffect(() => {
    if (!me) {
      setProfileForm({ name: "", bio: "", avatarUrl: "", location: "" });
      setIsEditingProfile(false);
      return;
    }

    setProfileForm({
      name: me.name,
      bio: me.bio,
      avatarUrl: me.avatarUrl,
      location: me.location,
    });
  }, [me]);

  const featuredUser = profileUser ?? me ?? users[0];
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
    setPosts((current) =>
      profileUsername && profileUsername.toLowerCase() !== post.author.username.toLowerCase()
        ? current
        : [post, ...current],
    );
    setPostForm({ title: "", body: "", imageUrl: "" });
  }

  async function submitProfile(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setNotice("");
    const response = await apiFetch("/api/me/profile", {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(profileForm),
    });
    if (!response.ok) {
      setNotice("资料保存失败");
      return;
    }

    const session = (await response.json()) as { user: User };
    applyUserUpdate(session.user);
    setIsEditingProfile(false);
    setNotice("资料已保存");
  }

  function applyUserUpdate(user: User) {
    setMe(user);
    setUsers((current) => current.map((item) => (item.id === user.id ? user : item)));
    setProfileUser((current) => (current?.id === user.id ? user : current));
    setPosts((current) =>
      current.map((post) => (post.author.id === user.id ? { ...post, author: user } : post)),
    );
  }

  async function updatePost(postId: string, input: { title: string; body: string; imageUrl: string }) {
    setNotice("");
    const response = await apiFetch(`/api/posts/${postId}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(input),
    });
    if (!response.ok) {
      setNotice("文章保存失败");
      return false;
    }

    const updatedPost = (await response.json()) as Post;
    setPosts((current) => current.map((post) => (post.id === updatedPost.id ? updatedPost : post)));
    return true;
  }

  async function deletePost(postId: string) {
    setNotice("");
    const response = await apiFetch(`/api/posts/${postId}`, { method: "DELETE" });
    if (!response.ok) {
      setNotice("文章删除失败");
      return false;
    }

    setPosts((current) => current.filter((post) => post.id !== postId));
    return true;
  }

  function navigateToProfile(username: string) {
    window.history.pushState({}, "", `/u/${encodeURIComponent(username)}`);
    setProfileUsername(username);
  }

  function navigateHome() {
    window.history.pushState({}, "", "/");
    setProfileUsername(null);
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
            <div className="signedInPanel">
              <div className="accountLine">
                <button type="button" className="accountIdentity" onClick={() => navigateToProfile(me.username)}>
                  <Avatar user={me} size="small" />
                  <span>@{me.username}</span>
                </button>
                <div className="accountActions">
                  <button
                    type="button"
                    onClick={() => setIsEditingProfile((current) => !current)}
                    aria-label="编辑资料"
                  >
                    <Settings size={17} />
                  </button>
                  <button type="button" onClick={logout} aria-label="退出登录">
                    <LogOut size={17} />
                  </button>
                </div>
              </div>

              {isEditingProfile ? (
                <form className="profileForm" onSubmit={submitProfile}>
                  <input
                    value={profileForm.name}
                    onChange={(event) => setProfileForm((current) => ({ ...current, name: event.target.value }))}
                    placeholder="昵称"
                    required
                  />
                  <textarea
                    value={profileForm.bio}
                    onChange={(event) => setProfileForm((current) => ({ ...current, bio: event.target.value }))}
                    placeholder="简介"
                  />
                  <input
                    value={profileForm.avatarUrl}
                    onChange={(event) =>
                      setProfileForm((current) => ({ ...current, avatarUrl: event.target.value }))
                    }
                    placeholder="头像 URL"
                  />
                  <input
                    value={profileForm.location}
                    onChange={(event) =>
                      setProfileForm((current) => ({ ...current, location: event.target.value }))
                    }
                    placeholder="位置"
                  />
                  <button type="submit">
                    <Save size={18} />
                    保存资料
                  </button>
                </form>
              ) : (
                <form className="composeForm" onSubmit={submitPost}>
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
              )}
            </div>
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
            <Avatar user={featuredUser} size="small" />
            <div>
              <strong>{featuredUser.name}</strong>
              <span>@{featuredUser.username}</span>
            </div>
            <p>{featuredUser.bio}</p>
            <button type="button" onClick={() => navigateToProfile(featuredUser.username)}>
              {profileUser?.id === featuredUser.id ? "当前主页" : "查看主页"}
            </button>
          </section>
        )}
      </aside>

      <section className="content">
        <header className="topbar">
          <div>
            <p className="eyebrow">Public Feed</p>
            <h2>{profileUser ? `${profileUser.name} 的主页` : "最新发布"}</h2>
          </div>
          {profileUsername && (
            <button className="homeButton" type="button" onClick={navigateHome} aria-label="回到全部内容">
              <Home size={17} />
              全部
            </button>
          )}
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

        {profileUsername && (
          <ProfileHeader user={profileUser} postCount={posts.length} isLoading={isLoading} />
        )}

        <div className="feed">
          {isLoading ? (
            <div className="emptyState">正在载入花园内容...</div>
          ) : posts.length === 0 ? (
            <div className="emptyState">{profileUsername ? "这个主页还没有发布内容" : "还没有发布内容"}</div>
          ) : (
            posts.map((post) => (
              <PostCard
                key={post.id}
                post={post}
                me={me}
                onReact={reactToPost}
                onAuthorClick={navigateToProfile}
                onUpdate={updatePost}
                onDelete={deletePost}
              />
            ))
          )}
        </div>
      </section>
    </main>
  );
}

function ProfileHeader({
  user,
  postCount,
  isLoading,
}: {
  user: User | null;
  postCount: number;
  isLoading: boolean;
}) {
  if (isLoading) {
    return <section className="profileHeader skeleton">正在载入主页...</section>;
  }

  if (!user) {
    return <section className="profileHeader missing">没有找到这个用户</section>;
  }

  return (
    <section className="profileHeader">
      <Avatar user={user} size="large" />
      <div className="profileMeta">
        <div>
          <h3>{user.name}</h3>
          <span>@{user.username}</span>
        </div>
        <p>{user.bio || "这个用户还没有填写简介。"}</p>
        <div className="profileFacts">
          {user.location && (
            <span>
              <MapPin size={15} />
              {user.location}
            </span>
          )}
          <span>{postCount} posts</span>
        </div>
      </div>
    </section>
  );
}

function PostCard({
  post,
  me,
  onReact,
  onAuthorClick,
  onUpdate,
  onDelete,
}: {
  post: Post;
  me: User | null;
  onReact: (postId: string, type: "upvote" | "downvote") => void;
  onAuthorClick: (username: string) => void;
  onUpdate: (postId: string, input: { title: string; body: string; imageUrl: string }) => Promise<boolean>;
  onDelete: (postId: string) => Promise<boolean>;
}) {
  const [isEditing, setIsEditing] = useState(false);
  const [editForm, setEditForm] = useState({
    title: post.title,
    body: post.body,
    imageUrl: post.imageUrl ?? "",
  });
  const isAuthor = me?.id === post.author.id;

  useEffect(() => {
    if (!isEditing) {
      setEditForm({
        title: post.title,
        body: post.body,
        imageUrl: post.imageUrl ?? "",
      });
    }
  }, [isEditing, post]);

  async function submitEdit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const ok = await onUpdate(post.id, editForm);
    if (ok) {
      setIsEditing(false);
    }
  }

  async function handleDelete() {
    const ok = await onDelete(post.id);
    if (ok) {
      setIsEditing(false);
    }
  }

  return (
    <article className="postCard">
      <div className="postAuthor">
        <Avatar user={post.author} size="small" />
        <div>
          <button type="button" onClick={() => onAuthorClick(post.author.username)}>
            {post.author.name}
          </button>
          <span>
            @{post.author.username} · {formatDate(post.createdAt)}
          </span>
        </div>
        <span className="location">
          <MapPin size={14} />
          {post.author.location}
        </span>
      </div>

      {isEditing ? (
        <form className="postEditForm" onSubmit={submitEdit}>
          <input
            value={editForm.title}
            onChange={(event) => setEditForm((current) => ({ ...current, title: event.target.value }))}
            placeholder="标题"
            required
          />
          <textarea
            value={editForm.body}
            onChange={(event) => setEditForm((current) => ({ ...current, body: event.target.value }))}
            placeholder="正文"
            required
          />
          <input
            value={editForm.imageUrl}
            onChange={(event) => setEditForm((current) => ({ ...current, imageUrl: event.target.value }))}
            placeholder="图片 URL"
          />
          <div className="postEditActions">
            <button type="submit">
              <Check size={17} />
              保存
            </button>
            <button type="button" onClick={() => setIsEditing(false)}>
              <X size={17} />
              取消
            </button>
          </div>
        </form>
      ) : (
        <>
          <h3>{post.title}</h3>
          <p>{post.body}</p>
          {post.imageUrl && <img className="postImage" src={post.imageUrl} alt="" />}
        </>
      )}

      <div className="reactionBar">
        <button type="button" onClick={() => onReact(post.id, "upvote")}>
          <ArrowUp size={18} />
          {post.upvotes}
        </button>
        <button type="button" onClick={() => onReact(post.id, "downvote")}>
          <ArrowDown size={18} />
          {post.downvotes}
        </button>
        {isAuthor && !isEditing && (
          <div className="ownerActions">
            <button type="button" onClick={() => setIsEditing(true)} aria-label="编辑文章">
              <Edit3 size={17} />
            </button>
            <button type="button" onClick={handleDelete} aria-label="删除文章">
              <Trash2 size={17} />
            </button>
          </div>
        )}
      </div>
    </article>
  );
}

function Avatar({ user, size }: { user: User; size: "small" | "large" }) {
  const className = `avatar avatar-${size}`;
  if (user.avatarUrl) {
    return <img className={className} src={user.avatarUrl} alt="" />;
  }

  return (
    <span className={className} aria-hidden="true">
      {initialsFor(user)}
    </span>
  );
}

function initialsFor(user: User) {
  return (user.name || user.username).trim().slice(0, 2).toUpperCase();
}

function getProfileUsername() {
  const match = window.location.pathname.match(/^\/u\/([^/]+)\/?$/);
  return match ? decodeURIComponent(match[1]) : null;
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
