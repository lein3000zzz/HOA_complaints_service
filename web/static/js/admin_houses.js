"use strict";

document.addEventListener("DOMContentLoaded", () => {
    let page = 1;
    let limit = 10;
    let lastPages = 1;

    const list = document.getElementById("houses-list");
    const out = document.getElementById("houses-output");
    const totalCountEl = document.getElementById("total-count");
    const currentPageEl = document.getElementById("current-page");
    const totalPagesEl = document.getElementById("total-pages");

    const prevBtn = document.getElementById("prev-page");
    const nextBtn = document.getElementById("next-page");
    const refreshBtn = document.getElementById("refresh-btn");
    const limitSelect = document.getElementById("limit-select");

    const patternInput = document.getElementById("pattern-input");
    const applyBtn = document.getElementById("apply-filters");

    const addHouseBtn = document.getElementById("add-house");
    const addHouseModal = document.getElementById("add-house-modal");
    const addHouseForm = document.getElementById("add-house-form");
    const addHouseOutput = document.getElementById("add-house-output");
    const addHouseCancel = document.getElementById("add-house-cancel");

    const updateHouseModal = document.getElementById("update-house-modal");
    const updateHouseForm = document.getElementById("update-house-form");
    const updateHouseId = document.getElementById("update-house-id");
    const updateHouseAddress = document.getElementById("update-house-address");
    const updateHouseCancel = document.getElementById("update-house-cancel");
    const updateHouseOutput = document.getElementById("update-house-output");

    const toggleAddModal = (show) => {
        if (!addHouseModal) return;
        if (show) addHouseModal.classList.remove("hidden");
        else addHouseModal.classList.add("hidden");
        window.scrollTo(0, 0);
    };

    const toggleUpdateModal = (show) => {
        if (!updateHouseModal) return;
        if (show) updateHouseModal.classList.remove("hidden");
        else updateHouseModal.classList.add("hidden");
        window.scrollTo(0, 0);
    };

    if (addHouseBtn) addHouseBtn.addEventListener("click", () => toggleAddModal(true));
    if (addHouseCancel) addHouseCancel.addEventListener("click", (e) => { e.preventDefault(); toggleAddModal(false); });

    if (addHouseForm) {
        addHouseForm.addEventListener("submit", async (e) => {
            e.preventDefault();
            if (addHouseOutput) { addHouseOutput.textContent = "Saving..."; addHouseOutput.className = "form-output"; }

            const formData = new FormData(addHouseForm);
            try {
                const res = await fetch(addHouseForm.dataset.endpoint || '/api/staff/houses/create', {
                    method: 'POST',
                    body: formData,
                    credentials: 'same-origin'
                });
                const text = await res.text();
                let data;
                try { data = JSON.parse(text || '{}'); } catch { data = { raw: text }; }

                if (!res.ok) {
                    if (addHouseOutput) { addHouseOutput.textContent = data.error || data.raw || ('HTTP ' + res.status); addHouseOutput.className = 'form-output error'; }
                    return;
                }

                if (addHouseOutput) { addHouseOutput.textContent = 'Created'; addHouseOutput.className = 'form-output success'; }
                if (addHouseForm) addHouseForm.reset()
                toggleAddModal(false);
                if (typeof load === 'function') load();
            } catch {
                if (addHouseOutput) { addHouseOutput.textContent = 'Network error'; addHouseOutput.className = 'form-output error'; }
            }
        });
    }

    if (updateHouseCancel) {
        updateHouseCancel.addEventListener("click", (e) => { e.preventDefault(); toggleUpdateModal(false); });
    }

    if (updateHouseForm) {
        updateHouseForm.addEventListener("submit", async (e) => {
            e.preventDefault();
            if (updateHouseOutput) { updateHouseOutput.textContent = "Saving..."; updateHouseOutput.className = "form-output"; }

            const formData = new FormData(updateHouseForm);
            try {
                const res = await fetch(updateHouseForm.dataset.endpoint || '/api/staff/users/resident/update-house', {
                    method: 'POST',
                    body: formData,
                    credentials: 'same-origin'
                });
                const text = await res.text();
                let data;
                try { data = JSON.parse(text || '{}'); } catch { data = { raw: text }; }

                if (!res.ok) {
                    if (updateHouseOutput) { updateHouseOutput.textContent = data.error || data.raw || ('HTTP ' + res.status); updateHouseOutput.className = 'form-output error'; }
                    return;
                }

                if (updateHouseOutput) { updateHouseOutput.textContent = 'Updated'; updateHouseOutput.className = 'form-output success'; }
                toggleUpdateModal(false);
                if (typeof load === 'function') load();
            } catch {
                if (updateHouseOutput) { updateHouseOutput.textContent = 'Network error'; updateHouseOutput.className = 'form-output error'; }
            }
        });
    }

    const buildUrl = () => {
        const url = new URL('/api/staff/houses/list', window.location.origin);
        url.searchParams.set('page', String(page));
        url.searchParams.set('limit', String(limit));
        const pattern = patternInput && patternInput.value.trim();
        if (pattern) url.searchParams.set('pattern', pattern);
        return url.toString();
    };

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

    const copyToClipboard = async (text) => {
        if (!text && text !== "") return Promise.reject(new Error("empty text"));
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

    const render = (data) => {
        clear();
        const houses = data.houses || [];

        const meta = data.meta || {};
        if (typeof meta.page === 'number') page = meta.page;
        if (typeof meta.pages === 'number') lastPages = meta.pages;
        else if (typeof data.total === 'number') {
            lastPages = Math.max(1, Math.ceil(data.total / limit));
        }

        if (totalCountEl) {
            const total = (typeof meta.total === 'number') ? meta.total : (data.total || 0);
            totalCountEl.textContent = String(total);
        }

        updateControls();

        if (!houses.length) {
            if (out) out.textContent = "No houses found";
            return;
        }

        houses.forEach(h => {
            const idText = (h.ID !== undefined && h.ID !== null) ? String(h.ID) : '';
            const addressText = h.Address || '';

            const card = document.createElement("div");
            card.className = "card";
            card.style.margin = "8px 0";

            const info = document.createElement("div");
            info.style.fontWeight = "700";

            const idRow = document.createElement("div");
            const idLabel = document.createElement("span");
            idLabel.style.color = "var(--muted)";
            idLabel.textContent = "ID: ";
            const idValue = document.createElement("span");
            idValue.className = "house-id";
            idValue.textContent = idText;

            idRow.appendChild(idLabel);
            idRow.appendChild(idValue);

            const copyBtn = document.createElement("button");
            copyBtn.className = "btn";
            copyBtn.style.marginLeft = "8px";
            copyBtn.style.padding = "2px 8px";
            copyBtn.style.fontSize = "12px";
            copyBtn.type = "button";
            copyBtn.title = "Copy ID";
            copyBtn.textContent = "Copy";

            copyBtn.addEventListener("click", async () => {
                const prevText = copyBtn.textContent;
                copyBtn.disabled = true;
                copyBtn.textContent = "Copying...";
                try {
                    await copyToClipboard(idText);
                    copyBtn.textContent = "Copied";
                    if (out) { out.textContent = "ID copied to clipboard"; out.className = "form-output success"; }
                } catch {
                    copyBtn.textContent = "Failed";
                    if (out) { out.textContent = "Copy failed"; out.className = "form-output error"; }
                } finally {
                    setTimeout(() => {
                        copyBtn.disabled = false;
                        copyBtn.textContent = prevText;
                        if (out) { out.textContent = ""; out.className = "form-output"; }
                    }, 1200);
                }
            });

            idRow.appendChild(copyBtn);

            const addrRow = document.createElement("div");
            const addrLabel = document.createElement("span");
            addrLabel.style.color = "var(--muted)";
            addrLabel.textContent = "Address: ";
            const addrValue = document.createElement("span");
            addrValue.textContent = addressText;

            addrRow.appendChild(addrLabel);
            addrRow.appendChild(addrValue);

            info.appendChild(idRow);
            info.appendChild(document.createElement("br"));
            info.appendChild(addrRow);

            const actions = document.createElement("div");
            actions.className = "form-row";
            actions.style.marginTop = "8px";
            const editBtn = document.createElement("button");
            editBtn.className = "btn";
            editBtn.type = "button";
            editBtn.textContent = "Edit address";
            editBtn.addEventListener("click", () => {
                if (updateHouseId) updateHouseId.value = idText;
                if (updateHouseAddress) updateHouseAddress.value = addressText;
                if (updateHouseOutput) { updateHouseOutput.textContent = ""; updateHouseOutput.className = "form-output"; }
                toggleUpdateModal(true);
            });

            actions.appendChild(editBtn);

            card.appendChild(info);
            card.appendChild(actions);
            list.appendChild(card);
        });
    };

    const load = () => {
        clear();
        if (out) { out.textContent = "Loading..."; out.className = "form-output"; }
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
            .then(render)
            .catch(err => {
                clear();
                if (out) {
                    out.className = "form-output error";
                    out.textContent = err && err.error ? err.error : (err && err.message ? err.message : String(err));
                }
                updateControls();
            });
    };

    if (prevBtn) prevBtn.addEventListener('click', () => {
        if (page > 1) { page--; load(); }
    });
    if (nextBtn) nextBtn.addEventListener('click', () => {
        if (page < lastPages) { page++; load(); }
    });

    if (applyBtn) applyBtn.addEventListener('click', () => { page = 1; load(); });
    if (refreshBtn) refreshBtn.addEventListener('click', () => load());
    if (limitSelect) limitSelect.addEventListener('change', (e) => {
        const v = parseInt(e.target.value || '10', 10);
        if (!isNaN(v) && v > 0) { limit = v; page = 1; load(); }
    });

    load();
});