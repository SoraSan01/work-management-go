(() => {
  'use strict';

  const table = document.getElementById('projectsTable');
  if (!table) return;

  const tbody = table.querySelector('tbody');
  const rows = Array.from(tbody.querySelectorAll('tr'));
  const searchInput = document.getElementById('projectsSearch');
  const info = document.getElementById('projectsTableInfo');
  const pagination = document.getElementById('projectsTablePagination');
  const sortButtons = table.querySelectorAll('[data-sort]');

  let currentPage = 1;
  const pageSize = 10;
  let currentSort = { key: 'name', direction: 'asc' };
  let filteredRows = rows.slice();

  const normalize = (value) => (value || '').toString().toLowerCase().trim();

  const compare = (a, b) => {
    const aValue = normalize(a.dataset[currentSort.key]);
    const bValue = normalize(b.dataset[currentSort.key]);
    if (aValue === bValue) return 0;
    if (currentSort.direction === 'asc') {
      return aValue > bValue ? 1 : -1;
    }
    return aValue < bValue ? 1 : -1;
  };

  const renderInfo = () => {
    const total = filteredRows.length;
    if (total === 0) {
      info.textContent = 'No results found.';
      return;
    }
    const start = (currentPage - 1) * pageSize + 1;
    const end = Math.min(currentPage * pageSize, total);
    info.textContent = `Showing ${start} to ${end} of ${total} projects`;
  };

  const renderPagination = () => {
    pagination.innerHTML = '';
    const totalPages = Math.max(1, Math.ceil(filteredRows.length / pageSize));

    const button = (label, page, disabled = false) => {
      const btn = document.createElement('button');
      btn.type = 'button';
      btn.textContent = label;
      btn.className = [
        'rounded-lg border px-3 py-2 text-xs font-medium transition',
        disabled ? 'cursor-not-allowed border-gray-200 text-gray-300' : 'border-gray-200 text-gray-600 hover:bg-gray-50'
      ].join(' ');
      btn.disabled = disabled;
      btn.addEventListener('click', () => {
        currentPage = page;
        render();
      });
      return btn;
    };

    pagination.appendChild(button('Prev', Math.max(1, currentPage - 1), currentPage === 1));

    for (let page = 1; page <= totalPages; page += 1) {
      const isActive = page === currentPage;
      const btn = document.createElement('button');
      btn.type = 'button';
      btn.textContent = page;
      btn.className = [
        'rounded-lg border px-3 py-2 text-xs font-medium transition',
        isActive ? 'border-blue-500 bg-blue-50 text-blue-600' : 'border-gray-200 text-gray-600 hover:bg-gray-50'
      ].join(' ');
      btn.addEventListener('click', () => {
        currentPage = page;
        render();
      });
      pagination.appendChild(btn);
    }

    pagination.appendChild(button('Next', Math.min(totalPages, currentPage + 1), currentPage === totalPages));
  };

  const renderRows = () => {
    tbody.innerHTML = '';
    if (filteredRows.length === 0) {
      const empty = document.createElement('tr');
      empty.innerHTML = `
        <td colspan="5" class="px-4 py-10 text-center text-sm text-gray-500">
          No projects match your search.
        </td>
      `;
      tbody.appendChild(empty);
      return;
    }

    const start = (currentPage - 1) * pageSize;
    const pageRows = filteredRows.slice(start, start + pageSize);
    pageRows.forEach(row => tbody.appendChild(row));
  };

  const render = () => {
    filteredRows.sort(compare);
    const totalPages = Math.max(1, Math.ceil(filteredRows.length / pageSize));
    if (currentPage > totalPages) currentPage = totalPages;
    renderRows();
    renderInfo();
    renderPagination();
  };

  const updateSortIcons = () => {
    sortButtons.forEach(btn => {
      const icon = btn.querySelector('i');
      if (!icon) return;
      const key = btn.dataset.sort;
      if (key === currentSort.key) {
        icon.className = currentSort.direction === 'asc'
          ? 'fas fa-sort-up text-blue-600'
          : 'fas fa-sort-down text-blue-600';
      } else {
        icon.className = 'fas fa-sort text-gray-300';
      }
    });
  };

  sortButtons.forEach(btn => {
    btn.addEventListener('click', () => {
      const key = btn.dataset.sort;
      if (currentSort.key === key) {
        currentSort.direction = currentSort.direction === 'asc' ? 'desc' : 'asc';
      } else {
        currentSort = { key, direction: 'asc' };
      }
      currentPage = 1;
      updateSortIcons();
      render();
    });
  });

  if (searchInput) {
    searchInput.addEventListener('input', () => {
      const query = normalize(searchInput.value);
      filteredRows = rows.filter(row => {
        const name = normalize(row.dataset.name);
        const status = normalize(row.dataset.status);
        const manager = normalize(row.dataset.manager);
        const approver = normalize(row.dataset.approver);
        return (
          name.includes(query) ||
          status.includes(query) ||
          manager.includes(query) ||
          approver.includes(query)
        );
      });
      currentPage = 1;
      render();
    });
  }

  updateSortIcons();
  render();
})();

