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

