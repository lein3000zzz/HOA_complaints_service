"use strict";

document.addEventListener("DOMContentLoaded", () => {
    let page = 1;
    let limit = 20;
    let lastPages = 1;
    const listEl = document.getElementById("users-list");
    const out = document.getElementById("users-output");
    const currentPageEl = document.getElementById("current-page");
    const totalPagesEl = document.getElementById("total-pages");
    const totalCountEl = document.getElementById("total-count");

    const prevBtn = document.getElementById("prev-page");
    const nextBtn = document.getElementById("next-page");
    const searchBtn = document.getElementById("search-btn");
    const searchInput = document.getElementById("search-input");
    const limitSelect = document.getElementById("limit-select");

    const modal = document.getElementById("user-modal");
    const modalOutput = document.getElementById("modal-output");
    const modalClose = document.getElementById("modal-close");

    const tabResidentBtn = document.getElementById("tab-resident");
    const tabStaffBtn = document.getElementById("tab-staff");
    const residentPanel = document.getElementById("resident-panel");
    const staffPanel = document.getElementById("staff-panel");

    const resIdEl = document.getElementById("res-id");
    const resPhoneEl = document.getElementById("res-phone");
    const resFullEl = document.getElementById("res-fullname");
    const housesList = document.getElementById("houses-list");
    const btnGetHouses = document.getElementById("btn-get-houses");

    const staffIdEl = document.getElementById("staff-id");
    const staffPhoneEl = document.getElementById("staff-phone");
    const staffFullEl = document.getElementById("staff-fullname");
    const staffStatusEl = document.getElementById("staff-status");
    const specsList = document.getElementById("specs-list");
    const btnGetSpecs = document.getElementById("btn-get-specs");

    const clearList = () => { if (listEl) listEl.innerHTML = ""; };
    const updateControls = () => {
        if (currentPageEl) currentPageEl.textContent = String(page);
        if (totalPagesEl) totalPagesEl.textContent = String(lastPages);
    };

    const buildUrl = () => {
        const url = new URL('/api/staff/users/list', window.location.origin);
        url.searchParams.set('page', String(page));
        url.searchParams.set('limit', String(limit));
        const q = searchInput && searchInput.value ? searchInput.value.trim() : '';
        if (q) url.searchParams.set('phoneNumber', q);
        return url.toString();
    };

    const renderUsers = (data) => {
        clearList();
        const phones = data.phones || [];
        const meta = data.meta || {};
        const total = Number(meta.total) || 0;
        const pageFromMeta = (meta.page && typeof meta.page === 'number') ? meta.page : page;
        const pages = (meta.pages && typeof meta.pages === 'number') ? meta.pages : Math.max(1, Math.ceil(total / limit));
        lastPages = pages;
        page = pageFromMeta;

        if (totalCountEl) totalCountEl.textContent = String(total);
        if (out) { out.className = 'form-output'; out.textContent = ''; }
        if (!phones.length) {
            if (out) out.textContent = 'No users found';
            updateControls();
            return;
        }

        phones.forEach(u => {
            const phone = u;
            const card = document.createElement('div');
            card.className = 'card';
            card.style.margin = '8px 0';

            const title = document.createElement('div');
            title.style.fontWeight = '700';
            title.innerHTML = '<span style="color:var(--muted);">Phone: </span>' + (phone || '—');
            card.appendChild(title);

            const actions = document.createElement('div');
            actions.style.marginTop = '8px';
            actions.style.display = 'flex';
            actions.style.gap = '8px';

            const detailsBtn = document.createElement('button');
            detailsBtn.className = 'btn';
            detailsBtn.textContent = 'Details';
            detailsBtn.addEventListener('click', () => openDetails(phone));

            const delBtn = document.createElement('button');
            delBtn.className = 'btn';
            delBtn.textContent = 'Delete';
            delBtn.addEventListener('click', async () => {
                if (!confirm('Delete user ' + phone + '?')) return;
                try {
                    const res = await fetch('/api/staff/users/delete/' + encodeURIComponent(phone), { method: 'GET', credentials: 'same-origin' });
                    const text = await res.text();
                    let json;
                    try { json = JSON.parse(text || '{}'); } catch { json = { raw: text }; }
                    if (!res.ok) {
                        alert(json.error || json.message || ('Delete failed: ' + res.status));
                    } else {
                        load();
                    }
                } catch (err) {
                    alert('Network error');
                }
            });

            actions.appendChild(detailsBtn);
            actions.appendChild(delBtn);
            card.appendChild(actions);
            listEl.appendChild(card);
        });

        updateControls();
    };

    const load = () => {
        clearList();
        if (out) { out.className = 'form-output'; out.textContent = 'Loading...'; }
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
            .then(renderUsers)
            .catch(err => {
                clearList();
                if (out) { out.className = 'form-output error'; out.textContent = err && err.error ? err.error : (err && err.message ? err.message : String(err)); }
                updateControls();
            });
    };

    const openModal = () => {
        if (!modal) return;
        modal.classList.remove('hidden');
        window.scrollTo(0, 0);
    };
    const closeModal = () => { if (!modal) return; modal.classList.add('hidden'); };

    const resetModal = () => {
        modalOutput.textContent = '';
        tabResidentBtn.classList.add('hidden');
        tabStaffBtn.classList.add('hidden');
        residentPanel.classList.add('hidden');
        staffPanel.classList.add('hidden');
        housesList.innerHTML = '';
        specsList.innerHTML = '';
    };

    const openDetails = async (phone) => {
        resetModal();
        openModal();
        if (modalOutput) { modalOutput.textContent = 'Loading...'; modalOutput.className = 'form-output'; }

        try {
            const res = await fetch('/api/staff/users/info/' + encodeURIComponent(phone), { method: 'GET', credentials: 'same-origin' });
            const data = await res.json();
            if (!res.ok) {
                modalOutput.textContent = data.error || 'Failed to fetch details';
                modalOutput.className = 'form-output error';
                return;
            }

            if (data.resident) {
                tabResidentBtn.classList.remove('hidden');
                resIdEl.textContent = data.resident.ID || data.resident.id || '—';
                resPhoneEl.textContent = data.resident.Phone || data.resident.phone || '—';
                resFullEl.textContent = data.resident.FullName || data.resident.full_name || '—';
                btnGetHouses.onclick = async () => {
                    housesList.textContent = 'Loading...';
                    try {
                        const r = await fetch('/api/staff/users/resident/info?residentID=' + encodeURIComponent(resIdEl.textContent), { credentials: 'same-origin' });
                        const jd = await r.json();
                        if (!r.ok) { housesList.textContent = jd.error || 'Failed'; return; }
                        renderHouses(jd.houses || []);
                    } catch (err) { housesList.textContent = 'Network error'; }
                };
            }

            if (data.staff) {
                tabStaffBtn.classList.remove('hidden');
                staffIdEl.textContent = data.staff.ID || data.staff.id || String(data.staff.ID || data.staff.id || '');
                staffPhoneEl.textContent = data.staff.Phone || data.staff.phone || '—';
                staffFullEl.textContent = data.staff.FullName || data.staff.full_name || '—';
                staffStatusEl.textContent = data.staff.Status || data.staff.status || '—';

                btnGetSpecs.onclick = async () => {
                    specsList.textContent = 'Loading...';
                    try {
                        const r = await fetch('/api/staff/users/staff/info?staffMemberID=' + encodeURIComponent(staffIdEl.textContent), { credentials: 'same-origin' });
                        const jd = await r.json();
                        if (!r.ok) { specsList.textContent = jd.error || 'Failed'; return; }
                        renderSpecs(jd.specializations || []);
                    } catch (err) { specsList.textContent = 'Network error'; }
                };
            }

            if (modalOutput) { modalOutput.textContent = ''; modalOutput.className = 'form-output'; }
        } catch (err) {
            if (modalOutput) { modalOutput.textContent = 'Network error'; modalOutput.className = 'form-output error'; }
        }
    };

    const renderHouses = (houses) => {
        housesList.innerHTML = '';
        if (!houses.length) { housesList.textContent = 'No houses'; return; }
        houses.forEach(h => {
            const row = document.createElement('div');
            row.style.display = 'flex';
            row.style.alignItems = 'center';
            row.style.justifyContent = 'space-between';
            row.style.gap = '12px';
            row.style.marginBottom = '6px';

            const left = document.createElement('div');
            left.innerHTML = '<strong>ID:</strong> ' + (h.ID || h.id) + ' • ' + (h.Address || h.address || '—');

            const right = document.createElement('div');

            const delBtn = document.createElement('button');
            delBtn.className = 'btn';
            delBtn.textContent = 'Delete';
            delBtn.addEventListener('click', async () => {
                if (!confirm('Remove house ' + (h.ID || h.id) + ' from resident?')) return;
                try {
                    const url = '/api/staff/users/resident/remove-house?residentID=' + encodeURIComponent(resIdEl.textContent) + '&houseID=' + encodeURIComponent(h.ID || h.id);
                    const r = await fetch(url, { method: 'GET', credentials: 'same-origin' });
                    const jd = await r.json();
                    if (!r.ok) { alert(jd.error || 'Remove failed'); return; }
                    btnGetHouses.click();
                } catch (err) { alert('Network error'); }
            });

            right.appendChild(delBtn);
            row.appendChild(left);
            row.appendChild(right);
            housesList.appendChild(row);
        });
    };

    const renderSpecs = (specs) => {
        specsList.innerHTML = '';
        if (!specs.length) { specsList.textContent = 'No specializations'; return; }
        specs.forEach(s => {
            const row = document.createElement('div');
            row.style.display = 'flex';
            row.style.alignItems = 'center';
            row.style.justifyContent = 'space-between';
            row.style.gap = '12px';
            row.style.marginBottom = '6px';

            const left = document.createElement('div');
            left.innerHTML = '<strong>ID:</strong> ' + (s.ID || s.id) + ' • ' + (s.Title || s.title || '—');

            const right = document.createElement('div');

            const deactBtn = document.createElement('button');
            deactBtn.className = 'btn';
            deactBtn.textContent = 'Deactivate';
            deactBtn.addEventListener('click', async () => {
                if (!confirm('Deactivate specialization ' + (s.ID || s.id) + '?')) return;
                try {
                    const url = '/api/staff/users/staff/delete-spec?staffMemberID=' + encodeURIComponent(staffIdEl.textContent) + '&jobID=' + encodeURIComponent(s.ID || s.id);
                    const r = await fetch(url, { method: 'GET', credentials: 'same-origin' });
                    const jd = await r.json();
                    if (!r.ok) { alert(jd.error || 'Deactivate failed'); return; }
                    btnGetSpecs.click();
                } catch (err) { alert('Network error'); }
            });

            right.appendChild(deactBtn);
            row.appendChild(left);
            row.appendChild(right);
            specsList.appendChild(row);
        });
    };

    if (prevBtn) prevBtn.addEventListener('click', () => { if (page > 1) { page--; load(); }});
    if (nextBtn) nextBtn.addEventListener('click', () => { if (page < lastPages) { page++; load(); }});
    if (searchBtn) searchBtn.addEventListener('click', () => { page = 1; load(); });
    if (limitSelect) limitSelect.addEventListener('change', (e) => { const v = parseInt(e.target.value || '20', 10); if (!isNaN(v) && v>0) { limit = v; page = 1; load(); }});
    if (modalClose) modalClose.addEventListener('click', () => closeModal());

    if (tabResidentBtn) tabResidentBtn.addEventListener('click', () => {
        residentPanel.classList.remove('hidden');
        staffPanel.classList.add('hidden');
    });
    if (tabStaffBtn) tabStaffBtn.addEventListener('click', () => {
        staffPanel.classList.remove('hidden');
        residentPanel.classList.add('hidden');
    });

    if (searchInput) {
        searchInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') {
                page = 1;
                load();
            }
        });
    }

    load();
});