(() => {
  'use strict';

  const modal = document.getElementById('projectModal');
  if (!modal) return;

  const overlay = document.getElementById('projectModalOverlay');
  const openButtons = [
    document.getElementById('openAddProjectModal'),
    document.getElementById('openAddProjectModalEmpty')
  ].filter(Boolean);
  const closeButtons = [
    document.getElementById('closeProjectModal'),
    document.getElementById('cancelProjectModal')
  ].filter(Boolean);

  const form = document.getElementById('projectModalForm');
  const title = document.getElementById('projectModalTitle');
  const subtitle = document.getElementById('projectModalSubtitle');
  const submitBtn = document.getElementById('projectModalSubmit');

  const inputName = document.getElementById('projectName');
  const inputDescription = document.getElementById('projectDescription');
  const selectStatus = document.getElementById('projectStatus');
  const selectManager = document.getElementById('projectManagerId');
  const selectApprover = document.getElementById('projectApproverId');

  const openModal = () => {
    modal.classList.remove('hidden');
    document.body.style.overflow = 'hidden';
    const firstInput = modal.querySelector('input, select, textarea');
    if (firstInput) firstInput.focus();
  };

  const closeModal = () => {
    modal.classList.add('hidden');
    document.body.style.overflow = '';
  };

  const setCreateMode = () => {
    if (title) title.textContent = 'Add Project';
    if (subtitle) subtitle.textContent = 'Create a new project for your team.';
    if (submitBtn) submitBtn.textContent = 'Create Project';
    if (form) form.action = '/admin/projects/store';
    if (inputName) inputName.value = '';
    if (inputDescription) inputDescription.value = '';
    if (selectStatus) selectStatus.value = 'draft';
    if (selectManager) selectManager.value = '';
    if (selectApprover) selectApprover.value = '';
  };

  const setEditMode = (data) => {
    if (title) title.textContent = 'Edit Project';
    if (subtitle) subtitle.textContent = 'Update project details.';
    if (submitBtn) submitBtn.textContent = 'Save Changes';
    if (form && data.editUrl) form.action = data.editUrl;
    if (inputName) inputName.value = data.name || '';
    if (inputDescription) inputDescription.value = data.description || '';
    if (selectStatus && data.status) selectStatus.value = data.status;
    if (selectManager) selectManager.value = data.managerId || '';
    if (selectApprover) selectApprover.value = data.approverId || '';
  };

  openButtons.forEach(btn => btn.addEventListener('click', () => {
    setCreateMode();
    openModal();
  }));

  document.querySelectorAll('[data-edit-project]').forEach(btn => {
    btn.addEventListener('click', () => {
      const row = btn.closest('tr');
      if (!row) return;
      setEditMode({
        editUrl: btn.dataset.editUrl,
        name: row.dataset.name,
        description: row.dataset.description,
        status: row.dataset.status,
        managerId: row.dataset.managerId,
        approverId: row.dataset.approverId
      });
      openModal();
    });
  });

  closeButtons.forEach(btn => btn.addEventListener('click', closeModal));

  if (overlay) {
    overlay.addEventListener('click', closeModal);
  }

  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && !modal.classList.contains('hidden')) {
      closeModal();
    }
  });
})();
