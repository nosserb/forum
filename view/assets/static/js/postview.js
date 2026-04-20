document.addEventListener('DOMContentLoaded', function() {
    document.addEventListener('click', function(e) {
        const postLink = e.target.closest('.post-link');
        if (postLink) {
            e.preventDefault();
            const url = postLink.getAttribute('href');
            const postId = new URL(url, window.location.origin).searchParams.get('id');
            if (postId) {
                showPostDetail(postId, true);
            }
        }
    });

    window.addEventListener('popstate', function() {
        const postId = new URLSearchParams(window.location.search).get('post');
        if (postId) {
            showPostDetail(postId, false);
        } else {
            showPostsList(false);
        }
    });

    const initialPostId = new URLSearchParams(window.location.search).get('post');
    if (initialPostId) {
        showPostDetail(initialPostId, false);
    }
});

function showPostDetail(postId, pushState) {
    const listView = document.getElementById('posts-list-view');
    const detailView = document.getElementById('post-detail-view');
    if (!detailView || !listView) return;

    listView.classList.add('is-hidden');
    detailView.classList.remove('is-hidden');
    detailView.innerHTML = '<p class="small">Loading post...</p>';

    if (pushState) {
        window.history.pushState({}, '', `/?post=${postId}`);
    }

    fetch(`/post?id=${postId}&format=json`, {
        headers: {
            'X-Requested-With': 'XMLHttpRequest',
            'Accept': 'application/json'
        }
    })
        .then(response => {
            if (!response.ok) {
                throw new Error(`Request failed with status ${response.status}`);
            }
            return response.json();
        })
        .then(data => {        
            detailView.innerHTML = renderPostDetail(data, postId);

            if (data.post.imageid != 0) {
                const imageId = data.post.imageid
                addPostImage(imageId, postId) 
            }

            const backBtn = document.getElementById('back-to-posts-btn');
            if (backBtn) {
                backBtn.addEventListener('click', function() {
                    showPostsList(true);
                });
            }
            document.dispatchEvent(new Event('likes:refresh'));
        })
        .catch(error => {
            console.error('Error loading post:', error);
            detailView.innerHTML = `
                <div>
                    <button type="button" class="back-btn" id="back-to-posts-btn">← Back to Posts</button>
                </div>
                <p class="small">Unable to load the post.</p>
            `;
            const backBtn = document.getElementById('back-to-posts-btn');
            if (backBtn) {
                backBtn.addEventListener('click', function() {
                    showPostsList(true);
                });
            }
        });
}

function addPostImage(imageID, postID) {
    if (!imageID || !postID) {
        return
    }

    const postCard = document.getElementsByClassName('post-detail-card')[0]
    const postActions = postCard.children[2]

    const img = document.createElement('img')
    img.src = `/images?id=${imageID}`
    
    // id for css
    img.id = 'post-image'

    postCard.insertBefore(img, postActions)
}

function showPostsList(pushState) {
    const listView = document.getElementById('posts-list-view');
    const detailView = document.getElementById('post-detail-view');
    if (!detailView || !listView) return;

    detailView.classList.add('is-hidden');
    detailView.innerHTML = '';
    listView.classList.remove('is-hidden');

    if (pushState) {
        window.history.pushState({}, '', '/');
    }
}

