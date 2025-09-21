"use strict";

(() => {
    const asJSON = async (res) => {
        const text = await res.text();
        try {
            return JSON.parse(text || "{}");
        } catch {
            return { raw: text };
        }
    };

    const handleSubmit = (formId, outputId) => {
        const form = document.getElementById(formId);
        if (!form)
            return;
        const out = document.getElementById(outputId);

        form.addEventListener("submit", async (e) => {
            e.preventDefault();
            const endpoint = form.dataset.endpoint || "Gay";
            if (!endpoint || endpoint === "Gay") {
                out.textContent = "Endpoint is not set or there is no endpoint.";
                out.className = "form-output error";
                return;
            }

            const btn = form.querySelector("button[type=submit]");
            if (btn)
                btn.disabled = true;
            out.textContent = "Submitting...";
            out.className = "form-output";

            try {
                const res = await fetch(endpoint, {
                    method: "POST",
                    body: new FormData(form),
                    credentials: "include"
                });

                const data = await asJSON(res);
                if (!res.ok) {
                    out.textContent = data.error || JSON.stringify(data) || `HTTP ${res.status}`;
                    out.className = "form-output error";
                } else {
                    out.textContent = "Success: " + (data.message || "");
                    if (data.type && data.type.toLowerCase() === "login") {
                        window.location.href = "/";
                    }
                    out.className = "form-output success";
                }
            } catch (err) {
                out.textContent = "Network error";
                out.className = "form-output error";
            } finally {
                if (btn)
                    btn.disabled = false;
            }
        });
    };

    handleSubmit("login-form", "login-output");
    handleSubmit("register-form", "register-output");
    handleSubmit("request-form", "request-output");

    const initUserDropdown = () => {
        const toggleBtn = document.getElementById('user-toggle');
        const menu = document.getElementById('user-menu');
        if (!toggleBtn || !menu) return;

        const setExpanded = (v) => {
            toggleBtn.setAttribute('aria-expanded', v ? 'true' : 'false');
            menu.setAttribute('aria-hidden', v ? 'false' : 'true');
            menu.style.display = v ? 'block' : 'none';
        };

        toggleBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            const isOpen = toggleBtn.getAttribute('aria-expanded') === 'true';
            setExpanded(!isOpen);
        });

        document.addEventListener('click', () => setExpanded(false));

        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') setExpanded(false);
        });

        menu.addEventListener('click', (e) => e.stopPropagation());
    };

    initUserDropdown();
})();