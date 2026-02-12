  (() => {
    'use strict';
    
    // Toast Notification System
    const Toast = {
      show(message, type = 'info', duration = 5000) {
        const container = document.getElementById('toastContainer');
        if (!container) return;
        
        const config = {
          success: { bg: 'bg-green-500', icon: 'fa-check-circle' },
          error: { bg: 'bg-red-500', icon: 'fa-exclamation-triangle' },
          warning: { bg: 'bg-amber-500', icon: 'fa-exclamation-circle' },
          info: { bg: 'bg-blue-500', icon: 'fa-info-circle' }
        };
        
        const { bg, icon } = config[type] || config.info;
        
        const toast = document.createElement('div');
        toast.className = `${bg} text-white rounded-lg shadow-lg p-4 max-w-sm transform translate-x-full opacity-0 transition-all duration-300 pointer-events-auto`;
        toast.innerHTML = `
          <div class="flex items-center gap-3">
            <i class="fas ${icon}"></i>
            <span class="flex-1 text-sm font-medium">${this.escapeHtml(message)}</span>
            <button class="text-white/80 hover:text-white transition-colors" aria-label="Close">
              <i class="fas fa-times"></i>
            </button>
          </div>
        `;
        
        // Close button handler
        toast.querySelector('button').addEventListener('click', () => {
          this.remove(toast);
        });
        
        container.appendChild(toast);
        
        // Animate in
        requestAnimationFrame(() => {
          requestAnimationFrame(() => {
            toast.classList.remove('translate-x-full', 'opacity-0');
          });
        });
        
        // Auto remove
        if (duration > 0) {
          setTimeout(() => this.remove(toast), duration);
        }
        
        return toast;
      },
      
      remove(toast) {
        toast.classList.add('translate-x-full', 'opacity-0');
        setTimeout(() => toast.remove(), 300);
      },
      
      escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
      }
    };
    
    // Utility Functions
    const Utils = {
      debounce(func, wait) {
        let timeout;
        return (...args) => {
          clearTimeout(timeout);
          timeout = setTimeout(() => func(...args), wait);
        };
      },
      
      formatDate(date, format = 'medium') {
        const d = new Date(date);
        const options = {
          short: { year: 'numeric', month: 'short', day: 'numeric' },
          medium: { year: 'numeric', month: 'long', day: 'numeric' },
          long: { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' }
        };
        return d.toLocaleDateString('en-US', options[format] || options.medium);
      },
      
      formatTime(date) {
        return new Date(date).toLocaleTimeString('en-US', { 
          hour: '2-digit', 
          minute: '2-digit' 
        });
      },
      
      truncate(text, maxLength = 100) {
        return text.length <= maxLength ? text : text.substring(0, maxLength) + '...';
      }
    };
    
    // Form Utilities
    const Form = {
      showLoading(button, text = 'Processing...') {
        if (!button) return;
        button.dataset.originalHtml = button.innerHTML;
        button.disabled = true;
        button.classList.add('opacity-75', 'cursor-not-allowed');
        button.innerHTML = `<i class="fas fa-spinner fa-spin mr-2"></i>${text}`;
      },
      
      resetLoading(button) {
        if (!button || !button.dataset.originalHtml) return;
        button.innerHTML = button.dataset.originalHtml;
        button.disabled = false;
        button.classList.remove('opacity-75', 'cursor-not-allowed');
        delete button.dataset.originalHtml;
      }
    };
    
    // Initialize on DOM ready
    const init = () => {
      // Auto-dismiss alerts
      document.querySelectorAll('[data-auto-dismiss]').forEach(alert => {
        const delay = parseInt(alert.dataset.delay) || 5000;
        setTimeout(() => {
          alert.style.opacity = '0';
          alert.style.transform = 'scale(0.95)';
          setTimeout(() => alert.remove(), 300);
        }, delay);
      });
      
      // Prevent double form submission
      document.querySelectorAll('form:not([data-no-prevent])').forEach(form => {
        form.addEventListener('submit', function() {
          const submitBtn = this.querySelector('button[type="submit"]');
          if (submitBtn && !submitBtn.disabled) {
            Form.showLoading(submitBtn, 'Submitting...');
            setTimeout(() => Form.resetLoading(submitBtn), 30000);
          }
        });
      });
      
      // Confirm actions
      document.querySelectorAll('[data-confirm]').forEach(element => {
        element.addEventListener('click', function(e) {
          const message = this.dataset.confirm || 'Are you sure?';
          if (!confirm(message)) {
            e.preventDefault();
            e.stopPropagation();
          }
        });
      });
      
      // AJAX form submissions
      document.querySelectorAll('form[data-ajax]').forEach(form => {
        form.addEventListener('submit', async function(e) {
          e.preventDefault();
          
          const submitBtn = this.querySelector('button[type="submit"]');
          if (submitBtn) Form.showLoading(submitBtn);
          
          try {
            const formData = new FormData(this);
            const response = await fetch(this.action, {
              method: this.method,
              body: formData,
              headers: { 'X-Requested-With': 'XMLHttpRequest' }
            });
            
            const data = await response.json();
            
            if (response.ok) {
              Toast.show(data.message || 'Success!', 'success');
              
              if (data.redirect) {
                setTimeout(() => window.location.href = data.redirect, 1000);
              }
              
              if (data.callback && typeof window[data.callback] === 'function') {
                window[data.callback](data);
              }
            } else {
              Toast.show(data.error || 'An error occurred', 'error');
            }
          } catch (error) {
            console.error('Form error:', error);
            Toast.show('An error occurred. Please try again.', 'error');
          } finally {
            if (submitBtn) Form.resetLoading(submitBtn);
          }
        });
      });
      
      // Auto-focus first error
      const firstError = document.querySelector('.border-red-500, input:invalid');
      if (firstError) {
        firstError.focus();
        firstError.scrollIntoView({ behavior: 'smooth', block: 'center' });
      }
    };
    
    // Expose globals
    window.Toast = Toast;
    window.Utils = Utils;
    window.Form = Form;
    window.showToast = Toast.show.bind(Toast);
    window.debounce = Utils.debounce;
    
    // Run init
    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', init);
    } else {
      init();
    }
  })();