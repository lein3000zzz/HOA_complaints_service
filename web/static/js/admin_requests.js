"use strict";

document.addEventListener("DOMContentLoaded", () => {
    let page = 1;
    let limit = 20;
    let lastPages = 1;
    let sort = "created_desc";

    const list = document.getElementById("requests-list");
    const out = document.getElementById("requests-output");
    const totalCountEl = document.getElementById("total-count");
    const currentPageEl = document.getElementById("current-page");
    const totalPagesEl = document.getElementById("total-pages");

    const prevBtn = document.getElementById("prev-page");
    const nextBtn = document.getElementById("next-page");
    const refreshBtn = document.getElementById("refresh-btn");
    const sortSelect = document.getElementById("sort-select");
    const limitSelect = document.getElementById("limit-select");

    const filterId = document.getElementById("filter-id");
    const filterResident = document.getElementById("filter-resident-id");
    const filterHouse = document.getElementById("filter-house");
    const filterResponsible = document.getElementById("filter-resp");
    const filterOrg = document.getElementById("filter-organization-id");
    const filterType = document.getElementById("filter-type");
    const filterStatus = document.getElementById("filter-status");
    const filterComplaint = document.getElementById("filter-complaint");
    const applyBtn = document.getElementById("apply-filters");

    const clear = () => {
        if (list) list.innerHTML = "";
        if (out) { out.textContent = ""; out.className = "form-output"; }
    };

    const updateControls = () => {
        if (currentPageEl) currentPageEl.textContent = String(page);
        if (totalPagesEl) totalPagesEl.textContent = String(lastPages);
    };

    const buildUrl = () => {
        const url = new URL('/api/staff/requests/panel', window.location.origin);
        url.searchParams.set('page', String(page));
        url.searchParams.set('limit', String(limit));
        if (sort) url.searchParams.set('sort', sort);

        if (filterId && filterId.value) url.searchParams.set('id', filterId.value);
        if (filterResident && filterResident.value) url.searchParams.set('residentID', filterResident.value);
        if (filterHouse && filterHouse.value) url.searchParams.set('houseID', filterHouse.value);
        if (filterResponsible && filterResponsible.value) url.searchParams.set('responsibleID', filterResponsible.value);
        if (filterOrg && filterOrg.value) url.searchParams.set('organizationID', filterOrg.value);
        if (filterType && filterType.value) url.searchParams.set('type', filterType.value);
        if (filterStatus && filterStatus.value) url.searchParams.set('status', filterStatus.value);
        if (filterComplaint && filterComplaint.value) url.searchParams.set('complaint', filterComplaint.value);

        return url.toString();
    };

    const renderRequests = (data) => {
        clear();

        const requests = data.requests || [];
        const meta = data.meta || {};

        const total = (typeof meta.total === 'number') ? meta.total : 0;
        const pageFromMeta = (meta.page && typeof meta.page === 'number') ? meta.page : page;
        const pages = (meta.pages && typeof meta.pages === 'number') ? meta.pages : Math.max(1, Math.ceil(total / limit));

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
            const residentID = r.ResidentID || '';
            const houseID = r.HouseID || '';
            const responsible = r.ResponsibleID || '';
            const organization = r.OrganizationID || '';

            const createdStr = created ? (new Date(created)).toLocaleString() : '';

            const orgPart = (status === 'передана_организации' && organization) ? (' • organization: ' + organization) : '';

            card.innerHTML = '<div style="font-weight:700;margin-bottom:6px;">' +
                '<span style="color:var(--muted);">ID: </span> ' + id + '<br><span style="color:var(--muted);">Resident: </span>' + residentID + '<br><span style="color:var(--muted);">House: </span>' + houseID + '<br><span style="color:var(--muted);">Тип: </span>' + type + '<br><span style="color:var(--muted);">Статус: </span>' + status +
                '</div>' +
                '<div style="margin-bottom:8px;">' + (complaint || '') + '</div>' +
                '<div style="font-size:12px;color:var(--muted);">' + createdStr + (responsible ? (' • responsible: '+responsible) : '') + orgPart + '</div>';

            list.appendChild(card);
        });

        updateControls();
    };

    const load = () => {
        clear();
        if (out) {
            out.textContent = 'Loading...';
            out.className = 'form-output';
        }

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
        if (page > 1) { page--; load(); }
    });
    if (nextBtn) nextBtn.addEventListener('click', () => {
        if (page < lastPages) { page++; load(); }
    });
    if (refreshBtn) refreshBtn.addEventListener('click', () => load());
    if (sortSelect) sortSelect.addEventListener('change', (e) => {
        sort = e.target.value || '';
        page = 1;
        load();
    });
    if (limitSelect) limitSelect.addEventListener('change', (e) => {
        const v = parseInt(e.target.value || '20', 10);
        if (!isNaN(v) && v > 0) { limit = v; page = 1; load(); }
    });
    if (applyBtn) applyBtn.addEventListener('click', () => { page = 1; load(); });

    load();
});
