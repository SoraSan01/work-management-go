(() => {
  'use strict';
  
  const init = () => {
    // Auto-expand active menu section
    const activeLinks = document.querySelectorAll('.nav-submenu a.bg-blue-50');
    activeLinks.forEach(link => {
      const menu = link.closest('.nav-submenu');
      if (menu) {
        menu.classList.remove('hidden');
        const toggle = menu.previousElementSibling;
        if (toggle?.classList.contains('nav-toggle')) {
          const chevron = toggle.querySelector('.fa-chevron-down');
          if (chevron) chevron.style.transform = 'rotate(180deg)';
          toggle.setAttribute('aria-expanded', 'true');
        }
      }
    });
    
    // Setup toggle handlers
    document.querySelectorAll('.nav-toggle').forEach(button => {
      button.addEventListener('click', () => {
        const targetId = button.dataset.target;
        const menu = document.getElementById(targetId);
        const chevron = button.querySelector('.fa-chevron-down');
        
        if (menu && chevron) {
          const isHidden = menu.classList.toggle('hidden');
          chevron.style.transform = isHidden ? 'rotate(0deg)' : 'rotate(180deg)';
          button.setAttribute('aria-expanded', isHidden ? 'false' : 'true');
        }
      });
    });
  };
  
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();