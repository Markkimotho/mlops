const host = location.hostname;
document.querySelectorAll(".workbench-link").forEach(link => { link.href = `http://${host}:8888`; });
document.querySelectorAll(".ide-link").forEach(link => { link.href = `http://${host}:13337`; });
