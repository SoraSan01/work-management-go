document.addEventListener('DOMContentLoaded', () => {
  'use strict';

  const table = document.getElementById('projectsTable');
  if (table) {
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
      if (!info) return;
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
      if (!pagination) return;
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
        const btn = document.createElement('button');
        btn.type = 'button';
        btn.textContent = page;
        btn.className = [
          'rounded-lg border px-3 py-2 text-xs font-medium transition',
          page === currentPage ? 'border-blue-500 bg-blue-50 text-blue-600' : 'border-gray-200 text-gray-600 hover:bg-gray-50'
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
          <td colspan="7" class="px-4 py-10 text-center text-sm text-gray-500">
            No projects match your search.
          </td>
        `;
        tbody.appendChild(empty);
        return;
      }

      const start = (currentPage - 1) * pageSize;
      filteredRows.slice(start, start + pageSize).forEach((row) => tbody.appendChild(row));
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
      sortButtons.forEach((btn) => {
        const icon = btn.querySelector('i');
        if (!icon) return;
        if (btn.dataset.sort === currentSort.key) {
          icon.className = currentSort.direction === 'asc'
            ? 'fas fa-sort-up text-blue-600'
            : 'fas fa-sort-down text-blue-600';
          return;
        }
        icon.className = 'fas fa-sort text-gray-300';
      });
    };

    sortButtons.forEach((btn) => {
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
        filteredRows = rows.filter((row) => {
          const name = normalize(row.dataset.name);
          const status = normalize(row.dataset.status);
          const manager = normalize(row.dataset.manager);
          const employee = normalize(row.dataset.employee);
          const priority = normalize(row.dataset.priority);
          const approver = normalize(row.dataset.approver);
          return (
            name.includes(query) ||
            status.includes(query) ||
            manager.includes(query) ||
            employee.includes(query) ||
            priority.includes(query) ||
            approver.includes(query)
          );
        });
        currentPage = 1;
        render();
      });
    }

    updateSortIcons();
    render();
  }

  const modal = document.getElementById('assignProjectModal');
  if (!modal) return;

  const overlay = document.getElementById('assignProjectModalOverlay');
  const openButton = document.getElementById('openAssignProjectModal');
  const rowButtons = document.querySelectorAll('[data-assign-project]');
  const closeButtons = [
    document.getElementById('closeAssignProjectModal'),
    document.getElementById('cancelAssignProjectModal')
  ].filter(Boolean);

  const form = document.getElementById('assignProjectForm');
  const selectProject = document.getElementById('assignProjectId');
  const selectEmployee = document.getElementById('assignEmployeeId');
  const selectPriority = document.getElementById('assignProjectPriority');
  const employeeHelp = document.getElementById('assignEmployeeHelp');
  const priorityHelp = document.getElementById('assignPriorityHelp');
  const submitButton = document.getElementById('submitAssignProject');
  let lastProjectId = '';

  const openModal = () => {
    modal.classList.remove('hidden');
    modal.classList.add('flex');
    document.body.style.overflow = 'hidden';
    const firstInput = modal.querySelector('input, select, textarea, button');
    if (firstInput) firstInput.focus();
  };

  const closeModal = () => {
    modal.classList.add('hidden');
    modal.classList.remove('flex');
    document.body.style.overflow = '';
  };

  const syncFormState = () => {
    if (!selectProject || !selectEmployee || !selectPriority || !submitButton || !form) return;

    const selectedOption = selectProject.options[selectProject.selectedIndex];
    const projectId = selectedOption?.value || '';
    const allowedMemberIds = new Set(
      (selectedOption?.dataset.memberIds || '')
        .split(',')
        .map((value) => value.trim())
        .filter(Boolean)
    );
    const assignedEmployeeId = selectedOption?.dataset.assignedEmployeeId || '';
    const projectPriority = selectedOption?.dataset.priority || 'low';
    const hasProject = projectId.length > 0;
    const hasAssignment = assignedEmployeeId.length > 0;
    const projectChanged = projectId !== lastProjectId;

    Array.from(selectEmployee.options).forEach((option) => {
      if (!option.value) {
        option.hidden = false;
        option.disabled = false;
        return;
      }
      const allowed = hasProject && allowedMemberIds.has(option.value);
      option.hidden = !allowed;
      option.disabled = !allowed;
      if (!allowed && option.selected) option.selected = false;
    });

    selectEmployee.disabled = !hasProject;
    selectPriority.disabled = !hasProject;

    Array.from(selectPriority.options).forEach((option) => {
      if (!option.value) {
        option.hidden = hasProject;
        option.disabled = hasProject;
      }
    });

    if (!hasProject) {
      selectEmployee.value = '';
      selectPriority.value = '';
    } else if (projectChanged) {
      selectEmployee.value = hasAssignment ? assignedEmployeeId : '';
      selectPriority.value = projectPriority;
    }

    if (selectEmployee.value) {
      const selectedEmployeeOption = Array.from(selectEmployee.options).find(
        (option) => option.value === selectEmployee.value
      );
      if (!selectedEmployeeOption || selectedEmployeeOption.disabled) {
        selectEmployee.value = '';
      }
    }

    form.action = hasProject ? `/supervisor/projects/${projectId}/assign-employee` : '';
    submitButton.disabled = !hasProject || !selectEmployee.value || !selectPriority.value;
    submitButton.textContent = hasAssignment ? 'Update Project Assignment' : 'Assign Project';
    lastProjectId = projectId;

    if (employeeHelp) {
      if (!hasProject) {
        employeeHelp.textContent = 'Select a project to assign or review its current employee.';
      } else if (hasAssignment) {
        employeeHelp.textContent = 'This project already has an employee. You can update the assignment here.';
      } else {
        employeeHelp.textContent = 'Choose one employee from the project team.';
      }
    }

    if (priorityHelp) {
      priorityHelp.textContent = hasProject
        ? 'Set the priority that should be used for this project.'
        : 'Select a project first to set its priority.';
    }
  };

  const selectProjectById = (projectId) => {
    if (!selectProject) return;
    selectProject.value = projectId || '';
    syncFormState();
  };

  if (openButton) {
    openButton.addEventListener('click', () => {
      selectProjectById('');
      openModal();
    });
  }

  rowButtons.forEach((button) => {
    button.addEventListener('click', () => {
      const row = button.closest('tr');
      if (!row) return;
      selectProjectById(row.dataset.projectId || '');
      openModal();
    });
  });

  closeButtons.forEach((button) => button.addEventListener('click', closeModal));
  if (overlay) overlay.addEventListener('click', closeModal);

  document.addEventListener('keydown', (event) => {
    if (event.key === 'Escape' && !modal.classList.contains('hidden')) {
      closeModal();
    }
  });

  if (selectProject) {
    selectProject.addEventListener('change', syncFormState);
  }

  if (selectEmployee) {
    selectEmployee.addEventListener('change', syncFormState);
  }

  if (selectPriority) {
    selectPriority.addEventListener('change', syncFormState);
  }

  syncFormState();
});
