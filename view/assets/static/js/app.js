// Authentication Modal Functions
function switchTab(tab) {
    const loginForm = document.getElementById('loginForm');
    const signupForm = document.getElementById('signupForm');
    const loginTabBtn = document.getElementById('loginTabBtn');
    const signupTabBtn = document.getElementById('signupTabBtn');

    if (tab === 'login') {
        loginForm.classList.add('active');
        signupForm.classList.remove('active');
        loginTabBtn.classList.add('active');
        signupTabBtn.classList.remove('active');
    } else {
        loginForm.classList.remove('active');
        signupForm.classList.add('active');
        loginTabBtn.classList.remove('active');
        signupTabBtn.classList.add('active');
    }
}

function openAuthModal(tab) {
    const authModal = document.getElementById('authModal');
    authModal.classList.add('active');
    switchTab(tab);
}

function closeAuthModal() {
    const authModal = document.getElementById('authModal');
    authModal.classList.remove('active');
}

// Error Modal Functions
function showErrorModal(title, message) {
    const errorModal = document.getElementById('errorModal');
    document.getElementById('errorTitle').textContent = title;
    document.getElementById('errorMessage').textContent = message;
    errorModal.classList.add('active');
}

function closeErrorModal() {
    const errorModal = document.getElementById('errorModal');
    errorModal.classList.remove('active');
}

function fillProfileModal(profileInfo) {
    if (!profileInfo) {
        return;
    }

    const profileFields = {
        Username: profileInfo.username,
        FirstName: profileInfo.firstName,
        LastName: profileInfo.lastName,
        Email: profileInfo.email,
        Age: profileInfo.age,
    };

    Object.entries(profileFields).forEach(([key, value]) => {
        const target = document.getElementById(`profileInfo${key}`);
        if (target) {
            const trimmed = String(value || '').trim();
            target.textContent = trimmed !== '' ? trimmed : '-';
        }
    });
}
// T

let profileBoardRequestToken = 0;

function loadProfilePostBoard() {
    const token = ++profileBoardRequestToken;

    renderProfilePostColumnState('profileCreatedPosts', 'profileCreatedCount', 'Loading posts...');
    renderProfilePostColumnState('profileCommentedPosts', 'profileCommentedCount', 'Loading posts...');
    renderProfilePostColumnState('profileLikedPosts', 'profileLikedCount', 'Loading posts...');
    renderProfilePostColumnState('profileDislikedPosts', 'profileDislikedCount', 'Loading posts...');

    Promise.all([
        fetchProfilePosts('/filter?created=on'),
        fetchProfilePosts('/filter?commented=on&profile=1'),
        fetchProfilePosts('/filter?liked=on'),
        fetchProfilePosts('/filter?disliked=on'),
    ])
        .then(([createdPosts, commentedPosts, likedPosts, dislikedPosts]) => {
            if (token !== profileBoardRequestToken) {
                return;
            }

            renderProfilePostColumn('profileCreatedPosts', 'profileCreatedCount', createdPosts, {
                emptyLabel: 'No created posts yet.',
                showAuthor: false,
            });
            renderProfilePostColumn('profileCommentedPosts', 'profileCommentedCount', commentedPosts, {
                emptyLabel: 'No commented posts yet.',
                showAuthor: true,
            });
            renderProfilePostColumn('profileLikedPosts', 'profileLikedCount', likedPosts, {
                emptyLabel: 'No liked posts yet.',
                showAuthor: true,
            });
            renderProfilePostColumn('profileDislikedPosts', 'profileDislikedCount', dislikedPosts, {
                emptyLabel: 'No disliked posts yet.',
                showAuthor: true,
            });
        })
        .catch(() => {
            if (token !== profileBoardRequestToken) {
                return;
            }

            renderProfilePostColumnState('profileCreatedPosts', 'profileCreatedCount', 'Unable to load posts.');
            renderProfilePostColumnState('profileCommentedPosts', 'profileCommentedCount', 'Unable to load posts.');
            renderProfilePostColumnState('profileLikedPosts', 'profileLikedCount', 'Unable to load posts.');
            renderProfilePostColumnState('profileDislikedPosts', 'profileDislikedCount', 'Unable to load posts.');
        });
}

function fetchProfilePosts(url) {
    return fetch(url, {
        headers: {
            'X-Requested-With': 'XMLHttpRequest'
        }
    })
        .then((response) => response.text())
        .then((html) => extractProfilePostsFromHtml(html));
}

function extractProfilePostsFromHtml(html) {
    const parser = new DOMParser();
    const doc = parser.parseFromString(html, 'text/html');
    const postElements = doc.querySelectorAll('.posts-list .post');

    return Array.from(postElements).map((postEl) => {
        const link = postEl.querySelector('.post-link');
        const title = postEl.querySelector('.post-link h4');
        const content = postEl.querySelector('.post-content p');
        const author = postEl.querySelector('.post-author');
        const date = postEl.querySelector('.post-date');
        const counts = postEl.querySelectorAll('.count');

        return {
            href: link ? link.getAttribute('href') : '#',
            title: title ? title.textContent.trim() : 'Sans titre',
            content: content ? content.textContent.trim() : '',
            author: author ? author.textContent.trim() : '',
            date: date ? date.textContent.trim() : '-',
            likes: counts[0] ? counts[0].textContent.trim() : '0',
            dislikes: counts[1] ? counts[1].textContent.trim() : '0',
        };
    });
}

