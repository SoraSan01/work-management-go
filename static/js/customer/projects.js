(() => {
  'use strict';

  const table = document.getElementById('projectsTable');
  if (!table) return;

  const tbody = table.querySelector('tbody');
  const rows = Array.from(tbody.querySelectorAll('tr'));
  const searchInput = document.getElementById('projectsSearch');
  const info = document.getElementById('projectsTableInfo');
  const pagination = document.getElementById('projectsTablePagination');

  let currentPage = 1;
  const pageSize = 10;
  let filteredRows = rows.slice();

  const normalize = v => (v || '').toLowerCase().trim();

  function renderRows() {
    tbody.innerHTML = '';

    if (!filteredRows.length) {
      tbody.innerHTML = `
        <tr>
          <td colspan="4" class="px-4 py-10 text-center text-sm text-gray-500">
            No projects match your search.
          </td>
        </tr>
      `;
      return;
    }

    const start = (currentPage - 1) * pageSize;
    const pageRows = filteredRows.slice(start, start + pageSize);

    pageRows.forEach(r => tbody.appendChild(r));
  }

  function renderInfo() {
    const total = filteredRows.length;
    const start = (currentPage - 1) * pageSize + 1;
    const end = Math.min(currentPage * pageSize, total);
    info.textContent = `Showing ${start} to ${end} of ${total} projects`;
  }

  function renderPagination() {
    pagination.innerHTML = '';
    const totalPages = Math.ceil(filteredRows.length / pageSize) || 1;

    const createBtn = (label, page) => {
      const b = document.createElement('button');
      b.textContent = label;
      b.className = 'rounded-lg border px-3 py-2 text-xs hover:bg-gray-50';
      b.onclick = () => {
        currentPage = page;
        render();
      };
      return b;
    };

    for (let i = 1; i <= totalPages; i++) {
      pagination.appendChild(createBtn(i, i));
    }
  }

  function render() {
    renderRows();
    renderInfo();
    renderPagination();
  }

  if (searchInput) {
    searchInput.addEventListener('input', () => {
      searchInput.value = searchInput.value.replace(/[^A-Za-z0-9\s\-_. ,]/g, '');
      const q = normalize(searchInput.value);

      filteredRows = rows.filter(r =>
        normalize(r.dataset.name).includes(q) ||
        normalize(r.dataset.description).includes(q) ||
        normalize(r.dataset.status).includes(q)
      );

      currentPage = 1;
      render();
    });
  }

  render();
})();


(() => {
  'use strict';

  const modal = document.getElementById('projectModal');
  if (!modal) return;

  const overlay = document.getElementById('projectModalOverlay');
  const openBtn = document.getElementById('openAddProjectModal');
  const closeBtns = [
    document.getElementById('closeProjectModal'),
    document.getElementById('cancelProjectModal')
  ];

  const form = document.getElementById('projectModalForm');
  const titleInput = document.getElementById('projectName');
  const descriptionInput = document.getElementById('projectDescription');
  const deadlineInput = document.getElementById('projectDeadline');

  const letterAndSpaceRegex = /^[A-Za-z][A-Za-z\s]*$/;

  const todayLocalDate = () => {
    const now = new Date();
    now.setMinutes(now.getMinutes() - now.getTimezoneOffset());
    return now.toISOString().split('T')[0];
  };

  if (deadlineInput) {
    deadlineInput.min = todayLocalDate();
  }

  if (titleInput) {
    titleInput.addEventListener('input', () => {
      titleInput.value = titleInput.value.replace(/[^A-Za-z\s]/g, '');
      titleInput.setCustomValidity('');
    });
  }

  if (descriptionInput) {
    descriptionInput.addEventListener('input', () => {
      if (descriptionInput.value.length > 500) {
        descriptionInput.value = descriptionInput.value.slice(0, 500);
      }
      descriptionInput.setCustomValidity('');
    });
  }

  if (deadlineInput) {
    deadlineInput.addEventListener('input', () => deadlineInput.setCustomValidity(''));
  }

  form?.addEventListener('submit', (e) => {
    if (!titleInput || !deadlineInput) return;

    const titleValue = titleInput.value.trim();
    titleInput.value = titleValue.replace(/\s{2,}/g, ' ');

    titleInput.setCustomValidity('');
    descriptionInput?.setCustomValidity('');
    deadlineInput.setCustomValidity('');

    if (titleValue.length < 2 || titleValue.length > 100) {
      titleInput.setCustomValidity('Project name must be between 2 and 100 characters.');
    } else if (!letterAndSpaceRegex.test(titleValue)) {
      titleInput.setCustomValidity('Project name must contain letters and spaces only.');
    }

    if (descriptionInput && descriptionInput.value.length > 500) {
      descriptionInput.setCustomValidity('Description must be 500 characters or fewer.');
    }

    if (deadlineInput.value && deadlineInput.value < todayLocalDate()) {
      deadlineInput.setCustomValidity('Deadline cannot be earlier than today.');
    }

    if (!form.checkValidity()) {
      e.preventDefault();
      form.reportValidity();
    }
  });

  const openModal = () => {
    modal.classList.remove('hidden');
    document.body.style.overflow = 'hidden';
  };

  const closeModal = () => {
    modal.classList.add('hidden');
    document.body.style.overflow = '';
    form.reset();
  };

  openBtn?.addEventListener('click', openModal);

  closeBtns.forEach(btn =>
    btn?.addEventListener('click', closeModal)
  );

  overlay?.addEventListener('click', closeModal);

  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') closeModal();
  });
})();

