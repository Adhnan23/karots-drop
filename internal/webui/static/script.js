(function () {
  var BASE = "";
  var page = document.body.dataset.page || "";

  // ── Auth token from query string ──────────────────────────
  var AUTH_TOKEN = (function () {
    var p = new URLSearchParams(window.location.search);
    return p.get("token") || "";
  })();

  function $(sel) { return document.querySelector(sel); }
  function $$(sel) { return document.querySelectorAll(sel); }

  // ── Theme ─────────────────────────────────────────────────
  (function initTheme() {
    var saved = localStorage.getItem("theme");
    var pref = window.matchMedia("(prefers-color-scheme: light)").matches ? "light" : "dark";
    var theme = saved || pref;
    document.documentElement.setAttribute("data-theme", theme);

    var btn = $(".theme-toggle");
    if (btn) {
      btn.setAttribute("aria-label", "Toggle " + (theme === "dark" ? "light" : "dark") + " theme");
      btn.textContent = theme === "dark" ? "\u2600" : "\u{1F319}";
      btn.addEventListener("click", function () {
        var cur = document.documentElement.getAttribute("data-theme");
        var next = cur === "dark" ? "light" : "dark";
        document.documentElement.setAttribute("data-theme", next);
        localStorage.setItem("theme", next);
        btn.setAttribute("aria-label", "Toggle " + (next === "dark" ? "light" : "dark") + " theme");
        btn.textContent = next === "dark" ? "\u2600" : "\u{1F319}";
      });
    }
  })();

  // ── Shared helpers ────────────────────────────────────────
  function authFetch(url, opts) {
    opts = opts || {};
    opts.headers = opts.headers || {};
    if (AUTH_TOKEN) {
      opts.headers["X-Auth-Token"] = AUTH_TOKEN;
    }
    return fetch(url, opts);
  }

  function copyText(text, btn) {
    navigator.clipboard.writeText(text).then(function () {
      var orig = btn.textContent;
      btn.textContent = "Copied!";
      setTimeout(function () { btn.textContent = orig; }, 1500);
    })["catch"](function () {});
  }

  function handleResponse(resp) {
    if (!resp.ok) {
      return resp.json().then(function (d) { throw new Error(d.error || "Upload failed"); });
    }
    return resp.json();
  }

  function showError(id, msg) {
    var el = $("#" + id);
    el.textContent = "Error: " + msg;
    el.className = "result error";
    el.classList.remove("hidden");
    if (el.getAttribute("role") !== "alert") {
      el.setAttribute("role", "alert");
    }
  }

  // ── Show result (shared across pages) ─────────────────────
  function showResult(id, data) {
    var el = $("#" + id);
    if (!el) return;
    el.innerHTML = "";
    el.className = "result success";
    el.classList.remove("hidden");

    var mins = Math.round(data.ttl / 60);

    var p = document.createElement("p");
    p.innerHTML = "Code: <strong>" + data.code + "</strong> &mdash; expires in " + mins + " min";
    el.appendChild(p);

    var urlEl = document.createElement("p");
    urlEl.textContent = "URL: " + data.url;
    el.appendChild(urlEl);

    // Copy buttons row
    var copyRow = document.createElement("div");
    copyRow.className = "copy-row";

    var copyCodeBtn = document.createElement("button");
    copyCodeBtn.className = "copy-btn";
    copyCodeBtn.textContent = "Copy Code";
    copyCodeBtn.addEventListener("click", function () { copyText(data.code, copyCodeBtn); });
    copyRow.appendChild(copyCodeBtn);

    var copyUrlBtn = document.createElement("button");
    copyUrlBtn.className = "copy-btn";
    copyUrlBtn.textContent = "Copy URL";
    copyUrlBtn.addEventListener("click", function () { copyText(data.url, copyUrlBtn); });
    copyRow.appendChild(copyUrlBtn);

    el.appendChild(copyRow);

    if (data.filename) {
      var fn = document.createElement("p");
      fn.textContent = "File: " + data.filename;
      el.appendChild(fn);
    }

    if (data.encrypted) {
      var enc = document.createElement("p");
      enc.textContent = "Encrypted";
      enc.style.color = "var(--warning)";
      el.appendChild(enc);
    }

    // Show QR modal
    var qrImg = $("#qr-image");
    if (qrImg) {
      qrImg.src = BASE + "/api/qr/" + data.code;
      qrImg.alt = "QR code for code " + data.code;
    }
    var qrUrl = $("#qr-url");
    if (qrUrl) qrUrl.textContent = data.url;
    var qrCode = $("#qr-code");
    if (qrCode) qrCode.textContent = data.code;
    var qrModal = $("#qr-modal");
    if (qrModal) {
      qrModal.classList.remove("hidden");
      qrModal.setAttribute("aria-hidden", "false");
      var closeBtn = qrModal.querySelector(".close");
      if (closeBtn) closeBtn.focus();
    }
  }

  // ── QR modal ──────────────────────────────────────────────
  var qrModal = $("#qr-modal");
  if (qrModal) {
    // Copy buttons
    $$("#qr-modal .copy-btn").forEach(function (btn) {
      btn.addEventListener("click", function () {
        var text = btn.dataset.copy === "code"
          ? ($("#qr-code") || {}).textContent
          : ($("#qr-url") || {}).textContent;
        if (text) copyText(text, btn);
      });
    });

    // Close on X
    var closeBtn = qrModal.querySelector(".close");
    if (closeBtn) {
      closeBtn.addEventListener("click", function () {
        qrModal.classList.add("hidden");
        qrModal.setAttribute("aria-hidden", "true");
      });
    }

    // Close on overlay click
    qrModal.addEventListener("click", function (e) {
      if (e.target === this) {
        qrModal.classList.add("hidden");
        qrModal.setAttribute("aria-hidden", "true");
      }
    });

    // Close on Escape
    document.addEventListener("keydown", function (e) {
      if (e.key === "Escape" && !qrModal.classList.contains("hidden")) {
        qrModal.classList.add("hidden");
        qrModal.setAttribute("aria-hidden", "true");
      }
    });
  }

  // ── Page: Text upload (index) ─────────────────────────────
  if (page === "text") {
    $("#upload-text-btn").addEventListener("click", function () {
      var text = $("#text-input").value.trim();
      if (!text) return;
      uploadText(text);
    });

    // Ctrl/Cmd + Enter to submit
    $("#text-input").addEventListener("keydown", function (e) {
      if ((e.ctrlKey || e.metaKey) && e.key === "Enter") {
        e.preventDefault();
        $("#upload-text-btn").click();
      }
    });

    function uploadText(text) {
      var form = new FormData();
      form.append("text", text);
      authFetch(BASE + "/api/store", { method: "POST", body: form })
        .then(handleResponse)
        .then(function (data) { showResult("text-result", data); })
        .catch(function (err) { showError("text-result", err.message); });
    }
  }

  // ── Page: File upload ─────────────────────────────────────
  if (page === "file") {
    var dropZone = $("#drop-zone");
    var fileInput = $("#file-input");
    var selectedFile = null;

    dropZone.addEventListener("click", function () { fileInput.click(); });

    dropZone.addEventListener("dragover", function (e) {
      e.preventDefault();
      dropZone.classList.add("drag-over");
    });

    dropZone.addEventListener("dragleave", function () {
      dropZone.classList.remove("drag-over");
    });

    dropZone.addEventListener("drop", function (e) {
      e.preventDefault();
      dropZone.classList.remove("drag-over");
      if (e.dataTransfer.files.length > 0) {
        selectedFile = e.dataTransfer.files[0];
        $("#upload-file-btn").disabled = false;
        dropZone.querySelector("p").textContent = selectedFile.name;
      }
    });

    dropZone.addEventListener("keydown", function (e) {
      if (e.key === "Enter" || e.key === " ") {
        e.preventDefault();
        fileInput.click();
      }
    });

    fileInput.addEventListener("change", function () {
      if (fileInput.files.length > 0) {
        selectedFile = fileInput.files[0];
        $("#upload-file-btn").disabled = false;
        dropZone.querySelector("p").textContent = selectedFile.name;
      }
    });

    $("#upload-file-btn").addEventListener("click", function () {
      if (!selectedFile) return;
      var form = new FormData();
      form.append("file", selectedFile);
      authFetch(BASE + "/api/store", { method: "POST", body: form })
        .then(handleResponse)
        .then(function (data) { showResult("file-result", data); })
        .catch(function (err) { showError("file-result", err.message); })
        .finally(function () {
          selectedFile = null;
          $("#upload-file-btn").disabled = true;
          fileInput.value = "";
          dropZone.querySelector("p").textContent = "Drag & drop a file here, or click to select";
        });
    });
  }

  // ── Page: Retrieve ────────────────────────────────────────
  if (page === "retrieve") {
    $("#paste-btn").addEventListener("click", function () {
      navigator.clipboard.readText().then(function (txt) {
        var m = txt.match(/\d{6}/);
        if (m) {
          $("#code-input").value = m[0];
          $("#retrieve-btn").click();
        }
      }).catch(function () {});
    });

    $("#retrieve-btn").addEventListener("click", function () {
      var code = $("#code-input").value.trim();
      if (!code || code.length !== 6) return;
      retrieveCode(code);
    });

    $("#code-input").addEventListener("keydown", function (e) {
      if (e.key === "Enter") {
        e.preventDefault();
        $("#retrieve-btn").click();
      }
    });

    function retrieveCode(code) {
      var resultEl = $("#retrieve-result");
      var contentEl = $("#retrieve-content");
      resultEl.classList.add("hidden");
      contentEl.classList.add("hidden");

      authFetch(BASE + "/api/get/" + encodeURIComponent(code))
        .then(function (resp) {
          if (!resp.ok) {
            return resp.json().then(function (d) { throw new Error(d.error || "Not found"); });
          }
          var ct = resp.headers.get("Content-Type") || "";
          var cd = resp.headers.get("Content-Disposition") || "";
          if (ct.startsWith("text/") || ct.startsWith("application/json") || ct.startsWith("application/octet-stream")) {
            return resp.text().then(function (txt) {
              contentEl.textContent = txt;
              contentEl.classList.remove("hidden");
            });
          }
          return resp.blob().then(function (blob) {
            var url = URL.createObjectURL(blob);
            var match = cd.match(/filename="?(.+?)"?$/);
            var name = match ? match[1] : code;
            var a = document.createElement("a");
            a.href = url;
            a.download = name;
            a.textContent = "Download " + name;
            resultEl.innerHTML = "";
            resultEl.appendChild(a);
            resultEl.classList.remove("hidden");
            resultEl.className = "result success";
          });
        })
        .catch(function (err) {
          showError("retrieve-result", err.message);
        });
    }
  }

  // ── Active nav link ───────────────────────────────────────
  $$("nav a").forEach(function (a) {
    var href = a.getAttribute("href");
    if (href === "/" + page || (page === "text" && href === "/")) {
      a.setAttribute("aria-current", "page");
    }
  });
})();
