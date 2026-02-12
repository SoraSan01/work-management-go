document.addEventListener('DOMContentLoaded', function() {
  const calendarEl = document.getElementById('calendar');
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

  const inputTitle = document.getElementById('eventTitle');
  const inputDescription = document.getElementById('eventDescription');
  const inputStart = document.getElementById('eventStart');
  const inputEnd = document.getElementById('eventEnd');

  let selectedDate = new Date();
  let currentEvent = null;

  const formatDateLabel = (date) => date.toLocaleDateString('en-US', { month: 'long', day: 'numeric', year: 'numeric' });
  const formatTime = (date) => date ? new Date(date).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : '';
  const toLocalInput = (date) => {
    const d = new Date(date);
    const pad = n => n.toString().padStart(2,'0');
    return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
  };
  const setInputsDisabled = (disabled) => [inputTitle, inputDescription, inputStart, inputEnd].forEach(i => i && (i.disabled = disabled));

  const openModal = () => { modal.classList.remove('hidden'); document.body.style.overflow='hidden'; };
  const closeModal = () => { modal.classList.add('hidden'); document.body.style.overflow=''; };

  const setCreateMode = date => {
    currentEvent = null;
    modalTitle.textContent='New Event';
    modalSubtitle.textContent='Create a new calendar event.';
    modalSubmit.textContent='Create Event';
    form.action='/employee/events/store';
    deleteForm.classList.add('hidden');
    editToggle.classList.add('hidden');
    modalSubmit.classList.remove('hidden');
    setInputsDisabled(false);
    const start = new Date(date); start.setHours(9,0,0,0);
    const end = new Date(start); end.setHours(10,0,0,0);
    inputTitle.value=''; inputDescription.value=''; inputStart.value=toLocalInput(start); inputEnd.value=toLocalInput(end);
  };

  const setViewMode = event => {
    currentEvent=event;
    modalTitle.textContent='Event Details';
    modalSubtitle.textContent='View event information.';
    modalSubmit.classList.add('hidden');
    editToggle.classList.remove('hidden');
    deleteForm.classList.remove('hidden');
    form.action=`/employee/events/update/${event.id}`;
    deleteForm.action=`/employee/events/delete/${event.id}`;
    setInputsDisabled(true);
    inputTitle.value=event.title||'';
    inputDescription.value=event.extendedProps?.description||'';
    inputStart.value=toLocalInput(event.start);
    inputEnd.value=toLocalInput(event.end||event.start);
  };

  const enableEditMode = () => {
    modalTitle.textContent='Edit Event';
    modalSubtitle.textContent='Update event details.';
    editToggle.classList.add('hidden');
    modalSubmit.classList.remove('hidden');
    setInputsDisabled(false);
  };

  const eventsForDate = (events, date) => {
    const dayStart = new Date(date); dayStart.setHours(0,0,0,0);
    const dayEnd = new Date(date); dayEnd.setHours(23,59,59,999);
    return events.filter(e => e.start <= dayEnd && (e.end||e.start) >= dayStart);
  };

  const renderEventList = events => {
    eventListEl.innerHTML='';
    selectedCountEl.textContent=`${events.length} event${events.length===1?'':'s'}`;
    if(events.length===0) {
      eventListEl.innerHTML=`<li class="rounded-xl border border-dashed border-gray-200 bg-gray-50 px-4 py-6 text-center text-sm text-gray-500">No events for this date.</li>`;
      return;
    }
    events.forEach(event=>{
      const li=document.createElement('li');
      li.className='rounded-xl border border-gray-200 bg-white px-4 py-3 shadow-sm hover:border-blue-200 hover:bg-blue-50 cursor-pointer';
      li.innerHTML=`<div class="flex items-start justify-between gap-3">
        <div>
          <p class="text-sm font-semibold text-gray-900">${event.title}</p>
          <p class="text-xs text-gray-500">${formatTime(event.start)}</p>
        </div>
        <span class="mt-1 h-2.5 w-2.5 rounded-full" style="background-color:${event.backgroundColor||'#3b82f6'}"></span>
      </div>`;
      li.addEventListener('click',()=>{ setViewMode(event); openModal(); });
      eventListEl.appendChild(li);
    });
  };

  const calendar = new FullCalendar.Calendar(calendarEl, {
    initialView:'dayGridMonth',
    height:'auto',
    headerToolbar:{ left:'prev,next today', center:'title', right:'dayGridMonth,timeGridWeek,timeGridDay,listWeek' },
    selectable:true,
    editable:false,
    events:'/employee/events/json',
    dateClick: info => { selectedDate=info.date; selectedDateEl.textContent=formatDateLabel(selectedDate); renderEventList(eventsForDate(calendar.getEvents(),selectedDate)); },
    eventClick: info => { setViewMode(info.event); openModal(); },
    eventsSet: events => { selectedDateEl.textContent=formatDateLabel(selectedDate); renderEventList(eventsForDate(events,selectedDate)); }
  });
  calendar.render();

  openCreateBtn?.addEventListener('click',()=>{ setCreateMode(selectedDate); openModal(); });
  editToggle?.addEventListener('click',enableEditMode);
  closeButtons.forEach(btn=>btn.addEventListener('click',closeModal));
  modalOverlay?.addEventListener('click',closeModal);
  document.addEventListener('keydown', e=>{ if(e.key==='Escape'&&!modal.classList.contains('hidden')) closeModal(); });
});