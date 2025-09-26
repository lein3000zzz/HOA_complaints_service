"use strict";

document.addEventListener("DOMContentLoaded", () => {
    let page = 1;
    let limit = 10;
    let lastPages = 1;

    const list = document.getElementById("orgs-list");
    const out = document.getElementById("orgs-output");
    const totalCountEl = document.getElementById("total-count");
    const currentPageEl = document.getElementById("current-page");
    const totalPagesEl = document.getElementById("total-pages");

    const prevBtn = document.getElementById("prev-page");
    const nextBtn = document.getElementById("next-page");
    const refreshBtn = document.getElementById("refresh-btn");
    const limitSelect = document.getElementById("limit-select");

    const patternInput = document.getElementById("pattern-input");
    const applyBtn = document.getElementById("apply-filters");

    const addOrgBtn = document.getElementById("add-org");
    const addOrgModal = document.getElementById("add-org-modal");
    const addOrgForm = document.getElementById("add-org-form");
    const addOrgOutput = document.getElementById("add-org-output");
    const addOrgCancel = document.getElementById("add-org-cancel");

    const updateOrgModal = document.getElementById("update-org-modal");
    const updateOrgForm = document.getElementById("update-org-form");
    const updateOrgId = document.getElementById("update-org-id");
    const updateOrgName = document.getElementById("update-org-name");
    const updateOrgCancel = document.getElementById("update-org-cancel");
    const updateOrgOutput = document.getElementById("update-org-output");

    const toggleAddModal = (show) => {
        if (!addOrgModal) return;
        if (show) addOrgModal.classList.remove("hidden");
        else addOrgModal.classList.add("hidden");
        window.scrollTo(0, 0);
    };

    const toggleUpdateModal = (show) => {
        if (!updateOrgModal) return;
        if (show) updateOrgModal.classList.remove("hidden");
        else updateOrgModal.classList.add("hidden");
        window.scrollTo(0, 0);
    };

    if (addOrgBtn) addOrgBtn.addEventListener("click", () => toggleAddModal(true));
    if (addOrgCancel) addOrgCancel.addEventListener("click", (e) => { e.preventDefault(); toggleAddModal(false); });

    if (addOrgForm) {
        addOrgForm.addEventListener("submit", async (e) => {
            e.preventDefault();
            if (addOrgOutput) { addOrgOutput.textContent = "Saving..."; addOrgOutput.className = "form-output"; }
            const formData = new FormData(addOrgForm);
            try {
                const res = await fetch(addOrgForm.dataset.endpoint || "/api/staff/organizations/create", {
                    method: "POST",
                    body: formData,
                    credentials: "same-origin",
                });
                const text = await res.text();
                let data;
                try { data = JSON.parse(text || "{}"); } catch { data = { raw: text }; }
                if (!res.ok) {
                    if (addOrgOutput) { addOrgOutput.textContent = data.error || data.raw || ("HTTP " + res.status); addOrgOutput.className = "form-output error"; }
                    return;
                }
                if (addOrgOutput) { addOrgOutput.textContent = "Created"; addOrgOutput.className = "form-output success"; }
                if (addOrgForm) addOrgForm.reset();
                toggleAddModal(false);
                if (typeof load === "function") load();
            } catch {
                if (addOrgOutput) { addOrgOutput.textContent = "Network error"; addOrgOutput.className = "form-output error"; }
            }
        });
    }

    if (updateOrgCancel) {
        updateOrgCancel.addEventListener("click", (e) => { e.preventDefault(); toggleUpdateModal(false); });
    }

    if (updateOrgForm) {
        updateOrgForm.addEventListener("submit", async (e) => {
            e.preventDefault();
            if (updateOrgOutput) { updateOrgOutput.textContent = "Saving..."; updateOrgOutput.className = "form-output"; }

            const id = (updateOrgId && updateOrgId.value.trim()) || "";
            const nameVal = (updateOrgName && updateOrgName.value.trim()) || "";

            if (!id) {
                if (updateOrgOutput) { updateOrgOutput.textContent = "Organization ID is empty"; updateOrgOutput.className = "form-output error"; }
                return;
            }
            if (!nameVal) {
                if (updateOrgOutput) { updateOrgOutput.textContent = "Name is required"; updateOrgOutput.className = "form-output error"; }
                return;
            }

            const endpoint = updateOrgForm.dataset.endpoint || "/api/staff/organizations/update";
            const formData = new FormData();
            formData.append("organizationID", id);
            formData.append("name", nameVal);

            try {
                const res = await fetch(endpoint, {
                    method: "POST",
                    body: formData,
                    credentials: "same-origin",
                });
                const text = await res.text();
                let data;
                try { data = JSON.parse(text || "{}"); } catch { data = { raw: text }; }

                if (!res.ok) {
                    if (updateOrgOutput) { updateOrgOutput.textContent = data.error || data.raw || ("HTTP " + res.status); updateOrgOutput.className = "form-output error"; }
                    return;
                }

                if (updateOrgOutput) { updateOrgOutput.textContent = "Updated"; updateOrgOutput.className = "form-output success"; }
                toggleUpdateModal(false);
                if (typeof load === "function") load();
            } catch {
                if (updateOrgOutput) { updateOrgOutput.textContent = "Network error"; updateOrgOutput.className = "form-output error"; }
            }
        });
    }

    const buildUrl = () => {
        const url = new URL("/api/staff/organizations/list", window.location.origin);
        url.searchParams.set("page", String(page));
        url.searchParams.set("limit", String(limit));
        const pattern = patternInput && patternInput.value.trim();
        if (pattern) url.searchParams.set("pattern", pattern);
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
        const orgs = data.organizations || [];

        const meta = data.meta || {};
        if (typeof meta.page === "number") page = meta.page;
        if (typeof meta.pages === "number") lastPages = meta.pages;
        else if (typeof data.total === "number") lastPages = Math.max(1, Math.ceil(data.total / limit));

        if (totalCountEl) {
            const total = (typeof meta.total === "number") ? meta.total : (data.total || 0);
            totalCountEl.textContent = String(total);
        }

        updateControls();

        if (!orgs.length) {
            if (out) out.textContent = "No organizations found";
            return;
        }

        orgs.forEach(o => {
            const idText = (o.ID !== undefined && o.ID !== null) ? String(o.ID) : "";
            const nameText = o.Name || "";

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
            idValue.className = "org-id";
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

            const nameRow = document.createElement("div");
            const nameLabel = document.createElement("span");
            nameLabel.style.color = "var(--muted)";
            nameLabel.textContent = "Name: ";
            const nameValue = document.createElement("span");
            nameValue.textContent = nameText;
            nameRow.appendChild(nameLabel);
            nameRow.appendChild(nameValue);

            info.appendChild(idRow);
            info.appendChild(document.createElement("br"));
            info.appendChild(nameRow);

            const actions = document.createElement("div");
            actions.className = "form-row";
            actions.style.marginTop = "8px";

            const editBtn = document.createElement("button");
            editBtn.className = "btn";
            editBtn.type = "button";
            editBtn.textContent = "Edit name";
            editBtn.addEventListener("click", () => {
                if (updateOrgId) updateOrgId.value = idText;
                if (updateOrgName) updateOrgName.value = nameText;
                if (updateOrgOutput) { updateOrgOutput.textContent = ""; updateOrgOutput.className = "form-output"; }
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
        fetch(buildUrl(), { credentials: "same-origin" })
            .then(async res => {
                const text = await res.text();
                try {
                    const json = JSON.parse(text || "{}");
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

    if (prevBtn) prevBtn.addEventListener("click", () => { if (page > 1) { page--; load(); } });
    if (nextBtn) nextBtn.addEventListener("click", () => { if (page < lastPages) { page++; load(); } });
    if (applyBtn) applyBtn.addEventListener("click", () => { page = 1; load(); });
    if (refreshBtn) refreshBtn.addEventListener("click", () => load());
    if (limitSelect) limitSelect.addEventListener("change", (e) => {
        const v = parseInt(e.target.value || "10", 10);
        if (!isNaN(v) && v > 0) { limit = v; page = 1; load(); }
    });

    load();
});