const API = (() => {
    const BASE = '/api/v1';

    function getToken() { return localStorage.getItem('vida_token'); }
    function setToken(t) { localStorage.setItem('vida_token', t); }
    function removeToken() { localStorage.removeItem('vida_token'); }

    function authHeaders() {
        const t = getToken();
        const h = {};
        if (t) h['Authorization'] = `Bearer ${t}`;
        return h;
    }

    async function request(url, opts = {}) {
        const defaults = {
            headers: { 'Content-Type': 'application/json', ...authHeaders() }
        };
        if (opts.headers) {
            opts.headers = { ...defaults.headers, ...opts.headers };
        }
        const res = await fetch(url, { ...defaults, ...opts });
        const text = await res.text();
        let data;
        try { data = JSON.parse(text); } catch {
            throw new Error(text || `请求失败: ${res.status}`);
        }
        if (!res.ok) {
            throw new Error(data.error?.message || data.detail || data.message || `请求失败: ${res.status}`);
        }
        return data;
    }

    const Auth = {
        login: (username, password) =>
            request(`${BASE}/auth/login`, { method: 'POST', body: JSON.stringify({ username, password }) }),
        register: (username, password) =>
            request(`${BASE}/auth/register`, { method: 'POST', body: JSON.stringify({ username, password }) }),
        logout: () =>
            request(`${BASE}/auth/logout`, { method: 'POST' }),
        me: () =>
            request(`${BASE}/auth/me`),
    };

    const Video = {
        feed: (page = 1, pageSize = 10) =>
            request(`${BASE}/videos/feed?page=${page}&page_size=${pageSize}`),
        detail: (id) =>
            request(`${BASE}/videos/${id}`),
        myList: (page = 1, pageSize = 20) =>
            request(`${BASE}/videos/my/list?page=${page}&page_size=${pageSize}`),
        upload: (file, title, description, onProgress) => {
            return new Promise((resolve, reject) => {
                const fd = new FormData();
                fd.append('video_file', file);
                fd.append('title', title);
                fd.append('description', description || '');
                const xhr = new XMLHttpRequest();
                xhr.open('POST', `${BASE}/videos/upload`);
                const t = getToken();
                if (t) xhr.setRequestHeader('Authorization', `Bearer ${t}`);
                xhr.upload.onprogress = (e) => {
                    if (e.lengthComputable && onProgress) {
                        onProgress(Math.round(e.loaded / e.total * 100));
                    }
                };
                xhr.onload = () => {
                    try {
                        const data = JSON.parse(xhr.responseText);
                        if (xhr.status >= 200 && xhr.status < 300) resolve(data);
                        else reject(new Error(data.message || '上传失败'));
                    } catch { reject(new Error('上传失败')); }
                };
                xhr.onerror = () => reject(new Error('网络错误'));
                xhr.send(fd);
            });
        },
        update: (id, data) =>
            request(`${BASE}/videos/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
        delete: (id) =>
            request(`${BASE}/videos/${id}`, { method: 'DELETE' }),
    };

    const Favorite = {
        like: (videoId) =>
            request(`${BASE}/favorites/${videoId}`, { method: 'POST' }),
        unlike: (videoId) =>
            request(`${BASE}/favorites/${videoId}`, { method: 'DELETE' }),
        status: (videoId) =>
            request(`${BASE}/favorites/${videoId}/status`),
        myList: (page = 1, pageSize = 20) =>
            request(`${BASE}/favorites/my/list?page=${page}&page_size=${pageSize}`),
        myVideos: (page = 1, pageSize = 20) =>
            request(`${BASE}/favorites/my/videos?page=${page}&page_size=${pageSize}`),
        batchStatus: (ids) =>
            request(`${BASE}/favorites/batch/status`, { method: 'POST', body: JSON.stringify({ video_ids: ids }) }),
    };

    const Comment = {
        create: (videoId, content, parentId = null) =>
            request(`${BASE}/comments/${videoId}`, {
                method: 'POST',
                body: JSON.stringify({ content, parent_id: parentId })
            }),
        update: (id, content) =>
            request(`${BASE}/comments/${id}`, { method: 'PUT', body: JSON.stringify({ content }) }),
        delete: (id) =>
            request(`${BASE}/comments/${id}`, { method: 'DELETE' }),
        listByVideo: (videoId, page = 1, pageSize = 20) =>
            request(`${BASE}/comments/video/${videoId}?page=${page}&page_size=${pageSize}`),
        replies: (id, page = 1, pageSize = 20) =>
            request(`${BASE}/comments/${id}/replies?page=${page}&page_size=${pageSize}`),
    };

    const Relation = {
        follow: (id) =>
            request(`${BASE}/relations/follow/${id}`, { method: 'POST' }),
        unfollow: (id) =>
            request(`${BASE}/relations/unfollow/${id}`, { method: 'POST' }),
        following: (id, page = 1) =>
            request(`${BASE}/relations/following/${id}?page=${page}`),
        followers: (id, page = 1) =>
            request(`${BASE}/relations/followers/${id}?page=${page}`),
        followStatus: (id) =>
            request(`${BASE}/relations/following/${id}/status`),
    };

    const Search = {
        videos: (q, sort = 'relevance', page = 1, pageSize = 20) => {
            const p = new URLSearchParams({ q, sort, page, page_size: pageSize });
            return request(`${BASE}/search/videos?${p}`);
        },
    };

    const Agent = {
        invoke: (message, chatId) =>
            request(`${BASE}/agent/invoke`, {
                method: 'POST',
                body: JSON.stringify({ message, chat_id: chatId })
            }),

        stream: async (message, chatId, onChunk, onDone, onError) => {
            try {
                const res = await fetch(`${BASE}/agent/stream`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json', ...authHeaders() },
                    body: JSON.stringify({ message, chat_id: chatId })
                });
                if (!res.ok) throw new Error('请求失败');
                const reader = res.body.getReader();
                const decoder = new TextDecoder();
                let buf = '';
                while (true) {
                    const { done, value } = await reader.read();
                    if (done) break;
                    buf += decoder.decode(value, { stream: true });
                    const lines = buf.split('\n');
                    buf = lines.pop() || '';
                    for (const line of lines) {
                        if (!line.startsWith('data: ')) continue;
                        try {
                            const d = JSON.parse(line.slice(6));
                            if (d.code === 200 && d.message === 'streaming' && d.data?.content) {
                                onChunk(d.data.content);
                            } else if (d.code === 200 && d.message === 'done') {
                                if (onDone) onDone(d.data?.chat_id);
                            } else if (d.code !== 200) {
                                if (onError) onError(new Error(d.message));
                            }
                        } catch {}
                    }
                }
            } catch (e) {
                if (onError) onError(e);
            }
        },

        chats: () => request(`${BASE}/agent/chats`),
        chatMessages: (chatId) =>
            request(`${BASE}/agent/chats/${chatId}`),
        deleteChat: (chatId) =>
            request(`${BASE}/agent/chats/${chatId}`, { method: 'DELETE' }),
    };

    const User = {
        get: (id) => request(`${BASE}/users/${id}`),
        getProfile: (id) => request(`${BASE}/users/${id}/profile`),
        getMe: () => request(`${BASE}/users/me`),
        uploadAvatar: (file) => {
            const fd = new FormData();
            fd.append('avatar', file);
            return fetch(`${BASE}/users/me/avatar`, {
                method: 'POST',
                headers: authHeaders(),
                body: fd,
            }).then(async (res) => {
                const data = await res.json();
                if (!res.ok) throw new Error(data.error?.message || data.message || '上传失败');
                return data;
            });
        },
    };

    return {
        Auth, Video, Favorite, Comment, Relation, Search, Agent, User,
        getToken, setToken, removeToken,
    };
})();