(() => {
  'use strict';

  const modal = document.getElementById('viewProjectModal');
  if (!modal) return;

  const overlay = document.getElementById('viewProjectOverlay');
  const closeBtns = [
    document.getElementById('closeViewProjectModal'),
    document.getElementById('closeViewProjectModalBtn')
  ];

  const viewName = document.getElementById('viewName');
  const viewDesc = document.getElementById('viewDescription');
  const viewDeadline = document.getElementById('viewDeadline');
  const viewStatus = document.getElementById('viewStatus');
  const viewProjectFiles = document.getElementById('viewProjectFiles');

  const openModal = () => {
    modal.classList.remove('hidden');
    document.body.style.overflow = 'hidden';
  };

  const closeModal = () => {
    modal.classList.add('hidden');
    document.body.style.overflow = '';
  };

  const renderProjectFiles = (files) => {
    if (!viewProjectFiles) return;

    viewProjectFiles.innerHTML = '';
    if (!files || files.length === 0) {
      viewProjectFiles.innerHTML = '<p class="text-xs text-gray-500">No approved or finished files yet.</p>';
      return;
    }

    files.forEach((file) => {
      const row = document.createElement('div');
      row.className = 'flex items-center justify-between gap-3 rounded-lg border border-gray-200 px-3 py-2';

      const name = document.createElement('span');
      name.className = 'truncate';
      name.textContent = file.file_name || 'File';

      const link = document.createElement('a');
      link.href = file.download_url || file.file_path || '#';
      link.target = '_blank';
      link.rel = 'noopener noreferrer';
      link.className = 'inline-flex items-center rounded-md border border-gray-300 px-2 py-1 text-xs font-medium hover:bg-gray-50';
      link.textContent = 'Download';

      row.appendChild(name);
      row.appendChild(link);
      viewProjectFiles.appendChild(row);
    });
  };

  const loadProjectFiles = async (requestId) => {
    if (!viewProjectFiles || !requestId) return;
    viewProjectFiles.innerHTML = '<p class="text-xs text-gray-500">Loading files...</p>';

    try {
      const res = await fetch(`/customer/projects/${requestId}/files`);
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed to load files');
      renderProjectFiles(data.files || []);
    } catch (err) {
      viewProjectFiles.innerHTML = `<p class="text-xs text-red-600">${err.message || 'Failed to load files'}</p>`;
    }
  };

  document.querySelectorAll('.viewProjectBtn').forEach(btn => {
    btn.addEventListener('click', () => {
      const row = btn.closest('tr');
      const requestId = row.dataset.id;

      viewName.textContent = row.dataset.name;
      viewDesc.textContent = row.dataset.description || 'No description';
      viewDeadline.textContent = row.dataset.deadline;

      const status = row.dataset.status;
      viewStatus.textContent = status;

      // status colors
      viewStatus.className =
        'inline-flex px-3 py-1 rounded-full text-xs font-medium ' +
        (status === 'pending' ? 'bg-yellow-100 text-yellow-800' :
         status === 'approved' ? 'bg-green-100 text-green-800' :
         status === 'rejected' ? 'bg-red-100 text-red-800' :
         status === 'in_progress' ? 'bg-blue-100 text-blue-800' :
         'bg-gray-100 text-gray-800');

      loadProjectFiles(requestId);
      openModal();
    });
  });

  closeBtns.forEach(btn => btn?.addEventListener('click', closeModal));
  overlay?.addEventListener('click', closeModal);

  document.addEventListener('keydown', e => {
    if (e.key === 'Escape') closeModal();
  });
})();