function renderPostDetail(data, postId) {
    if (!data || !data.post) {
        return '<p class="small">Post not found.</p>';
    }

    const post = data.post;
    const isAuthenticated = Boolean(data.username);
    const likedPost = Boolean(data.likedPosts && data.likedPosts[String(postId)]);
    const dislikedPost = Boolean(data.dislikedPosts && data.dislikedPosts[String(postId)]);

    const categoriesHtml = Array.isArray(post.categories) && post.categories.length
        ? `<div class="post-cats" style="margin-bottom:12px;">${post.categories.map(cat => `<span class="cat-badge">${escapeHtml(cat)}</span>`).join('')}</div>`
        : '';

    const replies = Array.isArray(post.replies) ? post.replies : [];
    const repliesHtml = replies.length
        ? replies.map(reply => renderReply(reply, data)).join('')
        : '<p class="muted-small">No replies yet. Be the first!</p>';

    return `
        <div class="back-btn-container">
            <button type="button" class="back-btn" id="back-to-posts-btn">← Back to Posts</button>
        </div>

        <article class="card post post-detail-card" data-id="${post.id}">
            <div class="post-header">
                <div class="post-header-left">
                    <h4 class="post-detail-title">${escapeHtml(post.title)}</h4>
                </div>
                <div class="post-header-right">
                    <div class="post-author">Author: <strong>${escapeHtml(post.username)}</strong></div>
                    <div class="post-date">${escapeHtml(post.createdAt)}</div>
                </div>
            </div>
            ${categoriesHtml}
            <div class="post-content">${escapeHtml(post.content)}</div>

            <div class="post-actions">
                ${isAuthenticated ? `
                <form method="POST" action="/like" style="display:inline;">
                    <input type="hidden" name="post_id" value="${post.id}">
                    <button type="submit" class="like-btn${likedPost ? ' active' : ''}">
                        <img src="/statics/image/like.png" alt="Like" />
                        <span class="count">${post.likes}</span>
                    </button>
                </form>

                <form method="POST" action="/dislike" style="display:inline;">
                    <input type="hidden" name="post_id" value="${post.id}">
                    <button type="submit" class="dislike-btn${dislikedPost ? ' active' : ''}">
                        <img src="/statics/image/dislike.png" alt="Dislike" />
                        <span class="count">${post.dislikes}</span>
                    </button>
                </form>
                ` : `
                <span class="disabled-like">
                    <img src="/statics/image/like.png" alt="Like" />
                    <span class="count">${post.likes}</span>
                </span>
                <span class="disabled-dislike">
                    <img src="/statics/image/dislike.png" alt="Dislike" />
                    <span class="count">${post.dislikes}</span>
                </span>
                `}
            </div>
        </article>

        <div class="card new-reply">
            ${isAuthenticated ? `
            <form method="POST" action="/reply">
                <input type="hidden" name="post_id" value="${post.id}" />
                <label>Commentaire
                    <textarea name="content" rows="4" placeholder="Write your reply..." required></textarea>
                </label>
                <div style="margin-top:8px;">
                    <button type="submit" class="button">Post reply</button>
                </div>
            </form>
            ` : `
            <p class="small muted-small">You must <a href="#" onclick="openAuthModal('login'); return false;">log in</a> to reply.</p>
            `}
        </div>

        <div class="replies-header">
            REPLIES: ${replies.length}
        </div>

        <div class="replies-scroll-wrapper">
            ${repliesHtml}
        </div>
    `;
}

function renderReply(reply, data) {
    const likedComment = Boolean(data.likedComments && data.likedComments[String(reply.id)]);
    const dislikedComment = Boolean(data.dislikedComments && data.dislikedComments[String(reply.id)]);
    const isAuthenticated = Boolean(data.username);

    return `
        <article class="reply" data-id="${reply.id}">
            <div class="reply-header">
                <strong>${escapeHtml(reply.username)}</strong>
                <span class="reply-date">${escapeHtml(reply.createdAt)}</span>
            </div>
            <div class="reply-content">${escapeHtml(reply.content)}</div>

            <div class="post-actions">
                ${isAuthenticated ? `
                <form method="POST" action="/like" style="display:inline;">
                    <input type="hidden" name="comment_id" value="${reply.id}">
                    <button type="submit" class="like-btn${likedComment ? ' active' : ''}">
                        <img src="/statics/image/like.png" alt="Like" />
                        <span class="count">${reply.likes}</span>
                    </button>
                </form>

                <form method="POST" action="/dislike" style="display:inline;">
                    <input type="hidden" name="comment_id" value="${reply.id}">
                    <button type="submit" class="dislike-btn${dislikedComment ? ' active' : ''}">
                        <img src="/statics/image/dislike.png" alt="Dislike" />
                        <span class="count">${reply.dislikes}</span>
                    </button>
                </form>
                ` : `
                <span class="disabled-like">
                    <img src="/statics/image/like.png" alt="Like" />
                    <span class="count">${reply.likes}</span>
                </span>
                <span class="disabled-dislike">
                    <img src="/statics/image/dislike.png" alt="Dislike" />
                    <span class="count">${reply.dislikes}</span>
                </span>
                `}
            </div>
        </article>
    `;
}

function escapeHtml(value) {
    if (value === null || value === undefined) return '';
    return String(value)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
}
