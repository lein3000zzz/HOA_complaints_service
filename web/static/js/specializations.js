"use strict";

document.addEventListener("DOMContentLoaded", () => {
    let page = 1;
    let limit = 10;
    let lastPages = 1;

    const list = document.getElementById("specs-list");
    const out = document.getElementById("specs-output");
    const totalCountEl = document.getElementById("total-count");
    const currentPageEl = document.getElementById("current-page");
    const totalPagesEl = document.getElementById("total-pages");

    const prevBtn = document.getElementById("prev-page");
    const nextBtn = document.getElementById("next-page");
    const refreshBtn = document.getElementById("refresh-btn");
    const limitSelect = document.getElementById("limit-select");

    const filterPattern = document.getElementById("filter-pattern");
    const applyBtn = document.getElementById("apply-filters");

    const addSpecBtn = document.getElementById("add-spec-btn");
    const createModal = document.getElementById("create-modal");
    const createForm = document.getElementById("create-form");
    const createOutput = document.getElementById("create-output");
    const createCancel = document.getElementById("create-cancel");

    const clear = () => {
        if (list) list.innerHTML = "";
        if (out) { out.textContent = ""; out.className = "form-output"; }
    };

    const updateControls = () => {
        if (currentPageEl) currentPageEl.textContent = String(page);
        if (totalPagesEl) totalPagesEl.textContent = String(lastPages);
        if (prevBtn) prevBtn.disabled = page <= 1;
        if (nextBtn) nextBtn.disabled = page >= lastPages;
    };

    const buildUrl = () => {
        const url = new URL('/api/staff/specializations/list', window.location.origin);
        url.searchParams.set('page', String(page));
        url.searchParams.set('limit', String(limit));
        const pattern = filterPattern && filterPattern.value ? filterPattern.value.trim() : "";
        if (pattern) url.searchParams.set('pattern', pattern);
        return url.toString();
    };

    const copyToClipboard = async (text) => {
        if (text === undefined || text === null) return Promise.reject(new Error("empty text"));
        if (navigator.clipboard && navigator.clipboard.writeText) {
            return navigator.clipboard.writeText(text);
        }
        return new Promise((resolve, reject) => {
            try {
                const ta = document.createElement("textarea");
                ta.value = text;
                ta.style.position = "fixed";
                ta.style.left = "-9999px";
                document.body.appendChild(ta);
                ta.select();
                const ok = document.execCommand("copy");
                document.body.removeChild(ta);
                if (ok) resolve();
                else reject(new Error("execCommand failed"));
            } catch (err) {
                reject(err);
            }
        });
    };

    const renderSpecs = (data) => {
        clear();
        const specs = data.specializations || [];
        const meta = data.meta || {};
        const total = (typeof meta.total === 'number') ? meta.total : 0;
        const pageFromMeta = (meta.page && typeof meta.page === 'number') ? meta.page : page;
        const pages = (meta.pages && typeof meta.pages === 'number') ? meta.pages : Math.max(1, Math.ceil(total / limit));

        lastPages = pages;
        page = pageFromMeta;

        if (totalCountEl) totalCountEl.textContent = String(total);
        if (currentPageEl) currentPageEl.textContent = String(page);
        if (totalPagesEl) totalPagesEl.textContent = String(pages);

        updateControls();

        if (!specs.length) {
            if (out) out.textContent = 'No specializations found';
            return;
        }

        specs.forEach(s => {
            const idText = (s.ID !== undefined && s.ID !== null) ? String(s.ID) : '';
            const titleText = s.Title || s.title || s.Name || s.name || '';

            const card = document.createElement('div');
            card.className = 'card';
            card.style.margin = '8px 0';

            const info = document.createElement('div');
            info.style.fontWeight = '700';
            info.style.marginBottom = '6px';

            const idRow = document.createElement('div');
            const idLabel = document.createElement('span');
            idLabel.style.color = 'var(--muted)';
            idLabel.textContent = 'ID: ';
            const idValue = document.createElement('span');
            idValue.textContent = idText;

            idRow.appendChild(idLabel);
            idRow.appendChild(idValue);

            const copyBtn = document.createElement('button');
            copyBtn.className = 'btn';
            copyBtn.type = 'button';
            copyBtn.style.marginLeft = '8px';
            copyBtn.style.padding = '2px 8px';
            copyBtn.style.fontSize = '12px';
            copyBtn.title = 'Copy ID';
            copyBtn.textContent = 'Copy';

            if (!idText) {
                copyBtn.disabled = true;
            } else {
                copyBtn.addEventListener('click', async () => {
                    const prevText = copyBtn.textContent;
                    copyBtn.disabled = true;
                    copyBtn.textContent = 'Copying...';
                    try {
                        await copyToClipboard(idText);
                        copyBtn.textContent = 'Copied';
                        if (out) { out.textContent = 'ID copied to clipboard'; out.className = 'form-output success'; }
                    } catch (err) {
                        copyBtn.textContent = 'Failed';
                        if (out) { out.textContent = 'Copy failed'; out.className = 'form-output error'; }
                    } finally {
                        setTimeout(() => {
                            copyBtn.disabled = false;
                            copyBtn.textContent = prevText;
                            if (out) { out.textContent = ''; out.className = 'form-output'; }
                        }, 1200);
                    }
                });
            }

            idRow.appendChild(copyBtn);

            const nameRow = document.createElement('div');
            const nameLabel = document.createElement('span');
            nameLabel.style.color = 'var(--muted)';
            nameLabel.textContent = 'Name: ';
            const nameValue = document.createElement('span');
            nameValue.textContent = titleText;

            nameRow.appendChild(nameLabel);
            nameRow.appendChild(document.createTextNode(' '));
            nameRow.appendChild(nameValue);

            info.appendChild(idRow);
            info.appendChild(nameRow);

            card.appendChild(info);
            list.appendChild(card);
        });
    };

    const load = () => {
        clear();
        if (out) { out.textContent = 'Loading...'; out.className = 'form-output'; }

        fetch(buildUrl(), { credentials: 'same-origin' })
            .then(async res => {
                const text = await res.text();
                try {
                    const json = JSON.parse(text || '{}');
                    if (!res.ok) return Promise.reject(json);
                    return json;
                } catch {
                    const raw = { raw: text };
                    if (!res.ok) return Promise.reject(raw);
                    return raw;
                }
            })
            .then(renderSpecs)
            .catch(err => {
                clear();
                if (out) {
                    out.className = 'form-output error';
                    out.textContent = err && err.error ? err.error : (err && err.message ? err.message : String(err));
                }
                updateControls();
            });
    };

    if (addSpecBtn) {
        addSpecBtn.addEventListener('click', () => {
            if (!createModal) return;
            createModal.classList.toggle('hidden');
            if (!createModal.classList.contains('hidden')) {
                document.getElementById("job-name").focus();
            }
        });
    }

    if (createCancel) {
        createCancel.addEventListener('click', () => {
            if (!createModal) return;
            createModal.classList.add('hidden');
        });
    }

    if (createForm) {
        createForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            if (createOutput) { createOutput.textContent = 'Creating...'; createOutput.className = 'form-output'; }
            const endpoint = createForm.dataset.endpoint || '/api/staff/specializations/create';
            const body = new FormData(createForm);
            try {
                const res = await fetch(endpoint, {
                    method: 'POST',
                    body,
                    credentials: 'same-origin'
                });
                const text = await res.text();
                let data;
                try { data = JSON.parse(text || '{}'); } catch { data = { raw: text }; }
                if (!res.ok) {
                    if (createOutput) { createOutput.textContent = data.error || data.raw || ('HTTP ' + res.status); createOutput.className = 'form-output error'; }
                    return;
                }
                if (createOutput) { createOutput.textContent = 'Created'; createOutput.className = 'form-output success'; }
                createForm.reset();
                createModal.classList.add('hidden');
                load();
            } catch (err) {
                if (createOutput) { createOutput.textContent = 'Network error'; createOutput.className = 'form-output error'; }
            }
        });
    }

    if (prevBtn) prevBtn.addEventListener('click', () => {
        if (page > 1) { page--; load(); }
    });
    if (nextBtn) nextBtn.addEventListener('click', () => {
        if (page < lastPages) { page++; load(); }
    });
    if (refreshBtn) refreshBtn.addEventListener('click', () => load());
    if (limitSelect) limitSelect.addEventListener('change', (e) => {
        const v = parseInt(e.target.value || '10', 10);
        if (!isNaN(v) && v > 0) { limit = v; page = 1; load(); }
    });
    if (applyBtn) applyBtn.addEventListener('click', () => { page = 1; load(); });

    load();
});