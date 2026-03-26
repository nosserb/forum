// Filter Management
document.addEventListener('DOMContentLoaded', function() {
    const filterButtons = document.querySelectorAll('.filter-btn');
    let activeFilter = null;

    filterButtons.forEach(button => {
        button.addEventListener('click', function() {
            const filterType = this.classList[1];
            
            if (activeFilter === filterType) {
                activeFilter = null;
                applyFilter(null);
                filterButtons.forEach(btn => btn.classList.remove('active'));
                return;
            }

            filterButtons.forEach(btn => btn.classList.remove('active'));

            this.classList.add('active');

            activeFilter = filterType;

            applyFilter(filterType);
        });
    });

    function applyFilter(filterType) {
        if (!filterType || filterType === 'filter-all') {
            loadFiltered('/');
            return;
        }

        let params = new URLSearchParams();

        switch(filterType) {
            case 'filter-posted':
                params.append('created', 'on');
                break;
            case 'filter-liked':
                params.append('liked', 'on');
                break;
            case 'filter-disliked':
                params.append('disliked', 'on');
                break;
            case 'filter-commented':
                params.append('commented', 'on');
                break;
        }

        loadFiltered('/filter?' + params.toString());
    }

    function loadFiltered(url) {
        const postsList = document.querySelector('.posts-list');
        if (!postsList) {
            return;
        }

        fetch(url, {
            headers: {
                'X-Requested-With': 'XMLHttpRequest'
            }
        })
            .then(response => response.text())
            .then(html => {
                const parser = new DOMParser();
                const doc = parser.parseFromString(html, 'text/html');
                const newPostsList = doc.querySelector('.posts-list');
                if (!newPostsList) {
                    return;
                }

                postsList.innerHTML = newPostsList.innerHTML;
                document.dispatchEvent(new Event('likes:refresh'));

                if (url === '/') {
                    window.history.replaceState({}, '', '/');
                } else {
                    window.history.replaceState({}, '', url);
                }
            })
            .catch(error => {
                console.error('Error:', error);
            });
    }

    const refreshIntervalMs = 5000;
    setInterval(() => {
        const detailView = document.getElementById('post-detail-view');
        if (detailView && !detailView.classList.contains('is-hidden')) {
            return;
        }

        applyFilter(activeFilter);
    }, refreshIntervalMs);
});
