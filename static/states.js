async function loadStatesForPage() {
  const items = Array.from(document.querySelectorAll("[data-item-id]"))
    .map((el) => el.dataset.itemId)
    .filter((v, i, arr) => v && arr.indexOf(v) === i);
  if (items.length === 0) return;

  const res = await fetch("/states", {
    method: "POST",
    credentials: "same-origin",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ items }),
  });
  if (!res.ok) return;
  const data = await res.json();
  document.querySelectorAll("[data-item-id]").forEach((el) => {
    const id = el.dataset.itemId;
    if (el.classList.contains("btn_like")) {
      if (data[id] && data[id].liked) el.classList.add("active");
      else el.classList.remove("active");
    } else if (el.classList.contains("btn_watched")) {
      if (data[id] && data[id].watched) el.classList.add("active");
      else el.classList.remove("active");
    }
  });
}

document.addEventListener("click", async (e) => {
  const likeBtn = e.target.closest("button.btn_like");
  const watchBtn = e.target.closest("button.btn_watched");
  if (likeBtn) {
    const id = likeBtn.dataset.itemId;
    // optimistic UI
    likeBtn.classList.toggle("active");
    try {
      const res = await fetch("/state", {
        method: "POST",
        credentials: "same-origin",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ item: id, action: "like" }),
      });
      if (!res.ok) {
        likeBtn.classList.toggle("active"); // rollback
      } else {
        const js = await res.json();
        if (!js.liked) likeBtn.classList.remove("active");
      }
    } catch (err) {
      likeBtn.classList.toggle("active"); // rollback
    }
  } else if (watchBtn) {
    const id = watchBtn.dataset.itemId;
    watchBtn.classList.toggle("active");
    try {
      const res = await fetch("/state", {
        method: "POST",
        credentials: "same-origin",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ item: id, action: "watched" }),
      });
      if (!res.ok) {
        watchBtn.classList.toggle("active");
      } else {
        const js = await res.json();
        if (!js.watched) watchBtn.classList.remove("active");
      }
    } catch (err) {
      watchBtn.classList.toggle("active");
    }
  }
});

document.addEventListener("DOMContentLoaded", loadStatesForPage);
