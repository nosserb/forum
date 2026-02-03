// Adds/removes `active` class on like/dislike buttons
// Persists per-post choice in localStorage so the counter stays colored after reload
(function () {
	const STORAGE_KEY = 'forum_likes_v1';

	function loadState() {
		try {
			return JSON.parse(localStorage.getItem(STORAGE_KEY)) || {};
		} catch (e) {
			return {};
		}
	}

	function saveState(state) {
		try {
			localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
		} catch (e) {
			// ignore quota errors
		}
	}

	// Return an entity key and the article element. Key format: "post:ID" or "comment:ID"
	function findEntity(form) {
		// prefer explicit hidden inputs
		const pid = form.querySelector('input[name="post_id"]');
		if (pid && pid.value) return { key: 'post:' + pid.value, article: form.closest('article') };
		const cid = form.querySelector('input[name="comment_id"]');
		if (cid && cid.value) return { key: 'comment:' + cid.value, article: form.closest('article') };
		// fallback: data-id on article with class post or reply
		const art = form.closest('article.post, article.reply');
		if (art && art.dataset && art.dataset.id) {
			if (art.classList.contains('post')) return { key: 'post:' + art.dataset.id, article: art };
			return { key: 'comment:' + art.dataset.id, article: art };
		}
		return { key: null, article: null };
	}

	function applyStateToKey(key, state) {
		if (!key) return;
		const parts = key.split(':');
		if (parts.length !== 2) return;
		const t = parts[0];
		const id = parts[1];
		let sel = '';
		if (t === 'post') sel = 'article.post[data-id="' + id + '"]';
		else sel = 'article.reply[data-id="' + id + '"]';
		const article = document.querySelector(sel);
		if (!article) return;

		const likeBtn = article.querySelector('.like-btn');
		const dislikeBtn = article.querySelector('.dislike-btn');

		if (likeBtn) likeBtn.classList.toggle('active', state === 'like');
		if (dislikeBtn) dislikeBtn.classList.toggle('active', state === 'dislike');
	}

	function updateCountsOptimistic(article, type, prevType) {
		const likeEl = article.querySelector('.like-btn .count');
		const dislikeEl = article.querySelector('.dislike-btn .count');

		const likeVal = likeEl ? parseInt(likeEl.textContent || '0', 10) : 0;
		const dislikeVal = dislikeEl ? parseInt(dislikeEl.textContent || '0', 10) : 0;

		let newLike = likeVal;
		let newDislike = dislikeVal;

		if (prevType === type) {
			// toggling off
			if (type === 'like') newLike = Math.max(0, likeVal - 1);
			else newDislike = Math.max(0, dislikeVal - 1);
		} else {
			// toggling on
			if (type === 'like') newLike = likeVal + 1;
			else newDislike = dislikeVal + 1;

			// if switching from other, decrement it
			if (prevType === 'like') newLike = Math.max(0, newLike - 1);
			if (prevType === 'dislike') newDislike = Math.max(0, newDislike - 1);
		}

		if (likeEl) likeEl.textContent = String(newLike);
		if (dislikeEl) dislikeEl.textContent = String(newDislike);
	}

	function init() {
		const state = loadState();

		// Apply stored state to existing posts/comments
		Object.keys(state).forEach(k => {
			applyStateToKey(k, state[k]);
		});

		// Intercept form submissions (like/dislike forms inside .post-actions)
		document.querySelectorAll('.post-actions form').forEach(form => {
			form.addEventListener('submit', async function (ev) {
				ev.preventDefault();

				const submitter = ev.submitter || form.querySelector('button, input[type="submit"]');
				let type = null;
				if (submitter && submitter.classList) {
					if (submitter.classList.contains('like-btn')) type = 'like';
					if (submitter.classList.contains('dislike-btn')) type = 'dislike';
				}

				// fallback: check action path
				if (!type) {
					try {
						const url = new URL(form.action, window.location.href);
						if (url.pathname.endsWith('/like')) type = 'like';
						if (url.pathname.endsWith('/dislike')) type = 'dislike';
					} catch (e) {
						// ignore
					}
				}

				const entity = findEntity(form);
				const key = entity.key;
				const article = entity.article || form.closest('article');
				if (!key || !type || !article) return;

				const prev = state[key];

				// Optimistic UI update
				updateCountsOptimistic(article, type, prev);

				// Update active classes immediately
				if (type === prev) {
					delete state[key];
				} else {
					state[key] = type;
				}
				applyStateToKey(key, state[key]);
				saveState(state);

				// Send the request to server to persist (do not cause full page reload)
				try {
					const response = await fetch(form.action, {
						method: 'POST',
						body: new FormData(form),
						credentials: 'same-origin',
						headers: {
							'X-Requested-With': 'XMLHttpRequest',
							'Accept': 'application/json, text/plain, */*'
						}
					});

					// If server responds with non-OK, revert optimistic change
					if (!response.ok) {
						// revert
						if (prev === undefined) delete state[key]; else state[key] = prev;
						saveState(state);
						applyStateToKey(key, state[key]);
						// attempt to restore counts by reloading the page as fallback
						console.error('Like/dislike request failed, status=', response.status);
					} else {
						// if server returns JSON with authoritative counts, update UI
						try {
							const ct = await response.json();
							// expects { likes: N, dislikes: M } optionally
							if (ct && article) {
								const likeEl = article.querySelector('.like-btn .count');
								const dislikeEl = article.querySelector('.dislike-btn .count');
								if (ct.likes != null && likeEl) likeEl.textContent = String(ct.likes);
								if (ct.dislikes != null && dislikeEl) dislikeEl.textContent = String(ct.dislikes);
							}
						} catch (e) {
							// non-json response is OK; we keep optimistic UI
						}
					}
				} catch (err) {
					// network error: revert optimistic UI
					if (prev === undefined) delete state[key]; else state[key] = prev;
					saveState(state);
					applyStateToKey(key, state[key]);
					console.error('Network error sending like/dislike', err);
				}
			});
		});
	}

	if (document.readyState === 'loading') {
		document.addEventListener('DOMContentLoaded', init);
	} else {
		init();
	}
})();
