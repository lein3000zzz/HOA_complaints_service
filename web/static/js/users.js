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
    const btnAddHouse = document.getElementById("btn-add-house");

    const staffIdEl = document.getElementById("staff-id");
    const staffPhoneEl = document.getElementById("staff-phone");
    const staffFullEl = document.getElementById("staff-fullname");
    const staffStatusEl = document.getElementById("staff-status");
    const specsList = document.getElementById("specs-list");
    const btnGetSpecs = document.getElementById("btn-get-specs");
    const btnAddSpec = document.getElementById("btn-add-spec");

    const createModal = document.getElementById("create-modal");
    const createTitle = document.getElementById("create-title");
    const createLabel = document.getElementById("create-label");
    const createField = document.getElementById("create-field");
    const createHint = document.getElementById("create-hint");
    const createSave = document.getElementById("create-save");
    const createCancel = document.getElementById("create-cancel");
    const createOutput = document.getElementById("create-output");
    const createForm = document.getElementById("create-form");

    let currentAddMode = null;

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
        if (modalOutput) { modalOutput.textContent = ''; modalOutput.className = 'form-output'; }
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
                if (btnAddHouse) {
                    btnAddHouse.onclick = () => openCreateModal('house');
                }
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
                if (btnAddSpec) {
                    btnAddSpec.onclick = () => openCreateModal('spec');
                }
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

    const openCreateModal = (mode) => {
        currentAddMode = mode;
        createField.value = '';
        createOutput.textContent = '';
        createOutput.className = 'form-output';

        if (mode === 'house') {
            createTitle.textContent = 'Add house';
            createLabel.textContent = 'House ID';
            createField.type = 'number';
            createField.min = '1';
            createField.step = '1';
            createField.placeholder = 'Enter house ID';
            createHint.textContent = 'Assign a house by its numeric ID to this resident.';
        } else {
            createTitle.textContent = 'Add specialization';
            createLabel.textContent = 'Specialization ID';
            createField.type = 'text';
            createField.removeAttribute('min');
            createField.removeAttribute('step');
            createField.placeholder = 'Enter specialization ID';
            createHint.textContent = 'Assign a specialization by its string ID to this staff member.';
        }
        createModal.classList.remove('hidden');
        createField.focus();
    };
    const closeCreateModal = () => {
        createModal.classList.add('hidden');
        currentAddMode = null;
    };

    const postForm = async (url, bodyObj) => {
        const body = new URLSearchParams();
        Object.entries(bodyObj).forEach(([k, v]) => body.append(k, String(v)));
        const res = await fetch(url, {
            method: 'POST',
            credentials: 'same-origin',
            headers: { 'Content-Type': 'application/x-www-form-urlencoded;charset=UTF-8' },
            body
        });
        const text = await res.text();
        let json;
        try { json = JSON.parse(text || '{}'); } catch { json = { raw: text }; }
        return { ok: res.ok, json };
    };

    if (createSave) {
        createForm.addEventListener('submit', async (e) => {
            e.preventDefault();

            try {
                if (currentAddMode === 'house') {
                    const idVal = parseInt(createField.value.trim(), 10);
                    if (!Number.isFinite(idVal) || idVal <= 0) {
                        createOutput.className = 'form-output error';
                        createOutput.textContent = 'Please enter a valid positive number.';
                        return;
                    }
                    const residentID = resIdEl.textContent.trim();
                    const url = '/api/staff/users/resident/add-house?residentID=' + encodeURIComponent(residentID);
                    const { ok, json } = await postForm(url, { houseID: idVal });
                    if (!ok) throw new Error(json.error || 'Add house failed');
                    createOutput.className = 'form-output success';
                    createOutput.textContent = 'House added.';
                    btnGetHouses && btnGetHouses.click();
                    setTimeout(closeCreateModal, 400);
                } else if (currentAddMode === 'spec') {
                    const specId = createField.value.trim();
                    if (!specId) {
                        createOutput.className = 'form-output error';
                        createOutput.textContent = 'Please enter a specialization ID.';
                        return;
                    }
                    const staffID = staffIdEl.textContent.trim();
                    const url = '/api/staff/users/staff/add-specialization?staffMemberID=' + encodeURIComponent(staffID);
                    const { ok, json } = await postForm(url, { specializationID: specId });
                    if (!ok) throw new Error(json.error || 'Add specialization failed');
                    createOutput.className = 'form-output success';
                    createOutput.textContent = 'Specialization added.';
                    btnGetSpecs && btnGetSpecs.click();
                    setTimeout(closeCreateModal, 400);
                }
            } catch (err) {
                createOutput.className = 'form-output error';
                createOutput.textContent = err.message || 'Network or server error.';
            }
        });
    }
    if (createCancel) createCancel.addEventListener('click', () => closeCreateModal());

    if (tabResidentBtn) tabResidentBtn.addEventListener('click', () => {
        residentPanel.classList.remove('hidden');
        staffPanel.classList.add('hidden');
    });
    if (tabStaffBtn) tabStaffBtn.addEventListener('click', () => {
        staffPanel.classList.remove('hidden');
        residentPanel.classList.add('hidden');
    });

    if (prevBtn) prevBtn.addEventListener('click', () => { if (page > 1) { page--; load(); }});
    if (nextBtn) nextBtn.addEventListener('click', () => { if (page < lastPages) { page++; load(); }});
    if (searchBtn) searchBtn.addEventListener('click', () => { page = 1; load(); });
    if (limitSelect) limitSelect.addEventListener('change', (e) => { const v = parseInt(e.target.value || '20', 10); if (!isNaN(v) && v>0) { limit = v; page = 1; load(); }});
    if (modalClose) modalClose.addEventListener('click', () => closeModal());
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