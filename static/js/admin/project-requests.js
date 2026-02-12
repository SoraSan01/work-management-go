(() => {
  'use strict';

  const table = document.getElementById('requestsTable');
  if (!table) return;

  const tbody = table.querySelector('tbody');
  const rows = Array.from(tbody.querySelectorAll('tr'));
  const searchInput = document.getElementById('requestsSearch');
  const info = document.getElementById('requestsTableInfo');
  const pagination = document.getElementById('requestsTablePagination');
  const sortButtons = table.querySelectorAll('[data-sort]');

  let currentPage = 1;
  const pageSize = 10;
  let currentSort = { key: 'title', direction: 'asc' };
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
    info.textContent = `Showing ${start} to ${end} of ${total} requests`;
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
        <td colspan="6" class="px-4 py-10 text-center text-sm text-gray-500">
          No requests match your search.
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
        const title = normalize(row.dataset.title);
        const customer = normalize(row.dataset.customer);
        const company = normalize(row.dataset.company);
        const deadline = normalize(row.dataset.deadline);
        const status = normalize(row.dataset.status);
        return (
          title.includes(query) ||
          customer.includes(query) ||
          company.includes(query) ||
          deadline.includes(query) ||
          status.includes(query)
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

  const approveModal = document.getElementById('approveModal');
  const approveForm = document.getElementById('approveForm');
  const approveOverlay = document.getElementById('approveModalOverlay');
  const approveCloseButtons = [
    document.getElementById('closeApproveModal'),
    document.getElementById('cancelApproveModal')
  ].filter(Boolean);

  const openApproveModal = (requestId) => {
    if (!approveModal) return;
    if (approveForm) approveForm.action = `/admin/projects/approve/${requestId}`;
    approveModal.classList.remove('hidden');
    document.body.style.overflow = 'hidden';
    const firstInput = approveModal.querySelector('select, input, textarea');
    if (firstInput) firstInput.focus();
  };

  const closeApproveModal = () => {
    if (!approveModal) return;
    approveModal.classList.add('hidden');
    document.body.style.overflow = '';
    if (approveForm) approveForm.reset();
  };

  document.querySelectorAll('[data-approve-request]').forEach(btn => {
    btn.addEventListener('click', () => {
      openApproveModal(btn.dataset.approveId);
    });
  });

  if (approveOverlay) {
    approveOverlay.addEventListener('click', closeApproveModal);
  }

  approveCloseButtons.forEach(btn => btn.addEventListener('click', closeApproveModal));

  const modal = document.getElementById('rejectModal');
  if (!modal) return;

  const overlay = document.getElementById('rejectModalOverlay');
  const closeButtons = [
    document.getElementById('closeRejectModal'),
    document.getElementById('cancelRejectModal')
  ].filter(Boolean);

  const form = document.getElementById('rejectForm');

  const openModal = (requestId) => {
    if (form) form.action = `/admin/projects/reject/${requestId}`;
    modal.classList.remove('hidden');
    document.body.style.overflow = 'hidden';
    const firstInput = modal.querySelector('textarea, input, select');
    if (firstInput) firstInput.focus();
  };

  const closeModal = () => {
    modal.classList.add('hidden');
    document.body.style.overflow = '';
    if (form) form.reset();
  };

  document.querySelectorAll('[data-reject-request]').forEach(btn => {
    btn.addEventListener('click', () => {
      openModal(btn.dataset.rejectId);
    });
  });

  closeButtons.forEach(btn => btn.addEventListener('click', closeModal));

  if (overlay) {
    overlay.addEventListener('click', closeModal);
  }

  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && approveModal && !approveModal.classList.contains('hidden')) {
      closeApproveModal();
      return;
    }
    if (e.key === 'Escape' && !modal.classList.contains('hidden')) {
      closeModal();
    }
  });
})();
