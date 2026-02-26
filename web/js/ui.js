/* ===== Toast ===== */
const Toast = {
    show(msg, type = 'info') {
        const el = document.createElement('div');
        el.className = `toast toast-${type}`;
        el.textContent = msg;
        document.getElementById('toastContainer').appendChild(el);
        setTimeout(() => el.remove(), 2500);
    }
};

/* ===== UI Module ===== */
const UI = (() => {

    /* ---------- Search ---------- */
    let searchSort = 'relevance';

    function initSearch() {
        document.getElementById('searchBtn').addEventListener('click', doSearch);
        document.getElementById('searchInput').addEventListener('keydown', (e) => {
            if (e.key === 'Enter') doSearch();
        });
        document.querySelectorAll('.sort-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                document.querySelectorAll('.sort-btn').forEach(b => b.classList.remove('active'));
                btn.classList.add('active');
                searchSort = btn.dataset.sort;
                doSearch();
            });
        });
    }

    async function doSearch() {
        const q = document.getElementById('searchInput').value.trim();
        if (!q) return;
        const results = document.getElementById('searchResults');
        results.innerHTML = '<div class="empty-state"><div class="spinner"></div><p>搜索中...</p></div>';
        try {
            const res = await API.Search.videos(q, searchSort, 1, 20);
            const videos = res.data?.videos || [];
            if (videos.length === 0) {
                results.innerHTML = '<div class="empty-state"><i class="fas fa-video-slash"></i><p>没有找到相关视频</p></div>';
                return;
            }
            results.innerHTML = '<div class="search-grid"></div>';
            const grid = results.querySelector('.search-grid');
            videos.forEach(v => grid.appendChild(createSearchCard(v)));
        } catch (e) {
            results.innerHTML = `<div class="empty-state"><i class="fas fa-exclamation-circle"></i><p>${e.message}</p></div>`;
        }
    }

    function createSearchCard(video) {
        const card = document.createElement('div');
        card.className = 'search-card';
        card.innerHTML = `
            <img src="${video.cover_url || ''}" alt="" onerror="this.style.display='none'">
            <div class="card-overlay">
                <div class="card-title">${video.title || '无标题'}</div>
                <div class="card-stats"><i class="fas fa-play"></i> ${formatCount(video.view_count || 0)}</div>
            </div>
        `;
        card.addEventListener('click', () => {
            Player.playSpecificVideo(video.id);
        });
        return card;
    }

    /* ---------- Upload ---------- */
    function initUpload() {
        const area = document.getElementById('uploadArea');
        const fileInput = document.getElementById('videoFile');
        const preview = document.getElementById('uploadPreview');
        const previewVid = document.getElementById('previewVideo');
        const titleInput = document.getElementById('videoTitle');
        const form = document.getElementById('uploadForm');

        area.addEventListener('click', () => {
            if (!API.getToken()) { App.showLogin(); return; }
            fileInput.click();
        });

        area.addEventListener('dragover', (e) => { e.preventDefault(); area.style.borderColor = 'var(--accent)'; });
        area.addEventListener('dragleave', () => { area.style.borderColor = ''; });
        area.addEventListener('drop', (e) => {
            e.preventDefault();
            area.style.borderColor = '';
            if (e.dataTransfer.files[0]) handleFile(e.dataTransfer.files[0]);
        });

        fileInput.addEventListener('change', () => {
            if (fileInput.files[0]) handleFile(fileInput.files[0]);
        });

        document.getElementById('removeVideo').addEventListener('click', () => {
            fileInput.value = '';
            preview.style.display = 'none';
            area.style.display = '';
            previewVid.src = '';
        });

        titleInput.addEventListener('input', () => {
            document.getElementById('titleCount').textContent = titleInput.value.length;
        });

        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            if (!API.getToken()) { App.showLogin(); return; }
            const file = fileInput.files[0];
            if (!file) { Toast.show('请选择视频文件', 'error'); return; }
            const title = titleInput.value.trim();
            if (!title) { Toast.show('请输入标题', 'error'); return; }

            const btn = document.getElementById('publishBtn');
            const prog = document.getElementById('uploadProgress');
            btn.disabled = true;
            prog.style.display = 'block';

            try {
                await API.Video.upload(file, title, document.getElementById('videoDesc').value, (pct) => {
                    document.getElementById('progressFill').style.width = pct + '%';
                    document.getElementById('progressText').textContent = `上传中 ${pct}%`;
                });
                Toast.show('发布成功！', 'success');
                form.reset();
                preview.style.display = 'none';
                area.style.display = '';
                document.getElementById('titleCount').textContent = '0';
                Player.refresh();
                App.navigate('home');
            } catch (err) {
                Toast.show(err.message, 'error');
            } finally {
                btn.disabled = false;
                prog.style.display = 'none';
                document.getElementById('progressFill').style.width = '0';
            }
        });

        function handleFile(file) {
            if (!file.type.startsWith('video/')) { Toast.show('请选择视频文件', 'error'); return; }
            const dt = new DataTransfer();
            dt.items.add(file);
            fileInput.files = dt.files;
            previewVid.src = URL.createObjectURL(file);
            preview.style.display = 'block';
            area.style.display = 'none';
            if (!titleInput.value) {
                titleInput.value = file.name.replace(/\.[^.]+$/, '');
                document.getElementById('titleCount').textContent = titleInput.value.length;
            }
        }
    }

    /* ---------- Agent Chat ---------- */
    let chatId = null;

    function initAgent() {
        chatId = 'chat-' + Date.now();
        document.getElementById('sendBtn').addEventListener('click', sendChat);
        document.getElementById('chatInput').addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); sendChat(); }
        });
        document.getElementById('newChatBtn').addEventListener('click', () => {
            chatId = 'chat-' + Date.now();
            const msgs = document.getElementById('chatMessages');
            msgs.innerHTML = `
                <div class="chat-welcome">
                    <div class="welcome-icon"><i class="fas fa-robot"></i></div>
                    <h3>你好，我是 AI 助手</h3>
                    <p>我可以帮你搜索和推荐视频，试试问我：</p>
                    <div class="welcome-suggestions">
                        <button class="suggestion-btn">帮我找搞笑视频</button>
                        <button class="suggestion-btn">推荐热门视频</button>
                        <button class="suggestion-btn">搜索教程类视频</button>
                    </div>
                </div>
            `;
            bindSuggestions();
        });
        bindSuggestions();
    }

    function bindSuggestions() {
        document.querySelectorAll('.suggestion-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                document.getElementById('chatInput').value = btn.textContent;
                sendChat();
            });
        });
    }

    async function sendChat() {
        const input = document.getElementById('chatInput');
        const msg = input.value.trim();
        if (!msg) return;
        input.value = '';

        const msgs = document.getElementById('chatMessages');
        const welcome = msgs.querySelector('.chat-welcome');
        if (welcome) welcome.remove();

        addChatBubble(msg, 'user');

        const botBubble = addChatBubble('', 'bot');
        botBubble.innerHTML = '<div class="typing-dots"><span></span><span></span><span></span></div>';

        try {
            let reply = '';
            await API.Agent.stream(msg, chatId,
                (chunk) => {
                    reply += chunk;
                    botBubble.innerHTML = formatMarkdown(reply);
                    msgs.scrollTop = msgs.scrollHeight;
                },
                () => { msgs.scrollTop = msgs.scrollHeight; },
                (err) => { botBubble.textContent = '出错了: ' + err.message; }
            );
            if (!reply) {
                const res = await API.Agent.invoke(msg, chatId);
                reply = res.ai_reply || '暂时无法回答';
                botBubble.innerHTML = formatMarkdown(reply);
            }
        } catch (e) {
            botBubble.textContent = '出错了: ' + e.message;
        }
        msgs.scrollTop = msgs.scrollHeight;
    }

    function addChatBubble(text, role) {
        const msgs = document.getElementById('chatMessages');
        const bubble = document.createElement('div');
        bubble.className = `msg-bubble ${role}`;
        if (text) bubble.innerHTML = role === 'bot' ? formatMarkdown(text) : escapeHtml(text);
        msgs.appendChild(bubble);
        msgs.scrollTop = msgs.scrollHeight;
        return bubble;
    }

    function formatMarkdown(text) {
        return text
            .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
            .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
            .replace(/\[(.+?)\]\((.+?)\)/g, '<a href="$2" target="_blank">$1</a>')
            .replace(/!\[.*?\]\((.+?)\)/g, '<img src="$1" style="max-width:100%;border-radius:8px;margin:4px 0">')
            .replace(/\n/g, '<br>');
    }

    function escapeHtml(s) {
        return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    }

    /* ---------- Profile ---------- */
    let profileTab = 'my-videos';

    function initProfile() {
        document.getElementById('profileLoginBtn').addEventListener('click', () => App.showLogin());
        document.getElementById('logoutBtn').addEventListener('click', () => App.logout());

        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
                btn.classList.add('active');
                profileTab = btn.dataset.tab;
                loadProfileContent();
            });
        });
    }

    function updateProfile(user) {
        if (!user) {
            document.getElementById('profileGuest').style.display = '';
            document.getElementById('profileUser').style.display = 'none';
            return;
        }
        document.getElementById('profileGuest').style.display = 'none';
        document.getElementById('profileUser').style.display = '';
        document.getElementById('profileName').textContent = '@' + (user.user_name || user.username || '');
        document.getElementById('profileAvatar').src = user.avatar || '';

        const stats = document.getElementById('profileUser').querySelector('.profile-stats');
        stats.children[0].querySelector('.stat-num').textContent = user.follow_count || 0;
        stats.children[1].querySelector('.stat-num').textContent = user.follower_count || 0;
        stats.children[2].querySelector('.stat-num').textContent = user.total_favorited || 0;

        loadProfileContent();
    }

    async function loadProfileContent() {
        const grid = document.getElementById('profileGrid');
        grid.innerHTML = '<div class="empty-state"><div class="spinner"></div></div>';
        try {
            let videos = [];
            if (profileTab === 'my-videos') {
                const res = await API.Video.myList(1, 30);
                videos = res.data?.videos || [];
            } else {
                const res = await API.Favorite.myVideos(1, 30);
                videos = res.data?.videos || [];
            }
            grid.innerHTML = '';
            if (videos.length === 0) {
                grid.innerHTML = `<div class="empty-state" style="grid-column:1/-1"><i class="fas fa-video-slash"></i><p>${profileTab === 'my-videos' ? '还没有发布作品' : '还没有点赞视频'}</p></div>`;
                return;
            }
            videos.forEach(v => grid.appendChild(createGridItem(v)));
        } catch (e) {
            grid.innerHTML = `<div class="empty-state" style="grid-column:1/-1"><p>${e.message}</p></div>`;
        }
    }

    function createGridItem(video) {
        const item = document.createElement('div');
        item.className = 'grid-item';
        item.innerHTML = `
            <img src="${video.cover_url || ''}" alt="" onerror="this.style.background='var(--bg-elevated)'">
            <div class="grid-play-count"><i class="fas fa-play"></i> ${formatCount(video.view_count || 0)}</div>
        `;
        item.addEventListener('click', () => Player.playSpecificVideo(video.id));
        return item;
    }

    /* ---------- User Profile (other user) ---------- */
    async function showUserProfile(userId) {
        const page = document.getElementById('pageUserProfile');
        page.style.display = 'flex';
        page.classList.add('active');

        try {
            const res = await API.User.get(userId);
            const user = res.data;
            document.getElementById('userProfileName').textContent = '@' + (user.user_name || '');
            document.getElementById('userProfileAvatar').src = user.avatar || '';
            document.getElementById('userFollowing').textContent = user.follow_count || 0;
            document.getElementById('userFollowers').textContent = user.follower_count || 0;

            const followBtn = document.getElementById('followBtn');
            if (API.getToken()) {
                try {
                    const status = await API.Relation.followStatus(userId);
                    const isFollowing = status.data?.is_following;
                    followBtn.textContent = isFollowing ? '已关注' : '关注';
                    followBtn.className = isFollowing ? 'btn-follow following' : 'btn-follow';
                } catch {}
            }

            followBtn.onclick = async () => {
                if (!API.getToken()) { App.showLogin(); return; }
                const isFollowing = followBtn.classList.contains('following');
                try {
                    if (isFollowing) {
                        await API.Relation.unfollow(userId);
                        followBtn.textContent = '关注';
                        followBtn.className = 'btn-follow';
                    } else {
                        await API.Relation.follow(userId);
                        followBtn.textContent = '已关注';
                        followBtn.className = 'btn-follow following';
                    }
                } catch (e) { Toast.show(e.message, 'error'); }
            };
        } catch (e) {
            Toast.show(e.message, 'error');
        }

        document.getElementById('userProfileBack').onclick = () => {
            page.style.display = 'none';
            page.classList.remove('active');
        };
    }

    /* ---------- Comments ---------- */
    let currentVideoId = null;

    function initComments() {
        document.getElementById('closeComments').addEventListener('click', closeComments);
        document.querySelector('#commentDrawer .drawer-overlay').addEventListener('click', closeComments);
        document.getElementById('sendComment').addEventListener('click', submitComment);
        document.getElementById('commentInput').addEventListener('keydown', (e) => {
            if (e.key === 'Enter') { e.preventDefault(); submitComment(); }
        });
    }

    async function openComments(videoId) {
        currentVideoId = videoId;
        const drawer = document.getElementById('commentDrawer');
        drawer.classList.add('show');
        Player.pauseCurrent();

        const list = document.getElementById('commentList');
        list.innerHTML = '<div class="empty-state"><div class="spinner"></div></div>';
        try {
            const res = await API.Comment.listByVideo(videoId, 1, 50);
            const comments = res.data?.comments || [];
            document.getElementById('commentCount').textContent = `${comments.length} 条评论`;
            list.innerHTML = '';
            if (comments.length === 0) {
                list.innerHTML = '<div class="comment-empty">还没有评论，来说点什么吧</div>';
                return;
            }
            comments.forEach(c => list.appendChild(createCommentItem(c)));
        } catch (e) {
            list.innerHTML = `<div class="comment-empty">${e.message}</div>`;
        }
    }

    function closeComments() {
        document.getElementById('commentDrawer').classList.remove('show');
        Player.resumeCurrent();
        currentVideoId = null;
    }

    function createCommentItem(comment) {
        const item = document.createElement('div');
        item.className = 'comment-item';
        item.innerHTML = `
            <div class="c-avatar"><i class="fas fa-user"></i></div>
            <div class="c-body">
                <div class="c-name">@${comment.user_name || comment.user?.user_name || '用户'}</div>
                <div class="c-text">${escapeHtml(comment.content || '')}</div>
                <div class="c-time">${timeAgo(comment.created_at)}</div>
            </div>
        `;
        return item;
    }

    async function submitComment() {
        if (!API.getToken()) { App.showLogin(); return; }
        if (!currentVideoId) return;
        const input = document.getElementById('commentInput');
        const content = input.value.trim();
        if (!content) return;
        input.value = '';
        try {
            await API.Comment.create(currentVideoId, content);
            openComments(currentVideoId);
        } catch (e) {
            Toast.show(e.message, 'error');
        }
    }

    /* ---------- Helpers ---------- */
    function formatCount(n) {
        if (n >= 10000) return (n / 10000).toFixed(1) + 'w';
        if (n >= 1000) return (n / 1000).toFixed(1) + 'k';
        return String(n);
    }

    function timeAgo(dateStr) {
        if (!dateStr) return '';
        const diff = (Date.now() - new Date(dateStr).getTime()) / 1000;
        if (diff < 60) return '刚刚';
        if (diff < 3600) return Math.floor(diff / 60) + '分钟前';
        if (diff < 86400) return Math.floor(diff / 3600) + '小时前';
        if (diff < 2592000) return Math.floor(diff / 86400) + '天前';
        return new Date(dateStr).toLocaleDateString('zh-CN');
    }

    return {
        initSearch, initUpload, initAgent, initProfile, initComments,
        updateProfile, openComments, showUserProfile,
    };
})();
