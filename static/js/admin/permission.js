(() => {
    'use strict';

    const table = document.getElementById('permissionsTable');
    if (!table) return;

    const tbody = table.querySelector('tbody');
    const rows = Array.from(tbody.querySelectorAll('tr'));
    const searchInput = document.getElementById('permissionsSearch');
    const info = document.getElementById('permissionsTableInfo');
    const pagination = document.getElementById('permissionsTablePagination');
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
                <td colspan="3" class="px-4 py-10 text-center text-sm text-gray-500">
                No permissions match your search.
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
            return name.includes(query);
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

  const modal = document.getElementById('addPermissionModal');
  if (!modal) return;

  const overlay = document.getElementById('addPermissionModalOverlay');
  const openButtons = [
    document.getElementById('openAddPermissionModal'),
    document.getElementById('openAddPermissionModalEmpty')
  ].filter(Boolean);
  const closeButtons = [
    document.getElementById('closeAddPermissionModal'),
    document.getElementById('cancelAddPermissionModal')
  ].filter(Boolean);

  const form = document.getElementById('permissionModalForm');
  const title = document.getElementById('permissionModalTitle');
  const subtitle = document.getElementById('permissionModalSubtitle');
  const submitBtn = document.getElementById('permissionModalSubmit');

  const inputName = document.getElementById('permissionName');

  const openModal = () => {
    modal.classList.remove('hidden');
    document.body.style.overflow = 'hidden';
    if (inputName) inputName.focus();
  };

  const closeModal = () => {
    modal.classList.add('hidden');
    document.body.style.overflow = '';
  };

  const setCreateMode = () => {
    if (title) title.textContent = 'Add Permission';
    if (subtitle) subtitle.textContent = 'Create a new Permission.';
    if (submitBtn) submitBtn.textContent = 'Create Permission';
    if (form) form.action = '/admin/permissions/store';
    if (inputName) inputName.value = '';
  };

  const setEditMode = (data) => {
    if (title) title.textContent = 'Edit Permission';
    if (subtitle) subtitle.textContent = 'Update permission details.';
    if (submitBtn) submitBtn.textContent = 'Save Changes';
    if (form && data.editUrl) form.action = data.editUrl;
    if (inputName) inputName.value = data.permissionName || '';
  };

  openButtons.forEach(btn => btn.addEventListener('click', () => {
    setCreateMode();
    openModal();
  }));

  document.querySelectorAll('[data-edit-permission]').forEach(btn => {
    btn.addEventListener('click', () => {
      setEditMode({
        editUrl: btn.dataset.editUrl,
        permissionName: btn.dataset.permissionName,
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