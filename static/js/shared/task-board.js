document.addEventListener("DOMContentLoaded", () => {
  const boardApp = document.getElementById("taskBoardApp");
  if (!boardApp) return;

  const startBase = boardApp.dataset.startBase || "";
  const uploadBase = boardApp.dataset.uploadBase || "";
  const filesBase = boardApp.dataset.filesBase || "";
  const commentsBase = boardApp.dataset.commentsBase || "";
  const reviewBase = boardApp.dataset.reviewBase || "";
  const canReview = boardApp.dataset.canReview === "true";
  const canUpload = boardApp.dataset.canUpload !== "false" && !!uploadBase;

  const modal = document.getElementById("taskModal");
  const title = document.getElementById("modalTitle");
  const desc = document.getElementById("modalDescription");
  const priority = document.getElementById("modalPriority");
  const assignee = document.getElementById("modalAssignee");
  const due = document.getElementById("modalDue");
  const timeEl = document.getElementById("modalTime");
  const uploadContainer = document.getElementById("taskUploadContainer");
  const startBtn = document.getElementById("taskStartBtn");
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
    if (body) {
      body.insertBefore(modalFeedback, body.firstChild);
    }
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
          : "text-gray-600";
    modalFeedback.textContent = message;
    modalFeedback.className = `text-xs ${colorClass}`;
  };

  const renderUploadedFiles = (files) => {
    if (!uploadedFilesList) return;

    uploadedFilesList.innerHTML = "";
    if (!files || files.length === 0) {
      uploadedFilesList.textContent = "No files uploaded yet.";
      return;
    }

    files.forEach((file) => {
      const item = document.createElement("div");
      const label = file.file_name || file.name || "Unnamed file";
      if (file.download_url) {
        const link = document.createElement("a");
        link.href = file.download_url;
        link.target = "_blank";
        link.rel = "noopener noreferrer";
        link.className = "text-blue-600 hover:underline";
        link.textContent = `- ${label}`;
        item.appendChild(link);
      } else {
        item.textContent = `- ${label}`;
      }
      uploadedFilesList.appendChild(item);
    });
  };

  const loadUploadedFiles = async (taskId) => {
    if (!uploadedFilesList || !filesBase || !taskId) return;

    uploadedFilesList.textContent = "Loading files...";

    try {
      const res = await fetch(`${filesBase}${taskId}`);
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || "Failed to load files");
      renderUploadedFiles(data.files || []);
    } catch (err) {
      uploadedFilesList.textContent = err.message || "Failed to load files";
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
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || "Failed to load comments");
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
      timeEl.textContent = base > 0 ? formatDuration(base) : "-";
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

  const openModal = (data) => {
    currentTaskId = data.id || "";
    const status = (data.status || "").toLowerCase();
    const durationSeconds = Number(data.durationSeconds || 0);

    if (title) title.textContent = data.title || "Task";
    if (desc) desc.textContent = data.description || "No description";
    if (priority) priority.textContent = data.priority || "-";
    if (assignee) assignee.textContent = data.assignee || "-";
    if (due) due.textContent = data.due || "-";
    setFeedback("");

    updateTotalTimer(durationSeconds, data.startedAt);

    const isTodo = status === "todo";
    const isInProgress = status === "in_progress";
    const canReviewNow = canReview && status === "for_review";

    if (canReview) {
      if (startBtn) startBtn.classList.add("hidden");
      if (uploadBtn) uploadBtn.classList.add("hidden");
      if (uploadContainer) uploadContainer.classList.add("hidden");
      if (reviewContainer) {
        if (canReviewNow) {
          reviewContainer.classList.remove("hidden");
        } else {
          reviewContainer.classList.add("hidden");
        }
      }
      if (approveBtn) {
        if (canReviewNow) {
          approveBtn.classList.remove("hidden");
        } else {
          approveBtn.classList.add("hidden");
        }
      }
      if (rejectBtn) {
        if (canReviewNow) {
          rejectBtn.classList.remove("hidden");
        } else {
          rejectBtn.classList.add("hidden");
        }
      }
    } else {
      if (reviewContainer) reviewContainer.classList.add("hidden");
      if (approveBtn) approveBtn.classList.add("hidden");
      if (rejectBtn) rejectBtn.classList.add("hidden");

      if (startBtn) {
        if (isTodo) {
          startBtn.classList.remove("hidden");
        } else {
          startBtn.classList.add("hidden");
        }
      }
      if (uploadBtn) {
        if (isInProgress && canUpload) {
          uploadBtn.classList.remove("hidden");
        } else {
          uploadBtn.classList.add("hidden");
        }
      }
      if (uploadContainer) {
        if (isInProgress && canUpload) {
          uploadContainer.classList.remove("hidden");
        } else {
          uploadContainer.classList.add("hidden");
        }
      }
    }

    if (uploadedFilesList) {
      uploadedFilesList.innerHTML = "";
    }
    loadUploadedFiles(currentTaskId);
    loadComments(currentTaskId);

    if (reviewMessage) {
      reviewMessage.value = "";
    }
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
    if (e.key === "Escape" && !modal.classList.contains("hidden")) {
      closeModal();
    }
  });

  modal.addEventListener("click", (e) => {
    if (e.target === modal) closeModal();
  });

  if (startBtn) startBtn.addEventListener("click", async () => {
    if (!currentTaskId) return;
    if (busy) return;
    busy = true;
    setFeedback("Starting task...", "info");
    startBtn.disabled = true;
    try {
      const res = await fetch(`${startBase}${currentTaskId}`, { method: "POST" });
      if (!res.ok) throw new Error("Failed to start task");
      location.reload();
    } catch (err) {
      setFeedback(err.message || "Failed to start task", "error");
    } finally {
      busy = false;
      startBtn.disabled = false;
    }
  });

  if (uploadBtn) uploadBtn.addEventListener("click", async () => {
    if (!currentTaskId) return;
    if (!uploadInput) return;
    if (busy) return;
    if (!uploadInput.files[0]) {
      setFeedback("Please select a file first.", "error");
      return;
    }

    const formData = new FormData();
    const file = uploadInput.files[0];
    formData.append("file", file);

    busy = true;
    uploadBtn.disabled = true;
    if (uploadStatus) {
      uploadStatus.textContent = "Uploading...";
      uploadStatus.className = "text-xs mt-2 text-gray-500";
    }

    try {
      const res = await fetch(`${uploadBase}${currentTaskId}`, {
        method: "POST",
        body: formData,
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || "Upload failed");

      if (uploadStatus) {
        uploadStatus.textContent = "Uploaded successfully";
        uploadStatus.className = "text-xs mt-2 text-green-600";
      }
      setFeedback("File uploaded. Task moved to review queue.", "success");

      if (timerInterval) clearInterval(timerInterval);
      timeEl.textContent = "-";

      renderUploadedFiles([{ file_name: data.file_name || file.name }]);

      uploadInput.value = "";

      // Upload marks the task as for_review and clears started_at in backend.
      // Reload board so the card moves to the correct status column.
      setTimeout(() => {
        location.reload();
      }, 500);
    } catch (err) {
      if (uploadStatus) {
        uploadStatus.textContent = err.message || "Upload failed";
        uploadStatus.className = "text-xs mt-2 text-red-600";
      }
      setFeedback(err.message || "Upload failed", "error");
    } finally {
      busy = false;
      uploadBtn.disabled = false;
    }
  });

  const submitReview = async (action) => {
    if (!currentTaskId || !reviewBase) return;
    if (busy) return;
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
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || "Failed to submit review");
      location.reload();
    } catch (err) {
      setFeedback(err.message || "Failed to submit review", "error");
    } finally {
      busy = false;
      if (approveBtn) approveBtn.disabled = false;
      if (rejectBtn) rejectBtn.disabled = false;
    }
  };

  if (approveBtn) {
    approveBtn.addEventListener("click", () => submitReview("approved"));
  }

  if (rejectBtn) {
    rejectBtn.addEventListener("click", () => submitReview("rejected"));
  }
});
