(() => {
  'use strict';

  const table = document.getElementById('employeesTable');
  if (!table) return;

  const tbody = table.querySelector('tbody');
  const rows = Array.from(tbody.querySelectorAll('tr'));
  const searchInput = document.getElementById('employeesSearch');
  const info = document.getElementById('employeesTableInfo');
  const pagination = document.getElementById('employeesTablePagination');
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
    info.textContent = `Showing ${start} to ${end} of ${total} employees`;
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
        <td colspan="4" class="px-4 py-10 text-center text-sm text-gray-500">
          No employees match your search.
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
        const role = normalize(row.dataset.role);
        const nationality = normalize(row.dataset.nationality);
        return name.includes(query) || role.includes(query) || nationality.includes(query);
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

  const modal = document.getElementById('addEmployeeModal');
  if (!modal) return;

  const overlay = document.getElementById('addEmployeeModalOverlay');
  const openButtons = [
    document.getElementById('openAddEmployeeModal'),
    document.getElementById('openAddEmployeeModalEmpty')
  ].filter(Boolean);
  const closeButtons = [
    document.getElementById('closeAddEmployeeModal'),
    document.getElementById('cancelAddEmployeeModal')
  ].filter(Boolean);

  const form = document.getElementById('employeeModalForm');
  const title = document.getElementById('employeeModalTitle');
  const subtitle = document.getElementById('employeeModalSubtitle');
  const submitBtn = document.getElementById('employeeModalSubmit');
  const helpText = document.getElementById('employeePasswordHelp');

  const inputFirst = document.getElementById('employeeFirstName');
  const inputLast = document.getElementById('employeeLastName');
  const inputEmail = document.getElementById('employeeEmail');
  const inputNationality = document.getElementById('employeeNationality');
  const selectRole = document.getElementById('employeeRoleId');
  const selectDepartment = document.getElementById('employeeDepartmentId');
  const selectManager = document.getElementById('employeeManagerId');
  const inputPassword = document.getElementById('employeePassword');
  const inputConfirm = document.getElementById('employeeConfirmPassword');

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
    if (title) title.textContent = 'Add Employee';
    if (subtitle) subtitle.textContent = 'Create a new team member profile.';
    if (submitBtn) submitBtn.textContent = 'Create Employee';
    if (helpText) helpText.textContent = 'Password is required when creating a new employee.';
    if (form) form.action = '/admin/employees/store';
    if (inputFirst) inputFirst.value = '';
    if (inputLast) inputLast.value = '';
    if (inputEmail) inputEmail.value = '';
    if (inputNationality) inputNationality.value = '';
    if (selectRole) selectRole.value = '';
    if (selectDepartment) selectDepartment.value = '';
    if (selectManager) selectManager.value = '';
    if (inputPassword) {
      inputPassword.value = '';
      inputPassword.required = true;
    }
    if (inputConfirm) {
      inputConfirm.value = '';
      inputConfirm.required = true;
    }
  };

  const setEditMode = (data) => {
    if (title) title.textContent = 'Edit Employee';
    if (subtitle) subtitle.textContent = 'Update employee details.';
    if (submitBtn) submitBtn.textContent = 'Save Changes';
    if (helpText) helpText.textContent = 'Leave password blank to keep the current password.';
    if (form && data.editUrl) form.action = data.editUrl;
    if (inputFirst) inputFirst.value = data.firstName || '';
    if (inputLast) inputLast.value = data.lastName || '';
    if (inputEmail) inputEmail.value = data.email || '';
    if (inputNationality) inputNationality.value = data.nationality || '';
    if (selectRole) selectRole.value = data.roleId || '';
    if (selectDepartment) selectDepartment.value = data.departmentId || '';
    if (selectManager) selectManager.value = data.managerId || '';
    if (inputPassword) {
      inputPassword.value = '';
      inputPassword.required = false;
    }
    if (inputConfirm) {
      inputConfirm.value = '';
      inputConfirm.required = false;
    }
  };

  openButtons.forEach(btn => btn.addEventListener('click', () => {
    setCreateMode();
    openModal();
  }));

  document.addEventListener('click', (event) => {
    const btn = event.target.closest('[data-edit-employee]');
    if (!btn) return;

    setEditMode({
      editUrl: btn.dataset.editUrl,
      firstName: btn.dataset.firstName,
      lastName: btn.dataset.lastName,
      email: btn.dataset.email,
      nationality: btn.dataset.nationality,
      roleId: btn.dataset.roleId,
      departmentId: btn.dataset.departmentId,
      managerId: btn.dataset.managerId
    });
    openModal();
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
