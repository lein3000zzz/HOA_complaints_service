"use strict"

document.addEventListener('DOMContentLoaded', function () {
    const list = document.getElementById('requests-list');
    const out = document.getElementById('requests-output');
    const totalCountEl = document.getElementById('total-count');
    const currentPageEl = document.getElementById('current-page');
    const totalPagesEl = document.getElementById('total-pages');

    const prevBtn = document.getElementById('prev-page');
    const nextBtn = document.getElementById('next-page');
    const refreshBtn = document.getElementById('refresh-btn');
    const sortSelect = document.getElementById('sort-select');
    const limitSelect = document.getElementById('limit-select');

    let page = 1;
    let limit = limitSelect ? parseInt(limitSelect.value || '10', 10) : 10;
    let sort = sortSelect ? sortSelect.value : '';
    let lastPages = 1;

    const clear = () => {
        if (list) list.innerHTML = '';
        if (out) {
            out.className = 'form-output';
            out.textContent = '';
        }
    };

    const updateControls = () => {
        if (currentPageEl) currentPageEl.textContent = String(page);
        if (totalPagesEl) totalPagesEl.textContent = String(lastPages);

        if (prevBtn) prevBtn.disabled = page <= 1;
        if (nextBtn) nextBtn.disabled = page >= lastPages;
    };

    const renderRequests = (data) => {
        clear();

        const requests = Array.isArray(data.requests) ? data.requests : [];
        const total = (data.meta && typeof data.meta.total === 'number') ? data.meta.total : (data.totalRequests || 0);
        const pageFromMeta = (data.meta && typeof data.meta.page === 'number') ? data.meta.page : page;
        const pages = (data.meta && typeof data.meta.pages === 'number') ? data.meta.pages : (Math.max(1, Math.ceil(total / limit)));

        lastPages = pages;
        page = pageFromMeta;

        if (totalCountEl) totalCountEl.textContent = String(total);
        if (currentPageEl) currentPageEl.textContent = String(page);
        if (totalPagesEl) totalPagesEl.textContent = String(pages);

        if (!requests.length) {
            if (out) out.textContent = 'No requests found';
            updateControls();
            return;
        }

        requests.forEach(r => {
            const card = document.createElement('div');
            card.className = 'card';
            card.style.margin = '8px 0';

            const id = r.ID || '';
            const type = r.RequestType || '';
            const status = r.Status || '';
            const complaint = r.Complaint || '';
            const created = r.CreatedAt || '';

            const createdStr = created ? (new Date(created)).toLocaleString() : '';

            card.innerHTML = '<div style="font-weight:700;margin-bottom:6px;">' +
                '<span style="color:var(--muted);">ID: </span> ' + id + '<br><span style="color:var(--muted);">Тип: </span>' + type + '<br><span style="color:var(--muted);">Статус: </span>' + status +
                '</div>' +
                '<div style="margin-bottom:8px;">' + (complaint || '') + '</div>' +
                '<div style="font-size:12px;color:var(--muted);">' + createdStr + '</div>';

            list.appendChild(card);
        });

        updateControls();
    };

    const buildUrl = () => {
        const url = new URL('/api/resident/requests', window.location.origin);
        url.searchParams.set('page', String(page));
        url.searchParams.set('limit', String(limit));
        if (sort) url.searchParams.set('sort', sort);
        return url.toString();
    };

    const load = () => {
        clear();
        if (out) {
            out.textContent = 'Loading...';
            out.className = 'form-output';
        }

        fetch(buildUrl(), { credentials: 'same-origin' })
            .then(res => {
                return res.text().then(text => {
                    try { return JSON.parse(text || '{}'); } catch { return { raw: text }; }
                }).then(json => {
                    if (!res.ok) return Promise.reject(json);
                    return json;
                });
            })
            .then(renderRequests)
            .catch(err => {
                clear();
                if (out) {
                    out.className = 'form-output error';
                    out.textContent = err && err.error ? err.error : (err && err.message ? err.message : String(err));
                }

                updateControls();
            });
    };

    if (prevBtn) prevBtn.addEventListener('click', () => {
        if (page > 1) {
            page--;
            load();
        }
    });

    if (nextBtn) nextBtn.addEventListener('click', () => {
        if (page < lastPages) {
            page++;
            load();
        }
    });

    if (refreshBtn) refreshBtn.addEventListener('click', () => {
        load();
    });

    if (sortSelect) sortSelect.addEventListener('change', (e) => {
        sort = e.target.value || '';
        page = 1;
        load();
    });

    if (limitSelect) limitSelect.addEventListener('change', (e) => {
        const v = parseInt(e.target.value || '10', 10);
        if (!isNaN(v) && v > 0) {
            limit = v;
            page = 1;
            load();
        }
    });

    load();
});