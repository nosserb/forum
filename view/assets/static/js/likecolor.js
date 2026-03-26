(function () {
	let state = {};
	const cooldownMs = 500;
	const lastActionAt = {};

	function findEntity(form) {
		const pid = form.querySelector('input[name="post_id"]');
		if (pid && pid.value) return { key: 'post:' + pid.value, article: form.closest('article') };
		const cid = form.querySelector('input[name="comment_id"]');
		if (cid && cid.value) return { key: 'comment:' + cid.value, article: form.closest('article') };
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
			if (type === 'like') newLike = Math.max(0, likeVal - 1);
			else newDislike = Math.max(0, dislikeVal - 1);
		} else {
			if (type === 'like') newLike = likeVal + 1;
			else newDislike = dislikeVal + 1;

			if (prevType === 'like') newLike = Math.max(0, newLike - 1);
			if (prevType === 'dislike') newDislike = Math.max(0, newDislike - 1);
		}

		if (likeEl) likeEl.textContent = String(newLike);
		if (dislikeEl) dislikeEl.textContent = String(newDislike);
	}

	function buildStateFromDom() {
		state = {};
		const articles = document.querySelectorAll('article.post[data-id], article.reply[data-id]');
		articles.forEach(article => {
			const id = article.dataset && article.dataset.id;
			if (!id) return;
			const key = article.classList.contains('post') ? 'post:' + id : 'comment:' + id;
			const likeBtn = article.querySelector('.like-btn');
			const dislikeBtn = article.querySelector('.dislike-btn');
			if (likeBtn && likeBtn.classList.contains('active')) {
				state[key] = 'like';
			} else if (dislikeBtn && dislikeBtn.classList.contains('active')) {
				state[key] = 'dislike';
			}
		});
	}

	function applyStateToAll() {
		buildStateFromDom();
	}

	function onSubmit(ev) {
		const form = ev.target;
		if (!form || !form.closest || !form.closest('.post-actions')) return;
		ev.preventDefault();

		const submitter = ev.submitter || form.querySelector('button, input[type="submit"]');
		let type = null;
		if (submitter && submitter.classList) {
			if (submitter.classList.contains('like-btn')) type = 'like';
			if (submitter.classList.contains('dislike-btn')) type = 'dislike';
		}

		if (!type) {
			try {
				const url = new URL(form.action, window.location.href);
				if (url.pathname.endsWith('/like')) type = 'like';
				if (url.pathname.endsWith('/dislike')) type = 'dislike';
			} catch (e) {
			}
		}

		const entity = findEntity(form);
		const key = entity.key;
		const article = entity.article || form.closest('article');
		if (!key || !type || !article) return;

		const now = Date.now();
		if (lastActionAt[key] && now - lastActionAt[key] < cooldownMs) {
			return;
		}
		lastActionAt[key] = now;

		const likeBtn = article.querySelector('.like-btn');
		const dislikeBtn = article.querySelector('.dislike-btn');
		if (likeBtn) likeBtn.disabled = true;
		if (dislikeBtn) dislikeBtn.disabled = true;
		setTimeout(() => {
			if (likeBtn) likeBtn.disabled = false;
			if (dislikeBtn) dislikeBtn.disabled = false;
		}, cooldownMs);

		const prev = state[key];

		updateCountsOptimistic(article, type, prev);

		if (type === prev) {
			delete state[key];
		} else {
			state[key] = type;
		}
		applyStateToKey(key, state[key]);

		fetch(form.action, {
			method: 'POST',
			body: new FormData(form),
			credentials: 'same-origin',
			headers: {
				'X-Requested-With': 'XMLHttpRequest',
				'Accept': 'application/json, text/plain, */*'
			}
		})
			.then(response => {
				if (!response.ok) {
					if (prev === undefined) delete state[key]; else state[key] = prev;
					applyStateToKey(key, state[key]);
					console.error('Like/dislike request failed, status=', response.status);
					return null;
				}
				return response.json().catch(() => null);
			})
			.then(ct => {
				if (ct && article) {
					const likeEl = article.querySelector('.like-btn .count');
					const dislikeEl = article.querySelector('.dislike-btn .count');
					if (ct.likes != null && likeEl) likeEl.textContent = String(ct.likes);
					if (ct.dislikes != null && dislikeEl) dislikeEl.textContent = String(ct.dislikes);
				}
			})
			.catch(err => {
				if (prev === undefined) delete state[key]; else state[key] = prev;
				applyStateToKey(key, state[key]);
				console.error('Network error sending like/dislike', err);
			});
	}

	function init() {
		applyStateToAll();
		document.addEventListener('submit', onSubmit);
		document.addEventListener('likes:refresh', applyStateToAll);
	}

	if (document.readyState === 'loading') {
		document.addEventListener('DOMContentLoaded', init);
	} else {
		init();
	}
})();
