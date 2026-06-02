(() => {
  "use strict";

  const writeClipboard = (text) => {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      return navigator.clipboard.writeText(text);
    }
    const field = document.createElement("textarea");
    field.value = text;
    field.setAttribute("readonly", "");
    field.style.position = "fixed";
    field.style.left = "-9999px";
    document.body.appendChild(field);
    field.select();
    const copied = document.execCommand("copy");
    document.body.removeChild(field);
    return copied ? Promise.resolve() : Promise.reject(new Error("copy failed"));
  };

  document.addEventListener("click", (event) => {
    if (!(event.target instanceof Element)) return;

    const copyButton = event.target.closest("button[data-copy-target]");
    if (copyButton) {
      const copyTarget = document.getElementById(copyButton.getAttribute("data-copy-target"));
      if (!copyTarget) return;
      const original = copyButton.textContent;
      writeClipboard(copyTarget.innerText).then(() => {
        copyButton.textContent = "Copied";
        window.setTimeout(() => { copyButton.textContent = original; }, 1400);
      }).catch(() => {
        copyButton.textContent = "Copy failed";
        window.setTimeout(() => { copyButton.textContent = original; }, 1400);
      });
      return;
    }

    const toggleButton = event.target.closest("[data-toggle-target]");
    if (!toggleButton) return;
    const toggleTarget = document.getElementById(toggleButton.getAttribute("data-toggle-target"));
    if (toggleTarget) toggleTarget.classList.toggle("hidden");
  });

  document.addEventListener("submit", (event) => {
    const form = event.target;
    if (!(form instanceof HTMLFormElement)) return;
    const message = form.getAttribute("data-confirm");
    if (message && !window.confirm(message)) event.preventDefault();
  });
})();