function renderProfilePostColumn(containerId, countId, posts, options) {
    const container = document.getElementById(containerId);
    const count = document.getElementById(countId);
    if (!container || !count) {
        return;
    }

    count.textContent = String(posts.length);
    container.innerHTML = '';

    if (!posts.length) {
        const empty = document.createElement('div');
        empty.className = 'profile-post-empty';
        empty.textContent = options.emptyLabel;
        container.appendChild(empty);
        return;
    }

    posts.forEach((post) => {
        const card = document.createElement('a');
        card.className = 'profile-post-card';
        card.href = post.href || '#';

        const title = document.createElement('strong');
        title.textContent = post.title;

        const content = document.createElement('p');
        content.textContent = post.content;

        const meta = document.createElement('div');
        meta.className = 'profile-post-meta';

        const left = document.createElement('span');
        left.textContent = options.showAuthor ? `par ${post.author || '-'}` : post.date;

        const right = document.createElement('span');
        right.textContent = `${post.likes} like · ${post.dislikes} dislike`;

        meta.appendChild(left);
        meta.appendChild(right);
        card.appendChild(title);
        card.appendChild(content);
        card.appendChild(meta);
        container.appendChild(card);
    });
}

function renderProfilePostColumnState(containerId, countId, label) {
    const container = document.getElementById(containerId);
    const count = document.getElementById(countId);
    if (!container || !count) {
        return;
    }

    count.textContent = '0';
    container.innerHTML = `<div class="profile-post-empty">${label}</div>`;
}
// T
function openProfileModal() {
    const profileModal = document.getElementById('profileModal');
    if (!profileModal) {
        return;
    }

    profileModal.classList.add('active');
    profileModal.setAttribute('aria-hidden', 'false');
    loadProfilePostBoard();
}

function closeProfileModal() {
    const profileModal = document.getElementById('profileModal');
    if (!profileModal) {
        return;
    }

    profileModal.classList.remove('active');
    profileModal.setAttribute('aria-hidden', 'true');
}

function toggleProfileModal(event) {
    if (event) {
        event.preventDefault();
        event.stopPropagation();
    }

    const profileModal = document.getElementById('profileModal');
    if (!profileModal) {
        return;
    }

    if (profileModal.classList.contains('active')) {
        closeProfileModal();
        return;
    }

    openProfileModal();
}

// User Functions
function logout() {
    window.location.href = '/logout';
}

// Comments Toggle Function
function toggleComments(event, postId) {
    event.preventDefault();
    const commentsSection = document.getElementById(`comments-${postId}`);
    if (commentsSection) {
        if (commentsSection.style.display === 'none') {
            commentsSection.style.display = 'block';
            event.target.classList.add('active');
        } else {
            commentsSection.style.display = 'none';
            event.target.classList.remove('active');
        }
    }
}

// Post Form AJAX Handler
document.addEventListener('DOMContentLoaded', function() {
    const profileToggle = document.getElementById('profileToggle');
    const profileModal = document.getElementById('profileModal');
    if (profileToggle && profileModal) {
        const initialUsername = profileToggle.dataset.username || '';
        fillProfileModal({ username: initialUsername });

        if (window.currentUserInfo) {
            fillProfileModal(window.currentUserInfo);
        }

        document.addEventListener('forum:userinfo', function(event) {
            const payload = event.detail || {};
            const currentUserID = Number(window.USER_ID);
            if (Number(payload.userID) === currentUserID) {
                fillProfileModal(payload);
            }
        });

        document.addEventListener('likes:refresh', function() {
            if (profileModal.classList.contains('active')) {
                loadProfilePostBoard();
            }
        });

        profileModal.addEventListener('click', function(e) {
            if (e.target === profileModal) {
                closeProfileModal();
            }
        });

        document.addEventListener('keydown', function(e) {
            if (e.key === 'Escape' && profileModal.classList.contains('active')) {
                closeProfileModal();
            }
        });
    }

    const authModal = document.getElementById('authModal');
    if (authModal) {
        authModal.addEventListener('click', function(e) {
            if (e.target === authModal) {
                closeAuthModal();
            }
        });
    }

    const errorModal = document.getElementById('errorModal');
    if (errorModal) {
        errorModal.addEventListener('click', function(e) {
            if (e.target === errorModal) {
                closeErrorModal();
            }
        });
    }

    // Intercept the post form to reload after creation
    const createPostForm = document.getElementById('createPostForm');
    if (createPostForm) {
        createPostForm.addEventListener('submit', function(e) {
            e.preventDefault();
            
            const formData = new FormData(this);
            
            fetch('/post', {
                method: 'POST',
                body: formData
            })
            .then(response => {
                // Reload page to show new post
                window.location.reload();
            })
            .catch(error => {
                console.error('Error:', error);
            });
        });
    }
});

