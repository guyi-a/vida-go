const App = (() => {
    let currentUser = null;
    let currentPage = 'home';

    function init() {
        initNav();
        initModals();

        UI.initSearch();
        UI.initUpload();
        UI.initAgent();
        UI.initProfile();
        UI.initComments();
        Player.init();

        checkAuth();
    }

    /* ---------- Navigation ---------- */
    function initNav() {
        document.querySelectorAll('.nav-item').forEach(btn => {
            btn.addEventListener('click', () => {
                const page = btn.dataset.page;
                if (page === 'upload' && !API.getToken()) {
                    showLogin();
                    return;
                }
                navigate(page);
            });
        });
    }

    function navigate(page) {
        if (page === currentPage) {
            if (page === 'home') Player.refresh();
            return;
        }

        if (currentPage === 'home') Player.pauseCurrent();

        document.querySelectorAll('.page').forEach(p => {
            p.classList.remove('active');
            if (p.id !== 'pageUserProfile') p.style.display = '';
        });
        document.getElementById('pageUserProfile').style.display = 'none';

        document.querySelectorAll('.nav-item').forEach(b => b.classList.remove('active'));
        const navBtn = document.querySelector(`.nav-item[data-page="${page}"]`);
        if (navBtn) navBtn.classList.add('active');

        const pageEl = document.getElementById('page' + capitalize(page));
        if (pageEl) {
            pageEl.classList.add('active');
        }

        if (page === 'home') Player.resumeCurrent();
        if (page === 'profile') UI.updateProfile(currentUser);

        currentPage = page;
    }

    function capitalize(s) {
        return s.charAt(0).toUpperCase() + s.slice(1);
    }

    /* ---------- Auth ---------- */
    async function checkAuth() {
        if (!API.getToken()) return;
        try {
            const res = await API.Auth.me();
            currentUser = res.data;
            UI.updateProfile(currentUser);
        } catch {
            API.removeToken();
            currentUser = null;
        }
    }

    function showLogin() {
        document.getElementById('loginModal').classList.add('show');
    }

    function hideLogin() {
        document.getElementById('loginModal').classList.remove('show');
    }

    function showRegister() {
        hideLogin();
        document.getElementById('registerModal').classList.add('show');
    }

    function hideRegister() {
        document.getElementById('registerModal').classList.remove('show');
    }

    function initModals() {
        // Login
        document.getElementById('loginForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            const user = document.getElementById('loginUsername').value.trim();
            const pass = document.getElementById('loginPassword').value;
            try {
                const res = await API.Auth.login(user, pass);
                API.setToken(res.data.token);
                currentUser = res.data.user;
                hideLogin();
                Toast.show('登录成功', 'success');
                UI.updateProfile(currentUser);
                Player.refresh();
            } catch (err) {
                Toast.show(err.message, 'error');
            }
        });

        document.getElementById('loginClose').addEventListener('click', hideLogin);
        document.querySelector('#loginModal .modal-overlay').addEventListener('click', hideLogin);
        document.getElementById('toRegister').addEventListener('click', (e) => { e.preventDefault(); showRegister(); });

        // Register
        document.getElementById('registerForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            const user = document.getElementById('regUsername').value.trim();
            const pass = document.getElementById('regPassword').value;
            const confirm = document.getElementById('regConfirm').value;
            if (pass !== confirm) { Toast.show('两次密码不一致', 'error'); return; }
            try {
                await API.Auth.register(user, pass);
                Toast.show('注册成功，请登录', 'success');
                hideRegister();
                showLogin();
            } catch (err) {
                Toast.show(err.message, 'error');
            }
        });

        document.getElementById('registerClose').addEventListener('click', hideRegister);
        document.querySelector('#registerModal .modal-overlay').addEventListener('click', hideRegister);
        document.getElementById('toLogin').addEventListener('click', (e) => { e.preventDefault(); hideRegister(); showLogin(); });
    }

    function logout() {
        API.Auth.logout().catch(() => {});
        API.removeToken();
        currentUser = null;
        UI.updateProfile(null);
        Toast.show('已退出登录');
        navigate('home');
    }

    function showUserProfile(userId) {
        UI.showUserProfile(userId);
    }

    function getUser() { return currentUser; }

    return { init, navigate, showLogin, showUserProfile, logout, getUser };
})();

document.addEventListener('DOMContentLoaded', () => App.init());
