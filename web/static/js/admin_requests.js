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

    const modal = document.getElementById("edit-modal");
    const editForm = document.getElementById("edit-form");
    const editOutput = document.getElementById("edit-output");
    const addJobBtn = document.getElementById("add-job-lookup");
    const jobContainer = document.getElementById("job-lookup-container");
    const jobInput = document.getElementById("job-id-input");
    const jobLookupBtn = document.getElementById("job-lookup-btn");
    const editCancelBtn = document.getElementById("edit-cancel");

    const organizationBlock = document.getElementById("organization-block");
    const organizationInput = document.getElementById("edit-organizationID");
    const statusSelect = document.getElementById("edit-status");

    const toggleOrganizationField = (status) => {
        if (!organizationBlock || !organizationInput) return;
        if (status === "передана_организации") {
            organizationBlock.classList.remove("hidden");
            organizationInput.required = true;
        } else {
            organizationBlock.classList.add("hidden");
            organizationInput.required = false;
            organizationInput.value = "";
        }
    };

    if (statusSelect) {
        statusSelect.addEventListener("change", (e) => toggleOrganizationField(e.target.value));
    }

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

    const openModal = (req) => {
        if (!modal) return;
        modal.classList.remove("hidden");

        document.getElementById("edit-id").value = req.ID || "";
        document.getElementById("edit-residentID").value = req.ResidentID || "";

        const houseEl = document.getElementById("edit-houseID");
        houseEl.value = (req.HouseID !== undefined && req.HouseID !== null) ? String(req.HouseID) : "";

        const costEl = document.getElementById("edit-cost");
        if (req.Cost !== undefined && req.Cost !== null && req.Cost !== "") {
            const c = parseFloat(req.Cost);
            costEl.value = Number.isFinite(c) ? c.toFixed(2) : "";
        } else {
            costEl.value = "";
        }

        const respEl = document.getElementById("edit-responsibleID");
        respEl.value = (req.ResponsibleID !== undefined && req.ResponsibleID !== null && req.ResponsibleID !== "") ? String(req.ResponsibleID) : "";

        document.getElementById("edit-type").value = req.RequestType || "";
        document.getElementById("edit-complaint").value = req.Complaint || "";
        document.getElementById("edit-status").value = req.Status || "";

        const orgEl = document.getElementById("edit-organizationID");
        if (orgEl) {
            orgEl.value = (req.OrganizationID !== undefined && req.OrganizationID !== null) ? String(req.OrganizationID) : "";
        }
        toggleOrganizationField(req.Status || (statusSelect ? statusSelect.value : ""));

        if (jobContainer) jobContainer.classList.add("hidden");
        if (editOutput) { editOutput.textContent = ""; editOutput.className = "form-output"; }
        window.scrollTo(0, 0);
    };

    const closeModal = () => {
        if (!modal) return;
        modal.classList.add("hidden");
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

            const actions = document.createElement('div');
            actions.style.marginTop = '8px';
            actions.style.display = 'flex';
            actions.style.gap = '8px';

            const editBtn = document.createElement('button');
            editBtn.className = 'btn';
            editBtn.textContent = 'Edit';
            editBtn.addEventListener('click', () => {
                openModal(r);
            });

            const delBtn = document.createElement('button');
            delBtn.className = 'btn';
            delBtn.textContent = 'Delete';
            delBtn.addEventListener('click', async () => {
                if (!confirm('Delete request ' + id + '?')) return;
                try {
                    const res = await fetch('/api/staff/requests/panel/delete/' + encodeURIComponent(id), {
                        method: 'GET',
                        credentials: 'same-origin'
                    });
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

            actions.appendChild(editBtn);
            actions.appendChild(delBtn);
            card.appendChild(actions);

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

    if (addJobBtn) {
        addJobBtn.addEventListener('click', () => {
            if (!jobContainer) return;
            jobContainer.classList.toggle('hidden');
        });
    }

    if (jobLookupBtn) {
        jobLookupBtn.addEventListener('click', async () => {
            const jobID = (jobInput && jobInput.value) ? jobInput.value.trim() : '';
            if (!jobID) {
                if (editOutput) { editOutput.textContent = 'Enter jobID'; editOutput.className = 'form-output error'; }
                return;
            }
            try {
                if (editOutput) { editOutput.textContent = 'Looking up...'; editOutput.className = 'form-output'; }
                const url = '/api/staff/requests/panel/update?jobID=' + encodeURIComponent(jobID);
                const res = await fetch(url, { method: 'GET', credentials: 'same-origin' });
                const data = await res.json();
                if (!res.ok) {
                    if (editOutput) { editOutput.textContent = data.error || 'Lookup failed'; editOutput.className = 'form-output error'; }
                    return;
                }
                if (data.leastBusy !== undefined && data.leastBusy !== null) {
                    document.getElementById("edit-responsibleID").value = String(data.leastBusy);
                    if (editOutput) { editOutput.textContent = 'Found responsible ID: ' + data.leastBusy; editOutput.className = 'form-output success'; }
                } else {
                    if (editOutput) { editOutput.textContent = 'No responsible found'; editOutput.className = 'form-output error'; }
                }
            } catch (err) {
                if (editOutput) { editOutput.textContent = 'Network error'; editOutput.className = 'form-output error'; }
            }
        });
    }

    if (editCancelBtn) {
        editCancelBtn.addEventListener('click', (e) => {
            e.preventDefault();
            closeModal();
        });
    }

    if (editForm) {
        editForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            if (editOutput) { editOutput.textContent = 'Saving...'; editOutput.className = 'form-output'; }
            const endpoint = editForm.dataset.endpoint || '/api/staff/requests/panel/update';
            const body = new FormData(editForm);
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
                    if (editOutput) { editOutput.textContent = data.error || data.raw || ('HTTP ' + res.status); editOutput.className = 'form-output error'; }
                    return;
                }
                if (editOutput) { editOutput.textContent = 'Saved'; editOutput.className = 'form-output success'; }
                closeModal();
                load();
            } catch (err) {
                if (editOutput) { editOutput.textContent = 'Network error'; editOutput.className = 'form-output error'; }
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
