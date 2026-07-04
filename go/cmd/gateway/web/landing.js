const host = location.hostname;
document.querySelectorAll(".workbench-link").forEach(link => { link.href = `http://${host}:8888`; });
document.querySelectorAll(".ide-link").forEach(link => { link.href = `http://${host}:13337`; });

fetch("/api/v1/blogs")
  .then(response => response.ok ? response.json() : Promise.reject(new Error("blog unavailable")))
  .then(data => {
    const posts = (data.items || []).slice(0, 3);
    document.querySelector("#landing-blog-grid").innerHTML = posts.length ? posts.map(post => `<article class="landing-blog-card"><div class="blog-card-meta"><span>${post.tags[0] || "Engineering"}</span><time>${new Date(post.published_at || post.created_at).toLocaleDateString(undefined,{year:"numeric",month:"short",day:"numeric"})}</time></div><h3>${escapeLanding(post.title)}</h3><p>${escapeLanding(post.summary)}</p><a href="/blog.html?slug=${encodeURIComponent(post.slug)}">Read article →</a></article>`).join("") : "<p>No engineering posts published yet.</p>";
  })
  .catch(() => { document.querySelector("#landing-blog-grid").innerHTML = "<p>Engineering posts are temporarily unavailable.</p>"; });

function escapeLanding(value) {
  return String(value || "").replace(/[&<>"']/g, character => ({"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#39;"}[character]));
}
