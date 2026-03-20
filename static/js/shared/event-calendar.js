document.addEventListener('DOMContentLoaded', () => {
  const page = document.getElementById('eventCalendarPage');
  const calendarEl = document.getElementById('calendar');
  if (!page || !calendarEl || typeof FullCalendar === 'undefined') return;

  const basePath = page.dataset.eventsBase;
  if (!basePath) return;

  const eventListEl = document.getElementById('event-list');
  const selectedDateEl = document.getElementById('selected-date');
  const selectedCountEl = document.getElementById('selected-count');

  const modal = document.getElementById('eventModal');
  const modalOverlay = document.getElementById('eventModalOverlay');
  const openCreateBtn = document.getElementById('openCreateEventModal');
  const closeButtons = [document.getElementById('closeEventModal'), document.getElementById('cancelEventModal')].filter(Boolean);

  const form = document.getElementById('eventModalForm');
  const deleteForm = document.getElementById('eventDeleteForm');
  const editToggle = document.getElementById('eventEditToggle');
  const modalTitle = document.getElementById('eventModalTitle');
  const modalSubtitle = document.getElementById('eventModalSubtitle');
  const modalSubmit = document.getElementById('eventModalSubmit');
  const modalError = document.getElementById('eventModalError');

  const inputTitle = document.getElementById('eventTitle');
  const inputDescription = document.getElementById('eventDescription');
  const inputStart = document.getElementById('eventStart');
  const inputEnd = document.getElementById('eventEnd');
  const calendarCard = calendarEl.parentElement;

  let selectedDate = new Date();
  let currentView = 'dayGridMonth';

  const toolbar = document.createElement('div');
  toolbar.className = 'mb-4 flex flex-col gap-3 border-b border-gray-100 pb-4 md:flex-row md:items-center md:justify-between';
  toolbar.innerHTML = `
    <div class="flex flex-wrap items-center gap-2">
      <button type="button" data-calendar-nav="prev" class="rounded-lg border border-gray-200 px-3 py-2 text-sm font-medium text-gray-700 transition hover:bg-gray-50">Prev</button>
      <button type="button" data-calendar-nav="today" class="rounded-lg border border-gray-200 px-3 py-2 text-sm font-medium text-gray-700 transition hover:bg-gray-50">Today</button>
      <button type="button" data-calendar-nav="next" class="rounded-lg border border-gray-200 px-3 py-2 text-sm font-medium text-gray-700 transition hover:bg-gray-50">Next</button>
    </div>
    <div class="flex flex-col gap-3 md:items-end">
      <h3 data-calendar-title class="text-base font-semibold text-gray-900"></h3>
      <div class="flex flex-wrap items-center gap-2">
        <button type="button" data-calendar-view="dayGridMonth" class="rounded-lg border border-gray-200 px-3 py-2 text-sm font-medium text-gray-700 transition hover:bg-gray-50">Month</button>
        <button type="button" data-calendar-view="timeGridWeek" class="rounded-lg border border-gray-200 px-3 py-2 text-sm font-medium text-gray-700 transition hover:bg-gray-50">Week</button>
        <button type="button" data-calendar-view="timeGridDay" class="rounded-lg border border-gray-200 px-3 py-2 text-sm font-medium text-gray-700 transition hover:bg-gray-50">Day</button>
        <button type="button" data-calendar-view="listWeek" class="rounded-lg border border-gray-200 px-3 py-2 text-sm font-medium text-gray-700 transition hover:bg-gray-50">List</button>
      </div>
    </div>
  `;

  if (calendarCard) {
    calendarCard.insertBefore(toolbar, calendarEl);
  }

  const toolbarTitle = toolbar.querySelector('[data-calendar-title]');
  const viewButtons = Array.from(toolbar.querySelectorAll('[data-calendar-view]'));

  const formatDateLabel = (date) => date.toLocaleDateString('en-US', {
    month: 'long',
    day: 'numeric',
    year: 'numeric'
  });

  const formatTimeRange = (event) => {
    const start = event.start ? new Date(event.start) : null;
    const end = event.end ? new Date(event.end) : start;
    if (!start) return '';

    const timeOptions = { hour: '2-digit', minute: '2-digit' };
    const startLabel = start.toLocaleTimeString([], timeOptions);
    const endLabel = end ? end.toLocaleTimeString([], timeOptions) : '';
    return endLabel && endLabel !== startLabel ? `${startLabel} - ${endLabel}` : startLabel;
  };

  const toLocalInput = (date) => {
    const d = new Date(date);
    const pad = (value) => value.toString().padStart(2, '0');
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
  };

  const setInputsDisabled = (disabled) => {
    [inputTitle, inputDescription, inputStart, inputEnd].forEach((input) => {
      if (input) input.disabled = disabled;
    });
  };

  const setLoadingState = (loading) => {
    if (!modalSubmit) return;
    modalSubmit.disabled = loading;
    modalSubmit.textContent = loading ? 'Saving...' : (modalSubmit.dataset.defaultLabel || 'Save Event');
  };

  const clearError = () => {
    if (!modalError) return;
    modalError.textContent = '';
    modalError.classList.add('hidden');
  };

  const showError = (message) => {
    if (!modalError) return;
    modalError.textContent = message;
    modalError.classList.remove('hidden');
  };

  const openModal = () => {
    modal.classList.remove('hidden');
    document.body.style.overflow = 'hidden';
  };

  const closeModal = () => {
    modal.classList.add('hidden');
    document.body.style.overflow = '';
    clearError();
  };

  const updateSelectedDate = (date) => {
    selectedDate = new Date(date);
    selectedDateEl.textContent = formatDateLabel(selectedDate);
    renderEventList(eventsForDate(calendar.getEvents(), selectedDate));
  };

  const setCreateMode = (date) => {
    clearError();
    modalTitle.textContent = 'New Event';
    modalSubtitle.textContent = 'Create a new calendar event.';
    modalSubmit.dataset.defaultLabel = 'Create Event';
    modalSubmit.textContent = 'Create Event';
    modalSubmit.classList.remove('hidden');
    form.action = `${basePath}/store`;
    deleteForm.classList.add('hidden');
    editToggle.classList.add('hidden');
    setInputsDisabled(false);

    const start = new Date(date);
    start.setHours(9, 0, 0, 0);
    const end = new Date(start);
    end.setHours(10, 0, 0, 0);

    inputTitle.value = '';
    inputDescription.value = '';
    inputStart.value = toLocalInput(start);
    inputEnd.value = toLocalInput(end);
  };

  const setViewMode = (event) => {
    clearError();
    modalTitle.textContent = 'Event Details';
    modalSubtitle.textContent = 'View event information.';
    modalSubmit.dataset.defaultLabel = 'Save Event';
    modalSubmit.classList.add('hidden');
    editToggle.classList.remove('hidden');
    deleteForm.classList.remove('hidden');
    form.action = `${basePath}/update/${event.id}`;
    deleteForm.action = `${basePath}/delete/${event.id}`;
    setInputsDisabled(true);

    inputTitle.value = event.title || '';
    inputDescription.value = event.extendedProps?.description || '';
    inputStart.value = toLocalInput(event.start);
    inputEnd.value = toLocalInput(event.end || event.start);
  };

  const enableEditMode = () => {
    clearError();
    modalTitle.textContent = 'Edit Event';
    modalSubtitle.textContent = 'Update event details.';
    modalSubmit.dataset.defaultLabel = 'Save Changes';
    modalSubmit.textContent = 'Save Changes';
    editToggle.classList.add('hidden');
    modalSubmit.classList.remove('hidden');
    setInputsDisabled(false);
  };

  const eventsForDate = (events, date) => {
    const dayStart = new Date(date);
    dayStart.setHours(0, 0, 0, 0);
    const dayEnd = new Date(date);
    dayEnd.setHours(23, 59, 59, 999);

    return events.filter((event) => {
      const start = new Date(event.start);
      const end = event.end ? new Date(event.end) : start;
      return start <= dayEnd && end >= dayStart;
    });
  };

  const renderEventList = (events) => {
    eventListEl.innerHTML = '';
    selectedCountEl.textContent = `${events.length} event${events.length === 1 ? '' : 's'}`;

    if (events.length === 0) {
      eventListEl.innerHTML = '<li class="rounded-xl border border-dashed border-gray-200 bg-gray-50 px-4 py-6 text-center text-sm text-gray-500">No events for this date.</li>';
      return;
    }

    events
      .sort((left, right) => new Date(left.start) - new Date(right.start))
      .forEach((event) => {
        const item = document.createElement('li');
        item.className = 'cursor-pointer rounded-xl border border-gray-200 bg-white px-4 py-3 shadow-sm transition hover:border-blue-200 hover:bg-blue-50';
        item.innerHTML = `
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <p class="text-sm font-semibold text-gray-900">${event.title}</p>
              <p class="mt-1 text-xs text-gray-500">${formatTimeRange(event)}</p>
              ${event.extendedProps?.description ? `<p class="mt-2 text-xs text-gray-500">${event.extendedProps.description}</p>` : ''}
            </div>
            <span class="mt-1 h-2.5 w-2.5 rounded-full bg-blue-500"></span>
          </div>
        `;
        item.addEventListener('click', () => {
          updateSelectedDate(event.start);
          setViewMode(event);
          openModal();
        });
        eventListEl.appendChild(item);
      });
  };

  const request = async (url, options = {}) => {
    const response = await fetch(url, {
      ...options,
      headers: {
        Accept: 'application/json',
        'X-Requested-With': 'XMLHttpRequest',
        ...(options.headers || {})
      }
    });

    let payload = null;
    const contentType = response.headers.get('content-type') || '';
    if (contentType.includes('application/json')) {
      payload = await response.json();
    } else {
      const text = await response.text();
      payload = { error: text || 'Request failed' };
    }

    if (!response.ok) {
      throw new Error(payload.error || 'Request failed');
    }

    return payload;
  };

  const loadEvents = async (fetchInfo, successCallback, failureCallback) => {
    try {
      const url = new URL(`${basePath}/json`, window.location.origin);
      url.searchParams.set('start', fetchInfo.startStr);
      url.searchParams.set('end', fetchInfo.endStr);
      url.searchParams.set('_ts', Date.now().toString());

      const response = await fetch(url.toString(), {
        headers: { Accept: 'application/json' },
        cache: 'no-store'
      });

      if (!response.ok) {
        throw new Error('Failed to load events');
      }

      const events = await response.json();
      successCallback(events);
    } catch (error) {
      failureCallback(error);
    }
  };

  const updateToolbarState = (view, title) => {
    currentView = view;
    if (toolbarTitle) {
      toolbarTitle.textContent = title;
    }

    viewButtons.forEach((button) => {
      const isActive = button.dataset.calendarView === currentView;
      button.classList.toggle('bg-blue-600', isActive);
      button.classList.toggle('border-blue-600', isActive);
      button.classList.toggle('text-white', isActive);
      button.classList.toggle('text-gray-700', !isActive);
      button.classList.toggle('hover:bg-gray-50', !isActive);
    });
  };

  const calendar = new FullCalendar.Calendar(calendarEl, {
    initialView: 'dayGridMonth',
    height: 'auto',
    headerToolbar: false,
    selectable: true,
    editable: false,
    nowIndicator: true,
    events: loadEvents,
    dateClick: (info) => updateSelectedDate(info.date),
    eventClick: (info) => {
      updateSelectedDate(info.event.start);
      setViewMode(info.event);
      openModal();
    },
    eventsSet: (events) => renderEventList(eventsForDate(events, selectedDate)),
    datesSet: (info) => updateToolbarState(info.view.type, info.view.title)
  });

  calendar.render();
  updateSelectedDate(selectedDate);

  toolbar.querySelector('[data-calendar-nav="prev"]')?.addEventListener('click', () => calendar.prev());
  toolbar.querySelector('[data-calendar-nav="today"]')?.addEventListener('click', () => {
    calendar.today();
    updateSelectedDate(new Date());
  });
  toolbar.querySelector('[data-calendar-nav="next"]')?.addEventListener('click', () => calendar.next());
  viewButtons.forEach((button) => {
    button.addEventListener('click', () => {
      const viewName = button.dataset.calendarView;
      if (viewName) {
        calendar.changeView(viewName);
      }
    });
  });

  form.addEventListener('submit', async (event) => {
    event.preventDefault();
    clearError();

    if (inputEnd.value <= inputStart.value) {
      showError('End time must be after start time.');
      return;
    }

    try {
      setLoadingState(true);
      await request(form.action, {
        method: 'POST',
        body: new FormData(form)
      });
      closeModal();
      calendar.refetchEvents();
      setTimeout(() => updateSelectedDate(selectedDate), 150);
    } catch (error) {
      showError(error.message);
    } finally {
      setLoadingState(false);
    }
  });

  deleteForm.addEventListener('submit', async (event) => {
    event.preventDefault();
    clearError();

    try {
      await request(deleteForm.action, { method: 'POST' });
      closeModal();
      calendar.refetchEvents();
      setTimeout(() => updateSelectedDate(selectedDate), 150);
    } catch (error) {
      showError(error.message);
    }
  });

  openCreateBtn?.addEventListener('click', () => {
    setCreateMode(selectedDate);
    openModal();
  });

  editToggle?.addEventListener('click', enableEditMode);
  closeButtons.forEach((button) => button.addEventListener('click', closeModal));
  modalOverlay?.addEventListener('click', closeModal);
  document.addEventListener('keydown', (event) => {
    if (event.key === 'Escape' && !modal.classList.contains('hidden')) {
      closeModal();
    }
  });
});
