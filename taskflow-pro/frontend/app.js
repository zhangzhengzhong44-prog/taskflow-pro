const API_BASE = window.API_BASE || "http://localhost:8090/api/v1";

const state = {
  token: localStorage.getItem("taskflow_token") || "",
  user: null,
  projects: [],
  activeProjectId: Number(localStorage.getItem("taskflow_project_id")) || 0,
  activeTaskId: 0,
  filter: "",
};

const $ = (id) => document.getElementById(id);

function toast(message) {
  const el = $("toast");
  el.textContent = message;
  el.classList.remove("hidden");
  setTimeout(() => el.classList.add("hidden"), 2600);
}

async function api(path, options = {}) {
  const headers = {
    "Content-Type": "application/json",
    ...(options.headers || {}),
  };
  if (state.token) headers.Authorization = `Bearer ${state.token}`;

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });
  const body = await res.json().catch(() => ({ message: "响应格式错误" }));
  if (!res.ok || body.code !== 0) {
    throw new Error(body.message || "请求失败");
  }
  return body.data;
}

function setAuthUI() {
  $("auth-panel").classList.toggle("hidden", Boolean(state.token));
  $("user-panel").classList.toggle("hidden", !state.token);
  $("current-user").textContent = state.user
    ? `${state.user.username} (#${state.user.id})`
    : "已登录";
}

async function register() {
  const data = await api("/auth/register", {
    method: "POST",
    body: JSON.stringify({
      email: $("email").value.trim(),
      username: $("username").value.trim(),
      password: $("password").value,
    }),
  });
  state.token = data.token;
  state.user = data.user;
  localStorage.setItem("taskflow_token", state.token);
  setAuthUI();
  await loadProjects();
  toast("注册并登录成功");
}

async function login() {
  const data = await api("/auth/login", {
    method: "POST",
    body: JSON.stringify({
      email: $("email").value.trim(),
      password: $("password").value,
    }),
  });
  state.token = data.token;
  state.user = data.user;
  localStorage.setItem("taskflow_token", state.token);
  setAuthUI();
  await loadProjects();
  toast("登录成功");
}

async function loadMe() {
  if (!state.token) {
    setAuthUI();
    return;
  }
  try {
    state.user = await api("/auth/me");
    setAuthUI();
  } catch {
    logout();
  }
}

function logout() {
  state.token = "";
  state.user = null;
  state.projects = [];
  state.activeProjectId = 0;
  localStorage.removeItem("taskflow_token");
  localStorage.removeItem("taskflow_project_id");
  renderProjects();
  renderTasks([]);
  setStats({ total: 0, todo: 0, doing: 0, done: 0 });
  setAuthUI();
}

async function createProject() {
  const name = $("project-name").value.trim();
  if (!name) return toast("项目名称不能为空");
  const project = await api("/projects", {
    method: "POST",
    body: JSON.stringify({
      name,
      description: $("project-desc").value.trim(),
    }),
  });
  $("project-name").value = "";
  $("project-desc").value = "";
  state.activeProjectId = project.id;
  localStorage.setItem("taskflow_project_id", project.id);
  await loadProjects();
  toast("项目创建成功");
}

async function loadProjects() {
  if (!state.token) return;
  state.projects = await api("/projects");
  if (!state.activeProjectId && state.projects.length) {
    state.activeProjectId = state.projects[0].id;
  }
  renderProjects();
  await refreshProjectData();
}

function renderProjects() {
  const list = $("project-list");
  list.innerHTML = "";
  state.projects.forEach((project) => {
    const btn = document.createElement("button");
    btn.className = `project-item ${project.id === state.activeProjectId ? "active" : ""}`;
    btn.textContent = `${project.name} #${project.id}`;
    btn.onclick = async () => {
      state.activeProjectId = project.id;
      localStorage.setItem("taskflow_project_id", project.id);
      renderProjects();
      await refreshProjectData();
    };
    list.appendChild(btn);
  });
}

async function addMember() {
  if (!state.activeProjectId) return toast("请先选择项目");
  const userId = Number($("member-user-id").value);
  if (!userId) return toast("请输入用户 ID");
  await api(`/projects/${state.activeProjectId}/members`, {
    method: "POST",
    body: JSON.stringify({ user_id: userId }),
  });
  $("member-user-id").value = "";
  toast("成员添加成功");
}

async function refreshProjectData() {
  const project = state.projects.find((item) => item.id === state.activeProjectId);
  $("active-project-title").textContent = project ? project.name : "请选择或创建一个项目";
  if (!state.activeProjectId) return;
  await Promise.all([loadStats(), loadTasks()]);
}

async function loadStats() {
  const stats = await api(`/projects/${state.activeProjectId}/stats`);
  setStats(stats);
}

