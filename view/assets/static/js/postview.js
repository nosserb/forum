document.addEventListener('DOMContentLoaded', function() {
    setupDeletePostModal();
    setupDeleteCommentModal();
    const editPostModalApi = setupEditPostModal();
    const editCommentModalApi = setupEditCommentModal();

    document.addEventListener('click', function(e) {
        const editButton = e.target.closest('.js-edit-post-btn');
        if (editButton) {
            e.preventDefault();
            const postId = editButton.dataset.postId;
            const rawTitle = editButton.dataset.postTitle || '';
            const rawContent = editButton.dataset.postContent || '';

            let currentTitle = rawTitle;
            let currentContent = rawContent;
            try {
                currentTitle = decodeURIComponent(rawTitle);
            } catch (_) { /* not encoded, use as-is */ }
            try {
                currentContent = decodeURIComponent(rawContent);
            } catch (_) { /* not encoded, use as-is */ }

            if (editPostModalApi && typeof editPostModalApi.open === 'function') {
                editPostModalApi.open(postId, currentTitle, currentContent);
            }

            const optionsRoot = editButton.closest('.post-options');
            if (optionsRoot) {
                optionsRoot.removeAttribute('open');
            }
            return;
        }

        const editCommentButton = e.target.closest('.js-edit-comment-btn');
        if (editCommentButton) {
            e.preventDefault();
            const commentId = editCommentButton.dataset.commentId;
            const postId = editCommentButton.dataset.postId;
            const rawContent = editCommentButton.dataset.commentContent || '';

            let currentContent = '';
            try {
                currentContent = decodeURIComponent(rawContent);
            } catch (_) {
                currentContent = '';
            }

            if (editCommentModalApi && typeof editCommentModalApi.open === 'function') {
                editCommentModalApi.open(commentId, postId, currentContent);
            }

            const optionsRoot = editCommentButton.closest('.post-options');
            if (optionsRoot) {
                optionsRoot.removeAttribute('open');
            }
            return;
        }

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
    const canDeletePost = Boolean(data.username) && String(data.username) === String(post.username);
    const likedPost = Boolean(data.likedPosts && data.likedPosts[String(postId)]);
    const dislikedPost = Boolean(data.dislikedPosts && data.dislikedPosts[String(postId)]);

    const categoriesHtml = Array.isArray(post.categories) && post.categories.length
        ? `<div class="post-cats" style="margin-bottom:12px;">${post.categories.map(cat => `<span class="cat-badge">${escapeHtml(cat)}</span>`).join('')}</div>`
        : '';

    const replies = Array.isArray(post.replies) ? post.replies : [];
    const repliesHtml = replies.length
        ? replies.map(reply => renderReply(reply, data, post.id)).join('')
        : '<p class="muted-small">No replies yet. Be the first!</p>';

    const postOptionsHtml = canDeletePost ? renderPostOptionsMenu(post) : '';

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
                ${postOptionsHtml ? `<div class="post-actions-right">${postOptionsHtml}</div>` : ''}
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

function renderPostOptionsMenu(post) {
    return `
        <details class="post-options">
            <summary class="post-options-trigger" aria-label="Post options">
                <img src="/statics/image/engrenage.png" alt="Options" class="post-options-icon">
            </summary>
            <div class="post-options-menu">
                <button
                    type="button"
                    class="post-option js-edit-post-btn"
                    data-post-id="${post.id}"
                    data-post-title="${encodeURIComponent(post.title || '')}"
                    data-post-content="${encodeURIComponent(post.content || '')}">
                    Edit
                </button>
                <form method="POST" action="/post" class="js-delete-post-form" data-post-title="${escapeHtml(post.title)}">
                    <input type="hidden" name="action" value="delete">
                    <input type="hidden" name="post_id" value="${post.id}">
                    <button type="submit" class="post-option post-option-danger">Delete</button>
                </form>
            </div>
        </details>
    `;
}

function renderCommentOptionsMenu(reply, postId) {
    return `
        <details class="post-options">
            <summary class="post-options-trigger" aria-label="Comment options">
                <img src="/statics/image/engrenage.png" alt="Options" class="post-options-icon">
            </summary>
            <div class="post-options-menu">
                <button
                    type="button"
                    class="post-option js-edit-comment-btn"
                    data-comment-id="${reply.id}"
                    data-post-id="${postId}"
                    data-comment-content="${encodeURIComponent(reply.content || '')}">
                    Edit
                </button>
                <form method="POST" action="/reply" class="js-delete-comment-form" data-comment-content="${escapeHtml(reply.content)}">
                    <input type="hidden" name="post_id" value="${postId}" />
                    <input type="hidden" name="action" value="delete">
                    <input type="hidden" name="comment_id" value="${reply.id}">
                    <button type="submit" class="post-option post-option-danger">Delete</button>
                </form>
            </div>
        </details>
    `;
}

function setupEditPostModal() {
    const modal = document.getElementById('editPostModal');
    const form = document.getElementById('editPostForm');
    const postIdInput = document.getElementById('editPostIdInput');
    const titleInput = document.getElementById('editPostTitleInput');
    const contentInput = document.getElementById('editPostContentInput');
    const errorTarget = document.getElementById('editPostError');
    const closeBtn = document.getElementById('editPostCloseBtn');
    const cancelBtn = document.getElementById('editPostCancelBtn');

    if (!modal || !form || !postIdInput || !titleInput || !contentInput) {
        return null;
    }

    function open(postId, currentTitle, currentContent) {
        postIdInput.value = String(postId || '');
        titleInput.value = currentTitle || '';
        contentInput.value = currentContent || '';
        if (errorTarget) {
            errorTarget.textContent = '';
        }

        modal.classList.add('active');
        modal.setAttribute('aria-hidden', 'false');
        titleInput.focus();
        titleInput.select();
    }

    function close() {
        modal.classList.remove('active');
        modal.setAttribute('aria-hidden', 'true');
    }

    form.addEventListener('submit', function(e) {
        const title = (titleInput.value || '').trim();
        const content = (contentInput.value || '').trim();

        if (!title || !content) {
            e.preventDefault();
            if (errorTarget) {
                errorTarget.textContent = 'Title and content cannot be empty.';
            }
            return;
        }

        if (title.length > 100) {
            e.preventDefault();
            if (errorTarget) {
                errorTarget.textContent = 'Title must be 100 characters or fewer.';
            }
            return;
        }

        if (content.length > 7500) {
            e.preventDefault();
            if (errorTarget) {
                errorTarget.textContent = 'Content must be 7500 characters or fewer.';
            }
            return;
        }
    });

    if (closeBtn) {
        closeBtn.addEventListener('click', close);
    }

    if (cancelBtn) {
        cancelBtn.addEventListener('click', close);
    }

    modal.addEventListener('click', function(e) {
        if (e.target === modal) {
            close();
        }
    });

    document.addEventListener('keydown', function(e) {
        if (e.key === 'Escape' && modal.classList.contains('active')) {
            close();
        }
    });

    return { open: open, close: close };
}

function setupEditCommentModal() {
    const modal = document.getElementById('editCommentModal');
    const form = document.getElementById('editCommentForm');
    const commentIdInput = document.getElementById('editCommentIdInput');
    const postIdInput = document.getElementById('editCommentPostIdInput');
    const contentInput = document.getElementById('editCommentContentInput');
    const errorTarget = document.getElementById('editCommentError');
    const closeBtn = document.getElementById('editCommentCloseBtn');
    const cancelBtn = document.getElementById('editCommentCancelBtn');

    if (!modal || !form || !commentIdInput || !postIdInput || !contentInput) {
        return null;
    }

    function open(commentId, postId, currentContent) {
        commentIdInput.value = String(commentId || '');
        postIdInput.value = String(postId || '');
        contentInput.value = currentContent || '';
        if (errorTarget) {
            errorTarget.textContent = '';
        }

        modal.classList.add('active');
        modal.setAttribute('aria-hidden', 'false');
        contentInput.focus();
        contentInput.select();
    }

    function close() {
        modal.classList.remove('active');
        modal.setAttribute('aria-hidden', 'true');
    }

    form.addEventListener('submit', function(e) {
        const content = (contentInput.value || '').trim();

        if (!content) {
            e.preventDefault();
            if (errorTarget) {
                errorTarget.textContent = 'Content cannot be empty.';
            }
            return;
        }

        if (content.length > 7500) {
            e.preventDefault();
            if (errorTarget) {
                errorTarget.textContent = 'Content must be 7500 characters or fewer.';
            }
            return;
        }
    });

    if (closeBtn) {
        closeBtn.addEventListener('click', close);
    }

    if (cancelBtn) {
        cancelBtn.addEventListener('click', close);
    }

    modal.addEventListener('click', function(e) {
        if (e.target === modal) {
            close();
        }
    });

    document.addEventListener('keydown', function(e) {
        if (e.key === 'Escape' && modal.classList.contains('active')) {
            close();
        }
    });

    return { open: open, close: close };
}

function setupDeletePostModal() {
    const modal = document.getElementById('deletePostModal');
    if (!modal) {
        return;
    }

    const cancelBtn = document.getElementById('deletePostCancelBtn');
    const confirmBtn = document.getElementById('deletePostConfirmBtn');
    const titleTarget = document.getElementById('deletePostModalPostTitle');

    let pendingDeleteForm = null;

    function openDeletePostModal(form) {
        pendingDeleteForm = form;

        const postTitle = form.dataset.postTitle || '';
        titleTarget.textContent = postTitle ? `Post: "${postTitle}"` : '';

        modal.classList.add('active');
        modal.setAttribute('aria-hidden', 'false');

        if (confirmBtn) {
            confirmBtn.focus();
        }
    }

    function closeDeletePostModal() {
        modal.classList.remove('active');
        modal.setAttribute('aria-hidden', 'true');
        pendingDeleteForm = null;
    }

    document.addEventListener('submit', function(e) {
        const form = e.target.closest('.js-delete-post-form');
        if (!form) {
            return;
        }

        e.preventDefault();
        openDeletePostModal(form);
    });

    if (cancelBtn) {
        cancelBtn.addEventListener('click', closeDeletePostModal);
    }

    if (confirmBtn) {
        confirmBtn.addEventListener('click', function() {
            if (!pendingDeleteForm) {
                closeDeletePostModal();
                return;
            }

            const formToSubmit = pendingDeleteForm;
            closeDeletePostModal();
            formToSubmit.submit();
        });
    }

    modal.addEventListener('click', function(e) {
        if (e.target === modal) {
            closeDeletePostModal();
        }
    });

    document.addEventListener('keydown', function(e) {
        if (e.key === 'Escape' && modal.classList.contains('active')) {
            closeDeletePostModal();
        }
    });
}

function setupDeleteCommentModal() {
    const modal = document.getElementById('deleteCommentModal');
    if (!modal) {
        return;
    }

    const cancelBtn = document.getElementById('deleteCommentCancelBtn');
    const confirmBtn = document.getElementById('deleteCommentConfirmBtn');
    const contentTarget = document.getElementById('deleteCommentModalCommentContent');

    let pendingDeleteForm = null;

    function openDeleteCommentModal(form) {
        pendingDeleteForm = form;

        const commentContent = form.dataset.commentContent || '';
        contentTarget.textContent = commentContent ? `Comment: "${commentContent}"` : '';

        modal.classList.add('active');
        modal.setAttribute('aria-hidden', 'false');

        if (confirmBtn) {
            confirmBtn.focus();
        }
    }

    function closeDeleteCommentModal() {
        modal.classList.remove('active');
        modal.setAttribute('aria-hidden', 'true');
        pendingDeleteForm = null;
    }

    document.addEventListener('submit', function(e) {
        const form = e.target.closest('.js-delete-comment-form');
        if (!form) {
            return;
        }

        e.preventDefault();
        openDeleteCommentModal(form);
    });

    if (cancelBtn) {
        cancelBtn.addEventListener('click', closeDeleteCommentModal);
    }

    if (confirmBtn) {
        confirmBtn.addEventListener('click', function() {
            if (!pendingDeleteForm) {
                closeDeleteCommentModal();
                return;
            }

            const formToSubmit = pendingDeleteForm;
            closeDeleteCommentModal();
            formToSubmit.submit();
        });
    }

    modal.addEventListener('click', function(e) {
        if (e.target === modal) {
            closeDeleteCommentModal();
        }
    });

    document.addEventListener('keydown', function(e) {
        if (e.key === 'Escape' && modal.classList.contains('active')) {
            closeDeleteCommentModal();
        }
    });
}

function renderReply(reply, data, postId) {
    const likedComment = Boolean(data.likedComments && data.likedComments[String(reply.id)]);
    const dislikedComment = Boolean(data.dislikedComments && data.dislikedComments[String(reply.id)]);
    const isAuthenticated = Boolean(data.username);
    const canManageComment = isAuthenticated && String(data.username) === String(reply.username);
    const commentOptionsHtml = canManageComment ? renderCommentOptionsMenu(reply, postId) : '';

    return `
        <article class="reply" data-id="${reply.id}">
            <div class="reply-header">
                <strong>${escapeHtml(reply.username)}</strong>
                <div class="reply-header-meta">
                    <span class="reply-date">${escapeHtml(reply.createdAt)}</span>
                    ${commentOptionsHtml}
                </div>
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
