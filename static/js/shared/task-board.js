document.addEventListener("DOMContentLoaded", () => {
  const boardApp = document.getElementById("taskBoardApp");
  if (!boardApp) return;

  const startBase = boardApp.dataset.startBase || "";
  const pauseBase = boardApp.dataset.pauseBase || "";
  const resumeBase = boardApp.dataset.resumeBase || "";
  const uploadBase = boardApp.dataset.uploadBase || "";
  const filesBase = boardApp.dataset.filesBase || "";
  const commentsBase = boardApp.dataset.commentsBase || "";
  const reviewBase = boardApp.dataset.reviewBase || "";
  const canReview = boardApp.dataset.canReview === "true";
  const canUpload = boardApp.dataset.canUpload !== "false" && !!uploadBase;

  const modal = document.getElementById("taskModal");
  const title = document.getElementById("modalTitle");
  const desc = document.getElementById("modalDescription");
  const project = document.getElementById("modalProject");
  const priority = document.getElementById("modalPriority");
  const assignee = document.getElementById("modalAssignee");
  const due = document.getElementById("modalDue");
  const timeEl = document.getElementById("modalTime");
  const timerStatusEl = document.getElementById("modalTimerStatus");
  const uploadContainer = document.getElementById("taskUploadContainer");
  const startBtn = document.getElementById("taskStartBtn");
  const pauseBtn = document.getElementById("taskPauseBtn");
  const resumeBtn = document.getElementById("taskResumeBtn");
  const uploadBtn = document.getElementById("taskUploadBtn");
  const uploadInput = document.getElementById("taskUpload");
  const uploadStatus = document.getElementById("uploadStatus");
  const uploadedFilesList = document.getElementById("uploadedFilesList");
  const commentsList = document.getElementById("taskCommentsList");
  const reviewContainer = document.getElementById("taskReviewContainer");
  const reviewMessage = document.getElementById("taskReviewMessage");
  const approveBtn = document.getElementById("taskApproveBtn");
  const rejectBtn = document.getElementById("taskRejectBtn");

  if (!modal) return;

  let timerInterval;
  let currentTaskId = "";
  let busy = false;
  let modalFeedback = document.getElementById("taskModalFeedback");

  if (!modalFeedback) {
    modalFeedback = document.createElement("p");
    modalFeedback.id = "taskModalFeedback";
    modalFeedback.className = "text-xs hidden";
    const body = modal.querySelector(".space-y-4");
    if (body) body.insertBefore(modalFeedback, body.firstChild);
  }

  const setFeedback = (message, type) => {
    if (!modalFeedback) return;
    if (!message) {
      modalFeedback.textContent = "";
      modalFeedback.className = "text-xs hidden";
      return;
    }

    const colorClass =
      type === "success"
        ? "text-green-600"
        : type === "error"
          ? "text-red-600"
          : type === "warning"
            ? "text-amber-600"
            : "text-gray-600";

    modalFeedback.textContent = message;
    modalFeedback.className = `text-xs ${colorClass}`;
  };

  const formatDateTime = (value) => {
    if (!value) return "";
    const parsed = new Date(value);
    if (Number.isNaN(parsed.getTime())) return "";
    return parsed.toLocaleString();
  };

  const readResponsePayload = async (res, fallbackMessage) => {
    const contentType = (res.headers.get("content-type") || "").toLowerCase();

    if (contentType.includes("application/json")) {
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || fallbackMessage);
      return data;
    }

    const text = (await res.text()).trim();
    const message = text
      ? text.replace(/<[^>]*>/g, " ").replace(/\s+/g, " ").trim()
      : fallbackMessage;

    throw new Error(message || fallbackMessage);
  };

  const renderUploadedFiles = (files) => {
    if (!uploadedFilesList) return;

    uploadedFilesList.innerHTML = "";
    if (!files || files.length === 0) {
      uploadedFilesList.innerHTML = '<p class="text-xs text-gray-500">No files uploaded yet.</p>';
      return;
    }

    files.forEach((file) => {
      const item = document.createElement("div");
      item.className = "rounded-lg border border-gray-200 bg-gray-50 px-3 py-2";

      const top = document.createElement("div");
      top.className = "flex flex-wrap items-center gap-2";

      const version = document.createElement("span");
      version.className = "inline-flex items-center rounded-full bg-slate-200 px-2 py-0.5 text-[11px] font-semibold text-slate-700";
      version.textContent = `Version ${file.version || 1}`;
      top.appendChild(version);

      if (file.is_latest) {
        const latest = document.createElement("span");
        latest.className = "inline-flex items-center rounded-full bg-green-100 px-2 py-0.5 text-[11px] font-semibold text-green-700";
        latest.textContent = "Latest";
        top.appendChild(latest);
      }

      item.appendChild(top);

      const meta = document.createElement("div");
      meta.className = "mt-2 flex items-center justify-between gap-3";

      const label = file.file_name || file.name || "Unnamed file";
      if (file.download_url) {
        const link = document.createElement("a");
        link.href = file.download_url;
        link.target = "_blank";
        link.rel = "noopener noreferrer";
        link.className = "truncate text-blue-600 hover:underline";
        link.textContent = label;
        meta.appendChild(link);
      } else {
        const text = document.createElement("span");
        text.className = "truncate text-gray-700";
        text.textContent = label;
        meta.appendChild(text);
      }

      const uploadedAt = document.createElement("span");
      uploadedAt.className = "shrink-0 text-[11px] text-gray-500";
      uploadedAt.textContent = formatDateTime(file.uploaded_at);
      meta.appendChild(uploadedAt);

      item.appendChild(meta);
      uploadedFilesList.appendChild(item);
    });
  };

  const loadUploadedFiles = async (taskId) => {
    if (!uploadedFilesList || !filesBase || !taskId) return;

    uploadedFilesList.innerHTML = '<p class="text-xs text-gray-500">Loading files...</p>';

    try {
      const res = await fetch(`${filesBase}${taskId}`);
      const data = await readResponsePayload(res, "Failed to load files");
      renderUploadedFiles(data.files || []);
    } catch (err) {
      uploadedFilesList.innerHTML = `<p class="text-xs text-red-600">${err.message || "Failed to load files"}</p>`;
    }
  };

  const renderComments = (comments) => {
    if (!commentsList) return;

    commentsList.innerHTML = "";
    if (!comments || comments.length === 0) {
      commentsList.textContent = "No review comments yet.";
      return;
    }

    comments.forEach((comment) => {
      const item = document.createElement("div");
      item.className = "rounded border border-gray-200 bg-gray-50 p-2";
      const by = comment.creator_name || "Reviewer";
      const action = comment.action || "commented";
      item.textContent = `${by} (${action}): ${comment.message || ""}`;
      commentsList.appendChild(item);
    });
  };

  const loadComments = async (taskId) => {
    if (!commentsList || !commentsBase || !taskId) return;

    commentsList.textContent = "Loading comments...";

    try {
      const res = await fetch(`${commentsBase}${taskId}`);
      const data = await readResponsePayload(res, "Failed to load comments");
      renderComments(data.comments || []);
    } catch (err) {
      commentsList.textContent = err.message || "Failed to load comments";
    }
  };

  const formatDuration = (seconds) => {
    const safe = Math.max(0, Number(seconds) || 0);
    const hours = Math.floor(safe / 3600);
    const minutes = Math.floor((safe % 3600) / 60);
    const secs = Math.floor(safe % 60);
    return `${hours.toString().padStart(2, "0")}:${minutes.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`;
  };

  const updateTotalTimer = (baseSeconds, startedAt) => {
    if (timerInterval) clearInterval(timerInterval);
    if (!timeEl) return;

    const base = Math.max(0, Number(baseSeconds) || 0);
    if (!startedAt) {
      timeEl.textContent = formatDuration(base);
      return;
    }

    const startTime = new Date(startedAt);
    const render = () => {
      const elapsed = Math.max(0, Math.floor((Date.now() - startTime.getTime()) / 1000));
      timeEl.textContent = formatDuration(base + elapsed);
    };

    render();
    timerInterval = setInterval(render, 1000);
  };

  const toggleButton = (button, visible) => {
    if (!button) return;
    if (visible) button.classList.remove("hidden");
    else button.classList.add("hidden");
  };

  const getTaskCard = (taskId) => document.querySelector(`.task-card[data-id="${taskId}"]`);

  const syncTaskCardState = (taskId, updates = {}) => {
    const card = getTaskCard(taskId);
    if (!card) return null;

    Object.entries(updates).forEach(([key, value]) => {
      if (value === undefined || value === null) {
        delete card.dataset[key];
        return;
      }

      card.dataset[key] = String(value);
    });

    const nextStatus = updates.status || card.dataset.status;
    if (nextStatus) {
      const nextColumn = document.querySelector(`[data-status="${nextStatus}"]`);
      if (nextColumn && card.parentElement !== nextColumn) {
        nextColumn.prepend(card);
      }
    }

    return card;
  };

  const resolveTimerState = (status, isPaused, startedAt, explicitState) => {
    if (status === "in_progress" && isPaused) return "paused";
    if (status === "in_progress" && startedAt) return "running";
    if (status === "for_review") return "for_review";
    if (status === "done") return "completed";

    const normalizedExplicitState = (explicitState || "").trim().toLowerCase();
    if (normalizedExplicitState === "paused" || normalizedExplicitState === "running") {
      return normalizedExplicitState;
    }

    return "not_started";
  };

  const renderTimerStatus = (timerState) => {
    if (!timerStatusEl) return;

    let text = "Not started";
    let className = "font-medium text-gray-700";

    if (timerState === "paused") {
      text = "Paused";
      className = "font-medium text-amber-600";
    } else if (timerState === "running") {
      text = "Running";
      className = "font-medium text-blue-600";
    } else if (timerState === "for_review") {
      text = "Submitted for review";
      className = "font-medium text-violet-600";
    } else if (timerState === "completed") {
      text = "Completed";
      className = "font-medium text-emerald-600";
    }

    timerStatusEl.textContent = text;
    timerStatusEl.className = className;
  };

  const openModal = (data) => {
    currentTaskId = data.id || "";
    const status = (data.status || "").toLowerCase();
    const durationSeconds = Number(data.durationSeconds || 0);
    const isPaused = data.isPaused === "true";
    const timerState = resolveTimerState(status, isPaused, data.startedAt, data.timerState);
    const isNotStarted = timerState === "not_started";
    const isRunning = timerState === "running";
    const isPausedState = timerState === "paused";

    if (title) title.textContent = data.title || "Task";
    if (desc) desc.textContent = data.description || "No description";
    if (project) project.textContent = data.project || "-";
    if (priority) priority.textContent = data.priority || "-";
    if (assignee) assignee.textContent = data.assignee || "-";
    if (due) due.textContent = data.due || "-";
    renderTimerStatus(timerState);
    setFeedback(isPaused ? "Timer is currently paused for this task." : "");

    updateTotalTimer(durationSeconds, data.startedAt);

    const canReviewNow = canReview && status === "for_review";

    if (canReview) {
      toggleButton(startBtn, false);
      toggleButton(pauseBtn, false);
      toggleButton(resumeBtn, false);
      toggleButton(uploadBtn, false);
      if (uploadContainer) uploadContainer.classList.add("hidden");
      if (reviewContainer) reviewContainer.classList.toggle("hidden", !canReviewNow);
      toggleButton(approveBtn, canReviewNow);
      toggleButton(rejectBtn, canReviewNow);
    } else {
      if (reviewContainer) reviewContainer.classList.add("hidden");
      toggleButton(approveBtn, false);
      toggleButton(rejectBtn, false);

      toggleButton(startBtn, isNotStarted && !!startBase);
      toggleButton(pauseBtn, isRunning && !!pauseBase);
      toggleButton(resumeBtn, isPausedState && !!resumeBase);
      toggleButton(uploadBtn, isRunning && canUpload);

      if (uploadContainer) {
        if (isRunning && canUpload) {
          uploadContainer.classList.remove("hidden");
        } else {
          uploadContainer.classList.add("hidden");
        }
      }
    }

    if (uploadedFilesList) uploadedFilesList.innerHTML = "";
    loadUploadedFiles(currentTaskId);
    loadComments(currentTaskId);

    if (reviewMessage) reviewMessage.value = "";
    if (uploadInput) uploadInput.value = "";
    if (uploadStatus) {
      uploadStatus.textContent = "";
      uploadStatus.className = "text-xs mt-2";
    }

    modal.classList.remove("hidden");
    modal.classList.add("flex");
  };

  const closeModal = () => {
    modal.classList.add("hidden");
    modal.classList.remove("flex");
    if (timerInterval) clearInterval(timerInterval);
    currentTaskId = "";
    busy = false;
    setFeedback("");
  };

  document.querySelectorAll(".task-card").forEach((card) => {
    card.addEventListener("click", () => openModal(card.dataset));
  });

  const close1 = document.getElementById("modalClose");
  const close2 = document.getElementById("modalClose2");
  if (close1) close1.onclick = closeModal;
  if (close2) close2.onclick = closeModal;
  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape" && !modal.classList.contains("hidden")) closeModal();
  });
  modal.addEventListener("click", (e) => {
    if (e.target === modal) closeModal();
  });

  const postTaskAction = async (url, loadingMessage, errorMessage, onSuccess) => {
    if (!currentTaskId || !url || busy) return;
    busy = true;
    setFeedback(loadingMessage, "info");

    if (startBtn) startBtn.disabled = true;
    if (pauseBtn) pauseBtn.disabled = true;
    if (resumeBtn) resumeBtn.disabled = true;

    try {
      const res = await fetch(url, { method: "POST" });
      if (!res.ok) throw new Error(errorMessage);
      if (typeof onSuccess === "function") onSuccess();
      busy = false;
    } catch (err) {
      setFeedback(err.message || errorMessage, "error");
      busy = false;
      if (startBtn) startBtn.disabled = false;
      if (pauseBtn) pauseBtn.disabled = false;
      if (resumeBtn) resumeBtn.disabled = false;
    }
  };

  if (startBtn) {
    startBtn.addEventListener("click", () => postTaskAction(
      `${startBase}${currentTaskId}`,
      "Starting task...",
      "Failed to start task",
      () => {
        const startedAt = new Date().toISOString();
        const card = syncTaskCardState(currentTaskId, {
          status: "in_progress",
          startedAt,
          isPaused: "false",
          timerState: "running",
        });

        if (card) openModal(card.dataset);
        setFeedback("Task started. Timer is running.", "success");
        if (startBtn) startBtn.disabled = false;
        if (pauseBtn) pauseBtn.disabled = false;
        if (resumeBtn) resumeBtn.disabled = false;
      },
    ));
  }

  if (pauseBtn) {
    pauseBtn.addEventListener("click", () => postTaskAction(
      `${pauseBase}${currentTaskId}`,
      "Pausing timer...",
      "Failed to pause task",
      () => {
        const card = syncTaskCardState(currentTaskId, {
          status: "in_progress",
          startedAt: "",
          isPaused: "true",
          timerState: "paused",
        });

        if (card) openModal(card.dataset);
        setFeedback("Timer paused for this task.", "success");
        if (startBtn) startBtn.disabled = false;
        if (pauseBtn) pauseBtn.disabled = false;
        if (resumeBtn) resumeBtn.disabled = false;
      },
    ));
  }

  if (resumeBtn) {
    resumeBtn.addEventListener("click", () => postTaskAction(
      `${resumeBase}${currentTaskId}`,
      "Resuming timer...",
      "Failed to resume task",
      () => {
        const startedAt = new Date().toISOString();
        const card = syncTaskCardState(currentTaskId, {
          status: "in_progress",
          startedAt,
          isPaused: "false",
          timerState: "running",
        });

        if (card) openModal(card.dataset);
        setFeedback("Timer resumed.", "success");
        if (startBtn) startBtn.disabled = false;
        if (pauseBtn) pauseBtn.disabled = false;
        if (resumeBtn) resumeBtn.disabled = false;
      },
    ));
  }

  if (uploadBtn) {
    uploadBtn.addEventListener("click", async () => {
      if (!currentTaskId || !uploadInput) return;
      if (busy) return;
      if (!uploadInput.files[0]) {
        setFeedback("Please select a file first.", "error");
        return;
      }

      const file = uploadInput.files[0];
      const formData = new FormData();
      formData.append("file", file);

      busy = true;
      uploadBtn.disabled = true;
      if (pauseBtn) pauseBtn.disabled = true;
      if (uploadStatus) {
        uploadStatus.textContent = "Uploading...";
        uploadStatus.className = "text-xs mt-2 text-gray-500";
      }

      try {
        const res = await fetch(`${uploadBase}${currentTaskId}`, {
          method: "POST",
          body: formData,
        });
        const data = await readResponsePayload(res, "Upload failed");

        if (uploadStatus) {
          uploadStatus.textContent = "Uploaded successfully";
          uploadStatus.className = "text-xs mt-2 text-green-600";
        }
        setFeedback("File uploaded as the latest submission. Task moved to review queue.", "success");

        if (timerInterval) clearInterval(timerInterval);
        if (timeEl) timeEl.textContent = formatDuration(data.file?.duration_seconds || 0);

        renderUploadedFiles([{
          file_name: data.file_name || file.name,
          version: data.file?.version || data.version || 1,
          is_latest: true,
          uploaded_at: new Date().toISOString(),
        }]);

        uploadInput.value = "";

        setTimeout(() => {
          location.reload();
        }, 500);
      } catch (err) {
        if (uploadStatus) {
          uploadStatus.textContent = err.message || "Upload failed";
          uploadStatus.className = "text-xs mt-2 text-red-600";
        }
        setFeedback(err.message || "Upload failed", "error");
        busy = false;
        uploadBtn.disabled = false;
        if (pauseBtn) pauseBtn.disabled = false;
      }
    });
  }

  const submitReview = async (action) => {
    if (!currentTaskId || !reviewBase || busy) return;
    busy = true;

    const message = reviewMessage ? reviewMessage.value : "";
    setFeedback(`Submitting ${action} review...`, "info");
    if (approveBtn) approveBtn.disabled = true;
    if (rejectBtn) rejectBtn.disabled = true;

    try {
      const res = await fetch(`${reviewBase}${currentTaskId}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action, message }),
      });
      await readResponsePayload(res, "Failed to submit review");
      location.reload();
    } catch (err) {
      setFeedback(err.message || "Failed to submit review", "error");
      busy = false;
      if (approveBtn) approveBtn.disabled = false;
      if (rejectBtn) rejectBtn.disabled = false;
    }
  };

  if (approveBtn) approveBtn.addEventListener("click", () => submitReview("approved"));
  if (rejectBtn) rejectBtn.addEventListener("click", () => submitReview("rejected"));
});
