document.addEventListener("DOMContentLoaded", () => {
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

    let timerInterval;
    let currentTaskId;

    const openModal = (data) => {
        title.textContent = data.title || "Task";
        desc.textContent = data.description || "No description";
        priority.textContent = data.priority || "-";
        assignee.textContent = data.assignee || "-";
        due.textContent = data.due || "-";

        currentTaskId = data.id;

        // Live timer
        if (timerInterval) clearInterval(timerInterval);
        if (data.startedAt) {
            const startTime = new Date(data.startedAt);
            const updateTime = () => {
                const diff = Date.now() - startTime.getTime();
                const hours = Math.floor(diff / (1000 * 60 * 60));
                const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
                const seconds = Math.floor((diff % (1000 * 60)) / 1000);
                timeEl.textContent = `${hours.toString().padStart(2,'0')}:${minutes.toString().padStart(2,'0')}:${seconds.toString().padStart(2,'0')}`;
            };
            updateTime();
            timerInterval = setInterval(updateTime, 1000);
        } else {
            timeEl.textContent = "-";
        }

        // Show correct buttons
        if (data.startedAt) {
            startBtn.classList.add("hidden");
            uploadBtn.classList.remove("hidden");
            uploadContainer.classList.remove("hidden");
        } else {
            startBtn.classList.remove("hidden");
            uploadBtn.classList.add("hidden");
            uploadContainer.classList.add("hidden");
        }

        modal.classList.remove("hidden");
        modal.classList.add("flex");
    };

    const closeModal = () => {
        modal.classList.add("hidden");
        modal.classList.remove("flex");
        if (timerInterval) clearInterval(timerInterval);
        uploadInput.value = ""; // reset file input
        uploadStatus.textContent = "";
    };

    document.querySelectorAll(".task-card").forEach(card => {
        card.addEventListener("click", () => openModal(card.dataset));
    });

    document.getElementById("modalClose").onclick = closeModal;
    document.getElementById("modalClose2").onclick = closeModal;
    modal.addEventListener("click", (e) => { if(e.target===modal) closeModal(); });

    // Start task button
    startBtn.addEventListener("click", async () => {
        if (!currentTaskId) return;
        try {
            const res = await fetch(`/employee/tasks/start/${currentTaskId}`, { method: "POST" });
            if (!res.ok) throw new Error("Failed to start task");
            location.reload(); // reload to update board
        } catch(err) {
            alert(err.message);
        }
    });

    // Upload file button
    uploadBtn.addEventListener("click", async () => {
        if (!uploadInput.files[0]) {
            alert("Please select a file first");
            return;
        }

        const file = uploadInput.files[0];
        const formData = new FormData();
        formData.append("file", file);

        uploadStatus.textContent = "Uploading...";
        uploadStatus.className = "text-xs text-gray-500";

        try {
            const res = await fetch(`/employee/tasks/upload/${currentTaskId}`, { method: "POST", body: formData });
            const data = await res.json();
            if (!res.ok) throw new Error(data.error || "Upload failed");
            uploadStatus.textContent = "Uploaded successfully ✔";
            uploadStatus.className = "text-xs text-green-600";

            // Optionally append file name to modal
            let uploadedFiles = document.getElementById("uploadedFilesList");
            if(!uploadedFiles){
                uploadedFiles = document.createElement("div");
                uploadedFiles.id = "uploadedFilesList";
                uploadedFiles.className = "mt-2 text-xs text-gray-700";
                uploadContainer.appendChild(uploadedFiles);
            }
            uploadedFiles.innerHTML += `<div>• ${file.name}</div>`;
            uploadInput.value = ""; // reset input
        } catch(err) {
            uploadStatus.textContent = err.message;
            uploadStatus.className = "text-xs text-red-600";
        }
    });
});
