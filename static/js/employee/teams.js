(() => {
  'use strict';

  const table = document.getElementById('teamsTable');
  if (!table) return;

  const tbody = table.querySelector('tbody');
  const rows = Array.from(tbody.querySelectorAll('tr'));
  const searchInput = document.getElementById('teamsSearch');
  const info = document.getElementById('teamsTableInfo');
  const pagination = document.getElementById('teamsTablePagination');
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
    info.textContent = `Showing ${start} to ${end} of ${total} teams`;
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
          No teams match your search.
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
        const supervisor = normalize(row.dataset.supervisor);
        const members = normalize(row.dataset.members);
        const created = normalize(row.dataset.created);
        return (
          name.includes(query) ||
          supervisor.includes(query) ||
          members.includes(query) ||
          created.includes(query)
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

  const modal = document.getElementById('teamModal');
  if (!modal) return;

  const overlay = document.getElementById('teamModalOverlay');
  const openButtons = [
    document.getElementById('openAddTeamModal'),
    document.getElementById('openAddTeamModalEmpty')
  ].filter(Boolean);
  const closeButtons = [
    document.getElementById('closeTeamModal'),
    document.getElementById('cancelTeamModal')
  ].filter(Boolean);

  const form = document.getElementById('teamModalForm');
  const title = document.getElementById('teamModalTitle');
  const subtitle = document.getElementById('teamModalSubtitle');
  const submitBtn = document.getElementById('teamModalSubmit');

  const inputName = document.getElementById('teamName');
  const selectSupervisor = document.getElementById('teamSupervisorId');
  const selectMembers = document.getElementById('teamMemberIds');
  const memberSearch = document.getElementById('teamMemberSearch');

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

  const clearMultiSelect = (select) => {
    if (!select) return;
    Array.from(select.options).forEach(option => {
      option.selected = false;
    });
  };

  const normalize = (value) => (value || '').toString().toLowerCase().trim();

  const setOptionEligibility = (select, teamId) => {
    if (!select) return;
    Array.from(select.options).forEach(option => {
      const optionTeamId = option.dataset.teamId || '';
      const eligible = optionTeamId === '' || optionTeamId === teamId;
      option.dataset.eligible = eligible ? 'true' : 'false';
      option.hidden = !eligible;
      option.disabled = !eligible;
      if (!eligible) option.selected = false;
    });
  };

  const applyMemberSearch = () => {
    if (!selectMembers || !memberSearch) return;
    const query = normalize(memberSearch.value);
    Array.from(selectMembers.options).forEach(option => {
      const eligible = option.dataset.eligible === 'true';
      if (!eligible) return;
      const text = normalize(option.textContent);
      option.hidden = query ? !text.includes(query) : false;
    });
  };

  const setCreateMode = () => {
    if (title) title.textContent = 'Create Team';
    if (subtitle) subtitle.textContent = 'Set up a team and assign members.';
    if (submitBtn) submitBtn.textContent = 'Create Team';
    if (form) form.action = '/employee/teams/store';
    if (inputName) inputName.value = '';
    if (selectSupervisor) selectSupervisor.value = '';
    clearMultiSelect(selectMembers);
    if (memberSearch) memberSearch.value = '';
    setOptionEligibility(selectSupervisor, '');
    setOptionEligibility(selectMembers, '');
    applyMemberSearch();
  };

  const setEditMode = (row) => {
    if (title) title.textContent = 'Edit Team';
    if (subtitle) subtitle.textContent = 'Update team details and members.';
    if (submitBtn) submitBtn.textContent = 'Save Changes';
    if (form && row.dataset.teamId) form.action = `/employee/teams/update/${row.dataset.teamId}`;
    if (inputName) inputName.value = row.dataset.name || '';
    if (selectSupervisor) selectSupervisor.value = row.dataset.supervisorId || '';
    clearMultiSelect(selectMembers);
    if (memberSearch) memberSearch.value = '';
    const currentTeamId = row.dataset.teamId || '';
    setOptionEligibility(selectSupervisor, currentTeamId);
    setOptionEligibility(selectMembers, currentTeamId);
    if (selectMembers) {
      const memberIds = (row.dataset.memberIds || '')
        .split(',')
        .map(id => id.trim())
        .filter(Boolean);
      Array.from(selectMembers.options).forEach(option => {
        if (memberIds.includes(option.value)) {
          option.selected = true;
        }
      });
    }
    applyMemberSearch();
  };

  openButtons.forEach(btn => btn.addEventListener('click', () => {
    setCreateMode();
    openModal();
  }));

  document.querySelectorAll('[data-edit-team]').forEach(btn => {
    btn.addEventListener('click', () => {
      const row = btn.closest('tr');
      if (!row) return;
      setEditMode(row);
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

  if (memberSearch) {
    memberSearch.addEventListener('input', applyMemberSearch);
  }
})();

(() => {
  'use strict';

  const viewModal = document.getElementById('teamViewModal');
  if (!viewModal) return;

  const overlay = document.getElementById('teamViewModalOverlay');
  const closeButtons = [
    document.getElementById('closeTeamViewModal'),
    document.getElementById('closeTeamViewModalFooter')
  ].filter(Boolean);
  const viewTitle = document.getElementById('teamViewTitle');
  const viewSupervisor = document.getElementById('teamViewSupervisor');
  const viewMembers = document.getElementById('teamViewMembers');

  const openView = (row) => {
    if (viewTitle) viewTitle.textContent = row.dataset.name || 'Team';
    if (viewSupervisor) viewSupervisor.textContent = row.dataset.supervisor || '-';
    if (viewMembers) {
      viewMembers.innerHTML = '';
      const members = (row.dataset.members || '').split('|').map(v => v.trim()).filter(Boolean);
      if (members.length === 0) {
        const li = document.createElement('li');
        li.textContent = 'No members assigned.';
        viewMembers.appendChild(li);
      } else {
        members.forEach(name => {
          const li = document.createElement('li');
          li.textContent = name;
          viewMembers.appendChild(li);
        });
      }
    }
    viewModal.classList.remove('hidden');
    document.body.style.overflow = 'hidden';
  };

  const closeView = () => {
    viewModal.classList.add('hidden');
    document.body.style.overflow = '';
  };

  document.querySelectorAll('[data-view-team]').forEach(button => {
    button.addEventListener('click', () => {
      const row = button.closest('tr');
      if (!row) return;
      openView(row);
    });
  });

  closeButtons.forEach(btn => btn.addEventListener('click', closeView));
  if (overlay) overlay.addEventListener('click', closeView);

  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && !viewModal.classList.contains('hidden')) {
      closeView();
    }
  });
})();
