(() => {
  'use strict';
  
  // Dropdown management
  class Dropdown {
    constructor(toggleId, dropdownId) {
      this.toggle = document.getElementById(toggleId);
      this.dropdown = document.getElementById(dropdownId);
      
      if (this.toggle && this.dropdown) {
        this.toggle.addEventListener('click', (e) => {
          e.stopPropagation();
          this.toggleDropdown();
        });
      }
    }
    
    toggleDropdown() {
      const isHidden = this.dropdown.classList.toggle('hidden');
      this.toggle.setAttribute('aria-expanded', !isHidden);
      
      if (!isHidden) {
        // Close other dropdowns
        document.querySelectorAll('[role="menu"]:not(#' + this.dropdown.id + ')').forEach(menu => {
          menu.classList.add('hidden');
        });
      }
    }
    
    close() {
      this.dropdown.classList.add('hidden');
      this.toggle.setAttribute('aria-expanded', 'false');
    }
  }
  
  const init = () => {
    // Initialize dropdowns
    const notifications = new Dropdown('notificationsToggle', 'notificationsDropdown');
    const profile = new Dropdown('profileToggle', 'profileDropdown');
    
    // Close dropdowns on outside click
    document.addEventListener('click', (e) => {
      if (!e.target.closest('[role="menu"]') && !e.target.closest('button[aria-haspopup]')) {
        notifications.close();
        profile.close();
      }
    });
    
    // Sidebar helpers
    const mobileSidebarToggle = document.getElementById('mobileSidebarToggle');
    const sidebar = document.getElementById('sidebar');
    const sidebarOverlay = document.getElementById('sidebarOverlay');
    const desktopSidebarToggle = document.getElementById('desktopSidebarToggle');
    const mainContent = document.getElementById('mainContent');
    
    const openMobileSidebar = () => {
      if (!sidebar || !sidebarOverlay) return;
      sidebar.classList.remove('-translate-x-full');
      sidebarOverlay.classList.remove('hidden');
      document.body.classList.add('sidebar-open');
      document.body.style.overflow = 'hidden';
      if (mobileSidebarToggle) mobileSidebarToggle.setAttribute('aria-expanded', 'true');
    };
    
    const closeMobileSidebar = () => {
      if (!sidebar || !sidebarOverlay) return;
      sidebar.classList.add('-translate-x-full');
      sidebarOverlay.classList.add('hidden');
      document.body.classList.remove('sidebar-open');
      document.body.style.overflow = '';
      if (mobileSidebarToggle) mobileSidebarToggle.setAttribute('aria-expanded', 'false');
    };
    
    const setDesktopCollapsed = (collapsed) => {
      document.body.classList.toggle('sidebar-collapsed', collapsed);
      if (desktopSidebarToggle) {
        desktopSidebarToggle.setAttribute('aria-expanded', collapsed ? 'false' : 'true');
        const chevron = desktopSidebarToggle.querySelector('i');
        if (chevron) chevron.style.transform = collapsed ? 'rotate(180deg)' : 'rotate(0deg)';
      }
      
      try {
        localStorage.setItem('sidebarCollapsed', String(collapsed));
      } catch (e) {
        console.warn('localStorage not available');
      }
    };
    
    if (mobileSidebarToggle && sidebar && sidebarOverlay) {
      mobileSidebarToggle.addEventListener('click', () => {
        if (sidebar.classList.contains('-translate-x-full')) {
          openMobileSidebar();
        } else {
          closeMobileSidebar();
        }
      });
      
      sidebarOverlay.addEventListener('click', () => {
        closeMobileSidebar();
      });
      
      // Close on escape
      document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && document.body.classList.contains('sidebar-open')) {
          closeMobileSidebar();
        }
      });
    }
    
    // Desktop sidebar toggle
    if (desktopSidebarToggle && mainContent) {
      desktopSidebarToggle.addEventListener('click', () => {
        const isCollapsed = document.body.classList.toggle('sidebar-collapsed');
        setDesktopCollapsed(isCollapsed);
      });
      
      // Restore sidebar state
      try {
        const isCollapsed = localStorage.getItem('sidebarCollapsed') === 'true';
        if (isCollapsed) {
          setDesktopCollapsed(true);
        }
      } catch (e) {
        console.warn('localStorage not available');
      }
    }
    
    // Close mobile sidebar on navigation click
    document.querySelectorAll('#sidebar a').forEach(link => {
      link.addEventListener('click', () => {
        if (window.innerWidth < 1024) {
          closeMobileSidebar();
        }
      });
    });
    
    // Reset mobile state on resize
    window.addEventListener('resize', () => {
      if (window.innerWidth >= 1024) {
        closeMobileSidebar();
      }
    });
    
    // Logout confirmation
    const logoutBtn = document.getElementById('logoutBtn');
    if (logoutBtn) {
      logoutBtn.addEventListener('click', (e) => {
        if (!confirm('Are you sure you want to sign out?')) {
          e.preventDefault();
        }
      });
    }
    
    // Mark all notifications as read
    const markAllReadBtn = document.getElementById('markAllReadBtn');
    if (markAllReadBtn) {
      markAllReadBtn.addEventListener('click', async () => {
        try {
          const response = await fetch('/api/notifications/mark-all-read', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' }
          });
          
          if (response.ok) {
            // Remove badge
            const badge = document.querySelector('#notificationsToggle span');
            if (badge) badge.remove();
            
            // Remove unread styles
            document.querySelectorAll('.notification-item.bg-blue-50\\/50').forEach(item => {
              item.classList.remove('bg-blue-50/50');
              const dot = item.querySelector('span.bg-blue-600');
              if (dot) dot.remove();
            });
            
            markAllReadBtn.classList.add('hidden');
            notifications.close();
          }
        } catch (error) {
          console.error('Error:', error);
        }
      });
    }
    
    // Mark individual notification as read
    document.querySelectorAll('.notification-item').forEach(item => {
      item.addEventListener('click', async function() {
        const id = this.dataset.notificationId;
        if (!id) return;
        
        try {
          const response = await fetch('/api/notifications/mark-read', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ id })
          });
          
          if (response.ok) {
            this.classList.remove('bg-blue-50/50');
            const dot = this.querySelector('span.bg-blue-600');
            if (dot) dot.remove();
            
            // Update badge count
            const badge = document.querySelector('#notificationsToggle span');
            if (badge) {
              const count = Math.max(0, (parseInt(badge.textContent) || 0) - 1);
              if (count > 0) {
                badge.textContent = count;
              } else {
                badge.remove();
              }
            }
          }
        } catch (error) {
          console.error('Error:', error);
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