function setStats(stats) {
  $("stat-total").textContent = stats.total || 0;
  $("stat-todo").textContent = stats.todo || 0;
  $("stat-doing").textContent = stats.doing || 0;
  $("stat-done").textContent = stats.done || 0;
}

async function createTask() {
  if (!state.activeProjectId) return toast("请先选择项目");
  const title = $("task-title").value.trim();
  if (!title) return toast("任务标题不能为空");
  await api(`/projects/${state.activeProjectId}/tasks`, {
    method: "POST",
    body: JSON.stringify({
      title,
      priority: $("task-priority").value,
      description: $("task-desc").value.trim(),
    }),
  });
  $("task-title").value = "";
  $("task-desc").value = "";
  await refreshProjectData();
  toast("任务创建成功");
}

async function loadTasks() {
  const query = new URLSearchParams();
  if (state.filter) query.set("status", state.filter);
  const keyword = $("keyword").value.trim();
  if (keyword) query.set("keyword", keyword);
  const suffix = query.toString() ? `?${query}` : "";
  const tasks = await api(`/projects/${state.activeProjectId}/tasks${suffix}`);
  renderTasks(tasks || []);
}

function renderTasks(tasks) {
  ["todo", "doing", "done"].forEach((status) => {
    $(`${status}-list`).innerHTML = "";
  });
  tasks.forEach((task) => {
    const card = document.createElement("article");
    card.className = "task-card";
    card.innerHTML = `
      <span class="badge ${task.priority}">${priorityText(task.priority)}</span>
      <h4>${escapeHtml(task.title)}</h4>
      <p>${escapeHtml(task.description || "暂无描述")}</p>
      <div class="task-actions">
        <button data-view>详情</button>
        <button data-status="todo">待处理</button>
        <button data-status="doing">进行中</button>
        <button data-status="done">完成</button>
        <button data-delete>删除</button>
      </div>
    `;
    card.querySelector("[data-view]").onclick = () => openTask(task);
    card.querySelectorAll("[data-status]").forEach((btn) => {
      btn.onclick = () => updateTaskStatus(task.id, btn.dataset.status);
    });
    card.querySelector("[data-delete]").onclick = () => deleteTask(task.id);
    $(`${task.status}-list`).appendChild(card);
  });
}

async function updateTaskStatus(taskId, status) {
  await api(`/tasks/${taskId}`, {
    method: "PUT",
    body: JSON.stringify({ status }),
  });
  await refreshProjectData();
}

async function deleteTask(taskId) {
  await api(`/tasks/${taskId}`, { method: "DELETE" });
  await refreshProjectData();
  toast("任务已删除");
}

async function openTask(task) {
  state.activeTaskId = task.id;
  $("detail-title").textContent = task.title;
  $("detail-desc").textContent = task.description || "暂无描述";
  await loadComments();
}

async function loadComments() {
  if (!state.activeTaskId) return;
  const comments = await api(`/tasks/${state.activeTaskId}/comments`);
  const list = $("comment-list");
  list.innerHTML = "";
  comments.forEach((comment) => {
    const el = document.createElement("div");
    el.className = "comment";
    el.innerHTML = `<p>${escapeHtml(comment.content)}</p><span>用户 #${comment.user_id}</span>`;
    list.appendChild(el);
  });
}

async function createComment() {
  if (!state.activeTaskId) return toast("请先选择任务");
  const content = $("comment-content").value.trim();
  if (!content) return toast("评论不能为空");
  await api(`/tasks/${state.activeTaskId}/comments`, {
    method: "POST",
    body: JSON.stringify({ content }),
  });
  $("comment-content").value = "";
  await loadComments();
}

function priorityText(priority) {
  return { low: "低", medium: "中", high: "高" }[priority] || priority;
}

function escapeHtml(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

function bindEvents() {
  $("register-btn").onclick = () => register().catch((err) => toast(err.message));
  $("login-btn").onclick = () => login().catch((err) => toast(err.message));
  $("logout-btn").onclick = logout;
  $("refresh-projects").onclick = () => loadProjects().catch((err) => toast(err.message));
  $("create-project-btn").onclick = () => createProject().catch((err) => toast(err.message));
  $("add-member-btn").onclick = () => addMember().catch((err) => toast(err.message));
  $("create-task-btn").onclick = () => createTask().catch((err) => toast(err.message));
  $("create-comment-btn").onclick = () => createComment().catch((err) => toast(err.message));
  $("keyword").oninput = () => loadTasks().catch(() => {});
  document.querySelectorAll("[data-filter]").forEach((btn) => {
    btn.onclick = () => {
      state.filter = btn.dataset.filter;
      loadTasks().catch((err) => toast(err.message));
    };
  });
}

bindEvents();
loadMe().then(loadProjects).catch(() => {});

