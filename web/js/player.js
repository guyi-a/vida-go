const Player = (() => {
    let videos = [];
    let currentIndex = 0;
    let isTransitioning = false;
    let touchStartY = 0;
    let touchDeltaY = 0;
    let container = null;
    let page = 1;
    let loading = false;
    let hasMore = true;
    let likedSet = new Set();

    function init() {
        container = document.getElementById('videoContainer');
        bindEvents();
        loadVideos();
    }

    function bindEvents() {
        container.addEventListener('touchstart', onTouchStart, { passive: true });
        container.addEventListener('touchmove', onTouchMove, { passive: true });
        container.addEventListener('touchend', onTouchEnd);
        container.addEventListener('wheel', onWheel, { passive: false });
    }

    function onTouchStart(e) {
        if (isTransitioning) return;
        touchStartY = e.touches[0].clientY;
        touchDeltaY = 0;
    }

    function onTouchMove(e) {
        if (isTransitioning) return;
        touchDeltaY = e.touches[0].clientY - touchStartY;
    }

    function onTouchEnd() {
        if (isTransitioning) return;
        const threshold = 80;
        if (touchDeltaY < -threshold) goNext();
        else if (touchDeltaY > threshold) goPrev();
        touchDeltaY = 0;
    }

    function onWheel(e) {
        e.preventDefault();
        if (isTransitioning) return;
        if (e.deltaY > 30) goNext();
        else if (e.deltaY < -30) goPrev();
    }

    function goNext() {
        if (currentIndex >= videos.length - 1) return;
        isTransitioning = true;
        currentIndex++;
        updateSlides();
        if (currentIndex >= videos.length - 2 && hasMore) loadVideos();
        setTimeout(() => { isTransitioning = false; }, 400);
    }

    function goPrev() {
        if (currentIndex <= 0) return;
        isTransitioning = true;
        currentIndex--;
        updateSlides();
        setTimeout(() => { isTransitioning = false; }, 400);
    }

    async function loadVideos() {
        if (loading || !hasMore) return;
        loading = true;
        document.getElementById('videoLoading').style.display =
            videos.length === 0 ? 'block' : 'none';
        try {
            const res = await API.Video.feed(page, 5);
            const list = res.data?.videos || [];
            if (list.length === 0) { hasMore = false; return; }
            const startIdx = videos.length;
            videos.push(...list);
            page++;

            if (API.getToken() && list.length > 0) {
                try {
                    const ids = list.map(v => v.id);
                    const statusRes = await API.Favorite.batchStatus(ids);
                    const statuses = statusRes.data?.favorites_status || statusRes.data || {};
                    for (const [id, liked] of Object.entries(statuses)) {
                        if (liked) likedSet.add(Number(id));
                    }
                } catch {}
            }

            for (let i = startIdx; i < videos.length; i++) {
                createSlide(videos[i], i);
            }
            if (startIdx === 0) updateSlides();
        } catch (e) {
            Toast.show(e.message, 'error');
        } finally {
            loading = false;
            document.getElementById('videoLoading').style.display = 'none';
        }
    }

    function createSlide(video, idx) {
        const slide = document.createElement('div');
        slide.className = 'video-slide';
        slide.dataset.index = idx;
        slide.dataset.videoId = video.id;

        const isLiked = likedSet.has(video.id);

        slide.innerHTML = `
            <video src="${video.play_url}" loop playsinline preload="none" poster="${video.cover_url || ''}"></video>
            <div class="play-icon"><i class="fas fa-play"></i></div>
            <div class="video-info">
                <div class="author">@${video.author?.username || video.author_name || '用户'}</div>
                <div class="desc">${video.title || ''}${video.description ? ' · ' + video.description : ''}</div>
            </div>
            <div class="video-actions">
                <div class="action-avatar" data-author-id="${video.author_id || video.author?.id || ''}">
                    <img src="${video.author?.avatar || ''}" onerror="this.parentElement.innerHTML='<i class=\\'fas fa-user\\'></i>'">
                </div>
                <button class="action-btn like-btn ${isLiked ? 'liked' : ''}" data-video-id="${video.id}">
                    <i class="${isLiked ? 'fas' : 'far'} fa-heart"></i>
                    <span>${formatCount(video.favorite_count || 0)}</span>
                </button>
                <button class="action-btn comment-btn" data-video-id="${video.id}">
                    <i class="fas fa-comment-dots"></i>
                    <span>${formatCount(video.comment_count || 0)}</span>
                </button>
                <button class="action-btn share-btn">
                    <i class="fas fa-share"></i>
                    <span>分享</span>
                </button>
            </div>
        `;

        const vid = slide.querySelector('video');
        vid.addEventListener('click', () => togglePlay(slide, vid));

        slide.querySelector('.desc').addEventListener('click', (e) => {
            e.stopPropagation();
            e.target.classList.toggle('expanded');
        });

        slide.querySelector('.like-btn').addEventListener('click', (e) => {
            e.stopPropagation();
            handleLike(slide, video);
        });

        slide.querySelector('.comment-btn').addEventListener('click', (e) => {
            e.stopPropagation();
            UI.openComments(video.id);
        });

        slide.querySelector('.share-btn').addEventListener('click', (e) => {
            e.stopPropagation();
            Toast.show('分享功能开发中');
        });

        const avatarEl = slide.querySelector('.action-avatar');
        avatarEl.addEventListener('click', (e) => {
            e.stopPropagation();
            const authorId = avatarEl.dataset.authorId;
            if (authorId) App.showUserProfile(authorId);
        });

        vid.addEventListener('dblclick', (e) => {
            e.preventDefault();
            if (!likedSet.has(video.id)) handleLike(slide, video);
            showLikeAnimation(slide);
        });

        container.appendChild(slide);
    }

    function togglePlay(slide, vid) {
        if (vid.paused) {
            vid.play();
            slide.classList.remove('paused');
        } else {
            vid.pause();
            slide.classList.add('paused');
        }
    }

    function updateSlides() {
        const slides = container.querySelectorAll('.video-slide');
        slides.forEach((s, i) => {
            const vid = s.querySelector('video');
            const offset = i - currentIndex;
            s.style.transform = `translateY(${offset * 100}%)`;

            if (i === currentIndex) {
                if (vid.preload === 'none') vid.preload = 'auto';
                vid.play().catch(() => {});
                s.classList.remove('paused');
            } else {
                vid.pause();
                vid.currentTime = 0;
            }
        });

        if (currentIndex + 1 < slides.length) {
            const nextVid = slides[currentIndex + 1].querySelector('video');
            if (nextVid.preload === 'none') nextVid.preload = 'auto';
        }
    }

    async function handleLike(slide, video) {
        if (!API.getToken()) {
            App.showLogin();
            return;
        }
        const btn = slide.querySelector('.like-btn');
        const countEl = btn.querySelector('span');
        const isLiked = likedSet.has(video.id);

        try {
            if (isLiked) {
                await API.Favorite.unlike(video.id);
                likedSet.delete(video.id);
                btn.classList.remove('liked');
                btn.querySelector('i').className = 'far fa-heart';
                video.favorite_count = Math.max(0, (video.favorite_count || 1) - 1);
            } else {
                await API.Favorite.like(video.id);
                likedSet.add(video.id);
                btn.classList.add('liked');
                btn.querySelector('i').className = 'fas fa-heart';
                video.favorite_count = (video.favorite_count || 0) + 1;
            }
            countEl.textContent = formatCount(video.favorite_count);
        } catch (e) {
            Toast.show(e.message, 'error');
        }
    }

    function showLikeAnimation(slide) {
        const heart = document.createElement('div');
        heart.className = 'like-animation';
        heart.innerHTML = '<i class="fas fa-heart"></i>';
        slide.appendChild(heart);
        setTimeout(() => heart.remove(), 900);
    }

    function formatCount(n) {
        if (n >= 10000) return (n / 10000).toFixed(1) + 'w';
        if (n >= 1000) return (n / 1000).toFixed(1) + 'k';
        return String(n);
    }

    function reset() {
        videos = [];
        currentIndex = 0;
        page = 1;
        hasMore = true;
        loading = false;
        likedSet.clear();
        if (container) container.innerHTML = '';
    }

    function refresh() {
        reset();
        loadVideos();
    }

    function pauseCurrent() {
        const slides = container?.querySelectorAll('.video-slide');
        if (slides && slides[currentIndex]) {
            const vid = slides[currentIndex].querySelector('video');
            if (vid) vid.pause();
        }
    }

    function resumeCurrent() {
        const slides = container?.querySelectorAll('.video-slide');
        if (slides && slides[currentIndex]) {
            const vid = slides[currentIndex].querySelector('video');
            if (vid) vid.play().catch(() => {});
        }
    }

    function playSpecificVideo(videoId) {
        const idx = videos.findIndex(v => v.id === videoId);
        if (idx >= 0) {
            currentIndex = idx;
            updateSlides();
            App.navigate('home');
        }
    }

    return { init, refresh, pauseCurrent, resumeCurrent, playSpecificVideo, reset };
})();
