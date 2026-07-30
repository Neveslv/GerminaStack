(function () {
  const state = {
    toastRail: null,
    lastFocusedElement: null,
    focusTrapStack: [],
  };

  function qs(root, sel) {
    return root ? root.querySelector(sel) : null;
  }

  function qsa(root, sel) {
    return root ? Array.from(root.querySelectorAll(sel)) : [];
  }

  function escapeHtml(value) {
    return String(value || "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#039;");
  }

  function createLiveRegion() {
    const region = document.createElement("div");
    region.id = "gs-live-region";
    region.className = "gs-sr-only";
    region.setAttribute("aria-live", "polite");
    region.setAttribute("aria-atomic", "true");
    document.body.appendChild(region);
    return region;
  }

  function announceToScreenReader(message, priority = "polite") {
    const liveRegion = document.getElementById("gs-live-region") || createLiveRegion();
    liveRegion.setAttribute("aria-live", priority);
    liveRegion.textContent = message;
    window.setTimeout(() => {
      liveRegion.textContent = "";
    }, 1000);
  }

  async function request(url, options = {}) {
    const { body, headers, ...init } = options;
    const isJson = body && typeof body === "object" && !(body instanceof FormData) && !(body instanceof URLSearchParams) && !(body instanceof Blob);
    const requestHeaders = new Headers(headers);
    if (!requestHeaders.has("Accept")) requestHeaders.set("Accept", "application/json");
    if (isJson && !requestHeaders.has("Content-Type")) requestHeaders.set("Content-Type", "application/json");

    const response = await fetch(url, { ...init, headers: requestHeaders, body: isJson ? JSON.stringify(body) : body });
    const text = await response.text();
    let data = null;
    try { data = text ? JSON.parse(text) : null; } catch { data = text; }
    if (response.ok) return data;

    const error = new Error(data?.message || `Falha na requisição (${response.status})`);
    error.status = response.status;
    error.data = data;
    throw error;
  }

  function hasAccessibleName(element) {
    if (element.getAttribute("aria-label")?.trim()) return true;
    const labelledBy = element.getAttribute("aria-labelledby");
    if (labelledBy && labelledBy.split(/\s+/).some((id) => document.getElementById(id)?.textContent.trim())) return true;
    if ("labels" in element && element.labels?.length) return true;
    if (element.closest("label")) return true;
    return /[A-Za-zÀ-ÿ0-9]/.test(element.textContent || "");
  }

  function validateAccessibility(root) {
    const scope = root || document.body;
    const elements = (selector) => qsa(scope, selector);
    const report = (element, message) => {
      if (!element.hasAttribute("data-gs-a11y-error")) console.error(`[GerminaStackUI] ${message}`, element);
      element.setAttribute("data-gs-a11y-error", message);
    };
    const clear = (element) => element.removeAttribute("data-gs-a11y-error");

    elements("img").forEach((image) => {
      if (image.hasAttribute("alt")) clear(image);
      else report(image, "Imagens precisam de alt. Use alt=\"\" apenas se forem decorativas.");
    });
    elements('input:not([type="hidden"]), textarea, select').forEach((control) => {
      if (hasAccessibleName(control)) clear(control);
      else report(control, "Campos precisam de <label>, aria-label ou aria-labelledby.");
    });
    elements("button, a[href]").forEach((control) => {
      if (hasAccessibleName(control)) clear(control);
      else report(control, "Botões e links precisam de texto ou aria-label.");
    });
    return elements("[data-gs-a11y-error]");
  }

  function icon(name) {
    const paths = {
      close: '<path d="M18 6 6 18"></path><path d="m6 6 12 12"></path>',
      like: '<path d="M19.5 12.6 12 20l-7.5-7.4a5 5 0 0 1 7.1-7.1l.4.4.4-.4a5 5 0 0 1 7.1 7.1Z"></path>',
      comment: '<path d="M21 15a4 4 0 0 1-4 4H8l-5 3V7a4 4 0 0 1 4-4h10a4 4 0 0 1 4 4Z"></path>',
      bookmark: '<path d="M6 4a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2v18l-6-4-6 4Z"></path>',
    };
    return `<svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">${paths[name] || ""}</svg>`;
  }

  function ensureToastRail() {
    if (state.toastRail && document.body.contains(state.toastRail)) return state.toastRail;
    let rail = document.querySelector("[data-gs-toast-rail]");
    if (!rail) {
      rail = document.createElement("div");
      rail.className = "gs-toast-rail";
      rail.setAttribute("data-gs-toast-rail", "");
      rail.setAttribute("role", "region");
      rail.setAttribute("aria-live", "polite");
      rail.setAttribute("aria-label", "Notificacoes");
      document.body.appendChild(rail);
    }
    state.toastRail = rail;
    return rail;
  }

  function showToast(opts) {
    const cfg = Object.assign({ title: "Aviso", message: "", tone: "", duration: 3200 }, opts || {});
    const rail = ensureToastRail();
    const item = document.createElement("div");
    item.className = `gs-toast ${cfg.tone ? `is-${cfg.tone}` : ""}`.trim();
    item.setAttribute("role", cfg.tone === "danger" || cfg.tone === "error" ? "alert" : "status");
    item.innerHTML = `<div><strong>${escapeHtml(cfg.title)}</strong><p>${escapeHtml(cfg.message)}</p></div><button class="gs-dismiss" type="button" aria-label="Fechar aviso">${icon("close")}</button>`;
    rail.appendChild(item);
    announceToScreenReader(`${cfg.title}: ${cfg.message}`, cfg.tone === "danger" || cfg.tone === "error" ? "assertive" : "polite");
    const remove = () => item.remove();
    qs(item, ".gs-dismiss").addEventListener("click", remove);
    window.setTimeout(remove, cfg.duration);
  }

  function nextCount(button, delta) {
    const countNode = qs(button, "[data-gs-count]");
    if (!countNode) return;
    const value = Number(countNode.textContent || "0") + delta;
    countNode.textContent = String(Math.max(0, value));
  }

  function closeMenus() {
    qsa(document, "[data-gs-menu-trigger]").forEach((trigger) => trigger.setAttribute("aria-expanded", "false"));
    qsa(document, "[data-gs-menu-panel]").forEach((panel) => {
      panel.hidden = true;
    });
  }

  function trapFocus(element) {
    const nodes = qsa(element, 'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])');
    const first = nodes[0];
    const last = nodes[nodes.length - 1];
    function handle(event) {
      if (event.key !== "Tab" || !first || !last) return;
      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first.focus();
      }
    }
    element.addEventListener("keydown", handle);
    return () => element.removeEventListener("keydown", handle);
  }

  function openModal(id) {
    const modal = document.querySelector(`[data-gs-modal="${id}"]`);
    if (!modal) return;
    state.lastFocusedElement = document.activeElement;
    modal.hidden = false;
    document.body.style.overflow = "hidden";
    state.focusTrapStack.push({ modal, cleanup: trapFocus(modal) });
    const target = qs(modal, "[data-gs-modal-close]") || qs(modal, "button, input, textarea, select, [href]");
    if (target) target.focus();
    announceToScreenReader("Dialogo aberto", "polite");
  }

  function closeModal(node) {
    const modal = node.closest("[data-gs-modal]");
    if (!modal) return;
    modal.hidden = true;
    document.body.style.overflow = "";
    const idx = state.focusTrapStack.findIndex((entry) => entry.modal === modal);
    if (idx >= 0) {
      state.focusTrapStack[idx].cleanup();
      state.focusTrapStack.splice(idx, 1);
    }
    if (state.lastFocusedElement && document.body.contains(state.lastFocusedElement)) state.lastFocusedElement.focus();
    announceToScreenReader("Dialogo fechado", "polite");
  }

  function activateTab(button, focusPanel) {
    const root = button.closest("[data-gs-tabs]");
    if (!root) return;
    const id = button.getAttribute("data-gs-tab");
    qsa(root, "[data-gs-tab]").forEach((tab) => {
      const active = tab === button;
      tab.classList.toggle("is-active", active);
      tab.setAttribute("aria-selected", String(active));
      tab.tabIndex = active ? 0 : -1;
    });
    qsa(root, "[data-gs-panel]").forEach((panel) => {
      const active = panel.getAttribute("data-gs-panel") === id;
      panel.hidden = !active;
      panel.setAttribute("role", "tabpanel");
      panel.setAttribute("aria-hidden", String(!active));
      if (!panel.hasAttribute("tabindex")) panel.tabIndex = -1;
      if (active && focusPanel) panel.focus();
    });
  }

  function toggleAccordion(trigger) {
    const item = trigger.closest("[data-gs-accordion-item]");
    const panel = item ? qs(item, "[data-gs-accordion-panel]") : null;
    if (!panel) return;
    const expanded = trigger.getAttribute("aria-expanded") === "true";
    trigger.setAttribute("aria-expanded", String(!expanded));
    panel.hidden = expanded;
    const iconNode = qs(trigger, "[data-gs-accordion-icon]");
    if (iconNode) iconNode.textContent = expanded ? "+" : "-";
  }

  function initTabs(root) {
    qsa(root, "[data-gs-tabs]").forEach((tabs) => {
      const active = qs(tabs, "[data-gs-tab].is-active") || qs(tabs, "[data-gs-tab]");
      qsa(tabs, "[data-gs-tab]").forEach((tab) => tab.setAttribute("role", "tab"));
      if (active) activateTab(active, false);
    });
  }

  function initAccordions(root) {
    qsa(root, "[data-gs-accordion-trigger]").forEach((trigger, index) => {
      const item = trigger.closest("[data-gs-accordion-item]");
      const panel = item ? qs(item, "[data-gs-accordion-panel]") : null;
      if (!panel) return;
      if (!panel.id) panel.id = `gs-accordion-panel-${index + 1}`;
      trigger.setAttribute("aria-controls", panel.id);
      trigger.setAttribute("aria-expanded", trigger.getAttribute("aria-expanded") === "true" ? "true" : "false");
      panel.setAttribute("role", "region");
      panel.hidden = trigger.getAttribute("aria-expanded") !== "true";
    });
  }

  function initMenus(root) {
    qsa(root, "[data-gs-menu]").forEach((menu) => {
      const trigger = qs(menu, "[data-gs-menu-trigger]");
      const panel = qs(menu, "[data-gs-menu-panel]");
      if (!trigger || !panel) return;
      trigger.setAttribute("aria-haspopup", "menu");
      if (!trigger.hasAttribute("aria-expanded")) trigger.setAttribute("aria-expanded", "false");
      panel.hidden = true;
      qsa(panel, "button, a").forEach((item) => item.setAttribute("role", "menuitem"));
    });
  }

  function initSelects(root) {
    qsa(root, "[data-gs-select]").forEach((select, index) => {
      const trigger = qs(select, ".gs-select-trigger");
      const panel = qs(select, ".gs-select-panel");
      const value = qs(select, ".gs-select-value");
      const search = qs(select, ".gs-select-search input");
      const options = panel ? qsa(panel, ".gs-select-option") : [];
      const placeholder = select.getAttribute("data-gs-select-placeholder") || "Selecionar";
      if (!trigger || !panel || !value) return;
      if (!trigger.id) trigger.id = `gs-select-trigger-${index + 1}`;
      if (!panel.id) panel.id = `gs-select-panel-${index + 1}`;
      trigger.setAttribute("aria-controls", panel.id);
      trigger.setAttribute("aria-haspopup", "listbox");
      panel.setAttribute("aria-labelledby", trigger.id);
      if (!value.textContent.trim()) {
        value.textContent = placeholder;
        value.classList.add("gs-select-placeholder");
      }

      function visibleOptions() {
        return options.filter((option) => !option.hidden);
      }

      function updateActiveOption(option) {
        options.forEach((item) => item.removeAttribute("aria-current"));
        if (option) option.setAttribute("aria-current", "true");
      }

      function closeSelect(focusTrigger = true) {
        trigger.setAttribute("aria-expanded", "false");
        panel.hidden = true;
        if (search) search.value = "";
        options.forEach((option) => {
          option.hidden = false;
          option.removeAttribute("aria-current");
        });
        if (focusTrigger) trigger.focus();
      }

      function selectOption(option) {
        options.forEach((item) => item.setAttribute("aria-selected", String(item === option)));
        value.textContent = option.textContent.trim();
        value.classList.remove("gs-select-placeholder");
        trigger.dataset.gsValue = option.getAttribute("data-value") || option.textContent.trim();
        closeSelect();
      }

      options.forEach((option) => {
        option.setAttribute("role", "option");
        if (!option.hasAttribute("aria-selected")) option.setAttribute("aria-selected", "false");
        option.tabIndex = -1;
        option.addEventListener("click", () => selectOption(option));
        option.addEventListener("keydown", (event) => {
          const items = visibleOptions();
          const pos = items.indexOf(option);
          if (event.key === "ArrowDown" && items[pos + 1]) {
            event.preventDefault();
            updateActiveOption(items[pos + 1]);
            items[pos + 1].focus();
          } else if (event.key === "ArrowUp" && items[pos - 1]) {
            event.preventDefault();
            updateActiveOption(items[pos - 1]);
            items[pos - 1].focus();
          } else if (event.key === "Home" && items[0]) {
            event.preventDefault();
            updateActiveOption(items[0]);
            items[0].focus();
          } else if (event.key === "End" && items[items.length - 1]) {
            event.preventDefault();
            updateActiveOption(items[items.length - 1]);
            items[items.length - 1].focus();
          } else if (event.key === "Enter" || event.key === " ") {
            event.preventDefault();
            selectOption(option);
          } else if (event.key === "Escape") {
            event.preventDefault();
            closeSelect();
          }
        });
      });

      trigger.addEventListener("click", (event) => {
        event.stopPropagation();
        const expanded = trigger.getAttribute("aria-expanded") === "true";
        if (expanded) {
          closeSelect();
        } else {
          trigger.setAttribute("aria-expanded", "true");
          panel.hidden = false;
          if (search) search.focus();
          else if (visibleOptions()[0]) {
            updateActiveOption(visibleOptions()[0]);
            visibleOptions()[0].focus();
          }
          announceToScreenReader("Lista de opcoes aberta", "polite");
        }
      });

      trigger.addEventListener("keydown", (event) => {
        if (event.key === "ArrowDown" || event.key === "Enter" || event.key === " ") {
          event.preventDefault();
          trigger.click();
        }
      });

      if (search) {
        search.addEventListener("input", () => {
          const query = search.value.toLowerCase();
          options.forEach((option) => {
            option.hidden = !option.textContent.toLowerCase().includes(query);
          });
          const items = visibleOptions();
          updateActiveOption(items[0] || null);
          announceToScreenReader(`${items.length} opcoes disponiveis`, "polite");
        });
        search.addEventListener("keydown", (event) => {
          if (event.key === "ArrowDown" && visibleOptions()[0]) {
            event.preventDefault();
            updateActiveOption(visibleOptions()[0]);
            visibleOptions()[0].focus();
          } else if (event.key === "Escape") {
            event.preventDefault();
            closeSelect();
          }
        });
      }

      document.addEventListener("click", (event) => {
        if (!select.contains(event.target)) closeSelect(false);
      });
    });
  }

  function initDatePickers(root) {
    qsa(root, "[data-gs-datepicker]").forEach((datepicker) => {
      const trigger = qs(datepicker, ".gs-datepicker-trigger");
      const panel = qs(datepicker, ".gs-datepicker-panel");
      const grid = panel ? qs(panel, ".gs-datepicker-grid") : null;
      const title = panel ? qs(panel, ".gs-datepicker-title") : null;
      const navButtons = panel ? qsa(panel, ".gs-datepicker-nav") : [];
      const value = trigger ? qs(trigger, ".gs-datepicker-value") : null;
      if (!trigger || !panel || !grid || !value) return;
      let currentDate = new Date();
      let selectedDate = null;

      function updateValue() {
        if (selectedDate) {
          value.textContent = selectedDate.toLocaleDateString("pt-BR");
          value.classList.remove("gs-select-placeholder");
        } else {
          value.textContent = "Selecionar data";
          value.classList.add("gs-select-placeholder");
        }
      }

      function renderCalendar() {
        const year = currentDate.getFullYear();
        const month = currentDate.getMonth();
        const weekdays = ["Dom", "Seg", "Ter", "Qua", "Qui", "Sex", "Sab"];
        const firstDay = new Date(year, month, 1);
        const lastDay = new Date(year, month + 1, 0);
        const startDay = firstDay.getDay();
        const daysInMonth = lastDay.getDate();
        const prevMonthDays = new Date(year, month, 0).getDate();
        const today = new Date();
        today.setHours(0, 0, 0, 0);
        let html = "";
        if (title) title.textContent = currentDate.toLocaleDateString("pt-BR", { month: "long", year: "numeric" });
        weekdays.forEach((day) => {
          html += `<div class="gs-datepicker-weekday" role="columnheader">${day}</div>`;
        });
        for (let i = startDay - 1; i >= 0; i--) {
          html += `<button type="button" class="gs-datepicker-day is-outside is-disabled" tabindex="-1">${prevMonthDays - i}</button>`;
        }
        for (let day = 1; day <= daysInMonth; day++) {
          const date = new Date(year, month, day);
          date.setHours(0, 0, 0, 0);
          const iso = date.toISOString().split("T")[0];
          const isSelected = selectedDate && date.getTime() === selectedDate.getTime();
          const isToday = date.getTime() === today.getTime();
          html += `<button type="button" class="gs-datepicker-day${isSelected ? " is-selected" : ""}${isToday ? " is-today" : ""}" data-date="${iso}" aria-selected="${isSelected ? "true" : "false"}">${day}</button>`;
        }
        const totalCells = startDay + daysInMonth;
        const nextMonthCells = Math.ceil(totalCells / 7) * 7 - totalCells;
        for (let day = 1; day <= nextMonthCells; day++) {
          html += `<button type="button" class="gs-datepicker-day is-outside is-disabled" tabindex="-1">${day}</button>`;
        }
        grid.innerHTML = html;
        qsa(grid, ".gs-datepicker-day:not(.is-disabled)").forEach((dayBtn) => {
          dayBtn.addEventListener("click", () => {
            selectedDate = new Date(`${dayBtn.getAttribute("data-date")}T00:00:00`);
            updateValue();
            renderCalendar();
            panel.hidden = true;
            trigger.setAttribute("aria-expanded", "false");
            announceToScreenReader(`Data selecionada ${selectedDate.toLocaleDateString("pt-BR")}`, "polite");
          });
          dayBtn.addEventListener("keydown", (event) => {
            const days = qsa(grid, ".gs-datepicker-day:not(.is-disabled)");
            const pos = days.indexOf(dayBtn);
            if (event.key === "ArrowRight" && days[pos + 1]) {
              event.preventDefault();
              days[pos + 1].focus();
            } else if (event.key === "ArrowLeft" && days[pos - 1]) {
              event.preventDefault();
              days[pos - 1].focus();
            } else if (event.key === "ArrowDown" && days[pos + 7]) {
              event.preventDefault();
              days[pos + 7].focus();
            } else if (event.key === "ArrowUp" && days[pos - 7]) {
              event.preventDefault();
              days[pos - 7].focus();
            } else if (event.key === "Home" && days[pos - (pos % 7)]) {
              event.preventDefault();
              days[pos - (pos % 7)].focus();
            } else if (event.key === "End" && days[Math.min(pos + (6 - (pos % 7)), days.length - 1)]) {
              event.preventDefault();
              days[Math.min(pos + (6 - (pos % 7)), days.length - 1)].focus();
            } else if (event.key === "PageUp") {
              event.preventDefault();
              currentDate.setMonth(currentDate.getMonth() - 1);
              renderCalendar();
            } else if (event.key === "PageDown") {
              event.preventDefault();
              currentDate.setMonth(currentDate.getMonth() + 1);
              renderCalendar();
            } else if (event.key === "Escape") {
              event.preventDefault();
              panel.hidden = true;
              trigger.setAttribute("aria-expanded", "false");
              trigger.focus();
            }
          });
        });
      }

      trigger.addEventListener("click", (event) => {
        event.stopPropagation();
        const expanded = trigger.getAttribute("aria-expanded") === "true";
        trigger.setAttribute("aria-expanded", String(!expanded));
        panel.hidden = expanded;
        if (!expanded) renderCalendar();
        if (!expanded) announceToScreenReader("Calendario aberto", "polite");
      });

      if (navButtons[0]) {
        navButtons[0].addEventListener("click", () => {
          currentDate.setMonth(currentDate.getMonth() - 1);
          renderCalendar();
        });
      }

      if (navButtons[1]) {
        navButtons[1].addEventListener("click", () => {
          currentDate.setMonth(currentDate.getMonth() + 1);
          renderCalendar();
        });
      }

      document.addEventListener("click", (event) => {
        if (!datepicker.contains(event.target)) {
          panel.hidden = true;
          trigger.setAttribute("aria-expanded", "false");
        }
      });

      updateValue();
    });
  }

  function initDataTables(root) {
    qsa(root, "[data-gs-datatable]").forEach((datatable) => {
      const table = qs(datatable, "table");
      const search = qs(datatable, ".gs-datatable-filter-input");
      const info = qs(datatable, ".gs-datatable-info");
      const buttons = qsa(datatable, ".gs-datatable-page-btn");
      const pageInfo = qs(datatable, ".gs-datatable-page-info");
      if (!table) return;
      const rows = qsa(table, "tbody tr");
      const headers = qsa(table, "th[data-gs-sort]");
      const pageSize = Number(datatable.getAttribute("data-gs-datatable-page-size") || "10");
      let currentPage = 1;
      let activeRows = [...rows];

      headers.forEach((header) => {
        header.setAttribute("aria-sort", "none");
        header.tabIndex = 0;
      });

      qsa(table, ".gs-datatable-row-select").forEach((checkbox, index) => {
        if (!checkbox.getAttribute("aria-label")) checkbox.setAttribute("aria-label", `Selecionar linha ${index + 1}`);
      });

      function compare(type, a, b) {
        if (type === "number") return (parseFloat(a) || 0) - (parseFloat(b) || 0);
        if (type === "date") return new Date(a).getTime() - new Date(b).getTime();
        return a.localeCompare(b, "pt-BR");
      }

      function render() {
        const totalPages = Math.max(1, Math.ceil(activeRows.length / pageSize));
        currentPage = Math.min(currentPage, totalPages);
        const start = (currentPage - 1) * pageSize;
        const visible = new Set(activeRows.slice(start, start + pageSize));
        rows.forEach((row) => {
          row.hidden = !visible.has(row);
          row.setAttribute("aria-hidden", String(row.hidden));
        });
        if (pageInfo) pageInfo.textContent = `Pagina ${currentPage} de ${totalPages}`;
        if (info) info.textContent = `${activeRows.length} resultado(s)`;
        buttons.forEach((button, index) => {
          if (index < 2) button.disabled = currentPage <= 1;
          if (index > 1) button.disabled = currentPage >= totalPages;
        });
      }

      function filterRows() {
        const query = search ? search.value.toLowerCase() : "";
        activeRows = rows.filter((row) => row.textContent.toLowerCase().includes(query));
        currentPage = 1;
        render();
      }

      headers.forEach((header, index) => {
        header.addEventListener("click", () => {
          const type = header.getAttribute("data-gs-sort") || "string";
          const next = header.getAttribute("aria-sort") === "ascending" ? "descending" : "ascending";
          headers.forEach((item) => item.setAttribute("aria-sort", "none"));
          header.setAttribute("aria-sort", next);
          activeRows.sort((a, b) => compare(type, a.cells[index].textContent.trim(), b.cells[index].textContent.trim()));
          if (next === "descending") activeRows.reverse();
          currentPage = 1;
          render();
          announceToScreenReader(`Tabela ordenada em ${next === "ascending" ? "ordem crescente" : "ordem decrescente"}`, "polite");
        });
        header.addEventListener("keydown", (event) => {
          if (event.key === "Enter" || event.key === " ") {
            event.preventDefault();
            header.click();
          }
        });
      });

      if (search) search.addEventListener("input", filterRows);
      if (buttons[0]) buttons[0].addEventListener("click", () => { currentPage = 1; render(); });
      if (buttons[1]) buttons[1].addEventListener("click", () => { currentPage -= 1; render(); });
      if (buttons[2]) buttons[2].addEventListener("click", () => { currentPage += 1; render(); });
      if (buttons[3]) buttons[3].addEventListener("click", () => {
        currentPage = Math.max(1, Math.ceil(activeRows.length / pageSize));
        render();
      });

      render();
    });
  }

  function initTooltips(root) {
    qsa(root, "[data-gs-tooltip], [data-gs-popover]").forEach((trigger) => {
      const isPopover = trigger.hasAttribute("data-gs-popover");
      const content = trigger.getAttribute("data-gs-tooltip-content") || trigger.getAttribute("data-gs-popover-content");
      const title = trigger.getAttribute("data-gs-popover-title");
      const placement = trigger.getAttribute("data-gs-tooltip-placement") || "top";
      const mode = trigger.getAttribute("data-gs-popover-trigger") || "click";
      if (!content) return;
      let layer = null;

      function ensureLayer() {
        if (layer) return layer;
        layer = document.createElement("div");
        layer.className = isPopover ? "gs-popover" : "gs-tooltip";
        layer.setAttribute("data-placement", placement);
        layer.setAttribute("role", isPopover ? "dialog" : "tooltip");
        layer.hidden = true;
        if (isPopover) {
          if (title) {
            const header = document.createElement("div");
            const heading = document.createElement("h4");
            const closeBtn = document.createElement("button");
            header.className = "gs-popover-header";
            heading.textContent = title;
            closeBtn.className = "gs-popover-close";
            closeBtn.type = "button";
            closeBtn.setAttribute("aria-label", "Fechar");
            closeBtn.textContent = "×";
            closeBtn.addEventListener("click", hideLayer);
            header.append(heading, closeBtn);
            layer.appendChild(header);
          }
          const body = document.createElement("div");
          body.className = "gs-popover-body";
          body.textContent = content;
          layer.appendChild(body);
        } else {
          layer.textContent = content;
        }
        document.body.appendChild(layer);
        return layer;
      }

      function positionLayer() {
        const box = ensureLayer();
        const t = trigger.getBoundingClientRect();
        const l = box.getBoundingClientRect();
        const gap = 8;
        let top = t.top - l.height - gap;
        let left = t.left + (t.width - l.width) / 2;
        if (placement === "bottom") top = t.bottom + gap;
        if (placement === "left") {
          top = t.top + (t.height - l.height) / 2;
          left = t.left - l.width - gap;
        }
        if (placement === "right") {
          top = t.top + (t.height - l.height) / 2;
          left = t.right + gap;
        }
        box.style.top = `${Math.max(8, top + window.scrollY)}px`;
        box.style.left = `${Math.max(8, left + window.scrollX)}px`;
      }

      function showLayer() {
        const box = ensureLayer();
        box.hidden = false;
        positionLayer();
        requestAnimationFrame(() => box.classList.add("is-visible"));
      }

      function hideLayer() {
        if (!layer) return;
        layer.classList.remove("is-visible");
        window.setTimeout(() => {
          if (layer) layer.hidden = true;
        }, 120);
      }

      if (isPopover && mode === "click") {
        trigger.addEventListener("click", (event) => {
          event.stopPropagation();
          if (layer && !layer.hidden) hideLayer();
          else showLayer();
        });
      } else {
        trigger.addEventListener("mouseenter", showLayer);
        trigger.addEventListener("mouseleave", hideLayer);
        trigger.addEventListener("focus", showLayer);
        trigger.addEventListener("blur", hideLayer);
      }

      document.addEventListener("click", (event) => {
        if (layer && !layer.hidden && !layer.contains(event.target) && event.target !== trigger) hideLayer();
      });
      window.addEventListener("resize", positionLayer);
      window.addEventListener("scroll", positionLayer, { passive: true });
    });
  }

  function bindComposer(root) {
    qsa(root, "[data-gs-compose]").forEach((compose) => {
      const input = qs(compose, "textarea");
      const list = document.querySelector(compose.getAttribute("data-gs-target"));
      const action = qs(compose, "[data-gs-compose-submit]");
      if (!input || !list || !action || action.dataset.bound === "true") return;
      action.dataset.bound = "true";
      action.addEventListener("click", () => {
        const text = input.value.trim();
        if (!text) return;
        const card = document.createElement("article");
        card.className = "gs-post";
        card.innerHTML = `<div class="gs-post-body"><div class="gs-post-head"><span class="gs-avatar" style="background:#ffb347">N</span><div class="gs-post-user"><div class="gs-meta"><strong>Novo Post</strong><span>2a Tec E</span><span>agora</span></div><div class="gs-post-copy">${escapeHtml(text)}</div></div></div></div><div class="gs-post-foot"><button class="gs-action" type="button" data-gs-like>${icon("like")} <span data-gs-count>0</span></button><button class="gs-action" type="button" data-gs-comment-toggle aria-expanded="false">${icon("comment")} <span data-gs-count>0</span></button><button class="gs-action" type="button" data-gs-save>${icon("bookmark")} <span data-gs-save-label>Salvar</span></button></div><div class="gs-post-comments gs-hidden"><div class="gs-input-shell" data-gs-comment-form><input type="text" placeholder="Responder post" aria-label="Responder post" /><button class="gs-btn gs-btn-ghost" type="button" data-gs-comment-submit>Enviar</button></div><div></div></div>`;
        list.prepend(card);
        input.value = "";
      });
    });
  }

  function bindFeed(root) {
    qsa(root, "[data-gs-comment-form]").forEach((form) => {
      const input = qs(form, "input");
      const list = form.nextElementSibling;
      const action = qs(form, "[data-gs-comment-submit]");
      if (!input || !action || action.dataset.bound === "true") return;
      action.dataset.bound = "true";
      action.addEventListener("click", () => {
        const text = input.value.trim();
        if (!text || !list) return;
        const item = document.createElement("div");
        item.className = "gs-comment";
        item.innerHTML = `<div class="gs-comment-row"><span class="gs-avatar-xs" style="background:#ffb347">N</span><div class="gs-comment-content"><div class="gs-comment-bubble"><div class="gs-meta"><strong>Voce</strong><span>agora</span></div><p>${escapeHtml(text)}</p></div></div></div>`;
        list.appendChild(item);
        input.value = "";
      });
    });
  }

  function bindInteractions(root) {
    if (root.dataset.gsInteractiveBound === "true") return;
    root.dataset.gsInteractiveBound = "true";
    root.addEventListener("click", (event) => {
      const button = event.target.closest("button, [role='tab']");
      if (!button) return;
      if (button.hasAttribute("data-gs-menu-trigger")) {
        const panel = qs(button.closest("[data-gs-menu]"), "[data-gs-menu-panel]");
        const expanded = button.getAttribute("aria-expanded") === "true";
        closeMenus();
        button.setAttribute("aria-expanded", String(!expanded));
        if (panel) panel.hidden = expanded;
        return;
      }
      if (button.hasAttribute("data-gs-toast")) {
        showToast({
          title: button.getAttribute("data-gs-toast-title") || "Componente acionado",
          message: button.getAttribute("data-gs-toast-message") || "Exemplo executado com sucesso.",
          tone: button.getAttribute("data-gs-toast-tone") || "info",
        });
        return;
      }
      if (button.hasAttribute("data-gs-like")) {
        const active = button.dataset.state === "on";
        button.dataset.state = active ? "off" : "on";
        button.classList.toggle("is-danger", !active);
        nextCount(button, active ? -1 : 1);
        return;
      }
      if (button.hasAttribute("data-gs-save")) {
        const active = button.dataset.state === "on";
        button.dataset.state = active ? "off" : "on";
        const label = qs(button, "[data-gs-save-label]");
        if (label) label.textContent = active ? "Salvar" : "Salvo";
        return;
      }
      if (button.hasAttribute("data-gs-comment-toggle")) {
        const comments = button.closest(".gs-post")?.querySelector(".gs-post-comments");
        if (comments) {
          const expanded = !comments.classList.contains("gs-hidden");
          comments.classList.toggle("gs-hidden", expanded);
          button.setAttribute("aria-expanded", String(!expanded));
        }
        return;
      }
      if (button.hasAttribute("data-gs-tab")) {
        activateTab(button, false);
        return;
      }
      if (button.hasAttribute("data-gs-modal-open")) {
        openModal(button.getAttribute("data-gs-modal-open"));
        return;
      }
      if (button.hasAttribute("data-gs-modal-close")) {
        closeModal(button);
        return;
      }
      if (button.hasAttribute("data-gs-accordion-trigger")) {
        toggleAccordion(button);
      }
    });

    root.addEventListener("keydown", (event) => {
      const target = event.target;
      if (event.key === "Escape") {
        closeMenus();
        qsa(document, "[data-gs-modal]").filter((modal) => !modal.hidden).forEach(closeModal);
        return;
      }
      if (target.hasAttribute("data-gs-tab")) {
        const tabs = qsa(target.closest("[data-gs-tabs]"), "[data-gs-tab]");
        const current = tabs.indexOf(target);
        if (event.key === "ArrowRight") activateTab(tabs[(current + 1) % tabs.length], true);
        if (event.key === "ArrowLeft") activateTab(tabs[(current - 1 + tabs.length) % tabs.length], true);
      }
    });

    document.addEventListener("click", (event) => {
      if (!event.target.closest("[data-gs-menu]")) closeMenus();
      const modal = event.target.closest("[data-gs-modal]");
      if (modal && event.target === modal) closeModal(modal);
    });
  }

  function initSplash(root) {
    qsa(root, "[data-gs-splash]").forEach((splash) => {
      const delay = Number(splash.getAttribute("data-gs-splash-delay") || "2200");
      window.setTimeout(() => {
        splash.classList.add("is-hidden");
        window.setTimeout(() => splash.remove(), 500);
      }, delay);
    });
  }

  function init(root) {
    const scope = root || document.body;
    if (scope === document.body && scope.dataset.gsInit === "true") return;
    initSplash(scope);
    initTabs(scope);
    initAccordions(scope);
    initMenus(scope);
    initSelects(scope);
    initDatePickers(scope);
    initDataTables(scope);
    initTooltips(scope);
    bindComposer(scope);
    bindFeed(scope);
    bindInteractions(scope);
    validateAccessibility(scope);
    if (scope === document.body) scope.dataset.gsInit = "true";
  }

  window.GerminaStackUI = {
    init,
    showToast,
    request,
    closeMenus,
    openModal,
    closeModal,
    activateTab,
    toggleAccordion,
    announceToScreenReader,
    trapFocus,
    initSelects,
    initDatePickers,
    initDataTables,
    initTooltips,
    validateAccessibility,
  };

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", () => init(document.body));
  } else {
    init(document.body);
  }
})();

function Button({ label = "Botão", variant = "primary", type = "button" } = {}) {
  const element = document.createElement("button");
  element.className = `gs-btn gs-btn-${variant}`;
  element.type = type;
  element.textContent = label;
  return element;
}

function Card(content = "") {
  const element = document.createElement("section");
  element.className = "gs-card";
  element.textContent = content;
  return element;
}

const ui = window.GerminaStackUI;

export { Button, Card, ui };
