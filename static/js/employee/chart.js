(() => {
  let chart;
  let zoomLevel = 0.7;

  const container = document.querySelector('.chart-container');
  if (!container) return;

  const chartApi = container.dataset.chartApi;
  if (!chartApi) return;

  const escapeHtml = (value) => (value || '')
    .toString()
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;');

  const stylesByType = {
    company: {
      badge: 'background:#dbeafe;color:#1d4ed8;',
      avatar: 'background:#dbeafe;color:#1d4ed8;',
      border: '#bfdbfe'
    },
    manager: {
      badge: 'background:#ede9fe;color:#6d28d9;',
      avatar: 'background:#ede9fe;color:#6d28d9;',
      border: '#ddd6fe'
    },
    supervisor: {
      badge: 'background:#dcfce7;color:#166534;',
      avatar: 'background:#dcfce7;color:#166534;',
      border: '#bbf7d0'
    },
    team: {
      badge: 'background:#fef3c7;color:#92400e;',
      avatar: 'background:#fef3c7;color:#92400e;',
      border: '#fde68a'
    },
    member: {
      badge: 'background:#e0f2fe;color:#075985;',
      avatar: 'background:#e0f2fe;color:#075985;',
      border: '#bae6fd'
    }
  };

  const getNodeStyle = (nodeType) => stylesByType[nodeType] || stylesByType.member;

  const renderNode = (node) => {
    const style = getNodeStyle(node.nodeType);
    const nationality = escapeHtml(node.nationality);
    const meta = escapeHtml(node.meta);
    const name = escapeHtml(node.name);
    const position = escapeHtml(node.positionName);
    const avatarText = escapeHtml((node.name || 'U').charAt(0).toUpperCase());

    return `
      <div style="height:100%;width:100%;padding:12px;">
        <div style="height:100%;width:100%;background:#fff;border-radius:18px;box-shadow:0 10px 30px rgba(15,23,42,.08);border:1px solid ${style.border};padding:16px;display:flex;flex-direction:column;align-items:center;justify-content:center;text-align:center;gap:8px;">
          <div style="width:56px;height:56px;border-radius:9999px;display:flex;align-items:center;justify-content:center;font-size:18px;font-weight:700;${style.avatar}">
            ${avatarText}
          </div>
          <div style="font-size:14px;font-weight:700;color:#111827;line-height:1.3;">${name}</div>
          <div style="padding:3px 10px;border-radius:9999px;font-size:10px;font-weight:700;letter-spacing:.02em;${style.badge}">
            ${position}
          </div>
          ${nationality ? `<div style="font-size:11px;color:#4b5563;">${nationality}</div>` : ''}
          ${meta ? `<div style="font-size:10px;color:#9ca3af;">${meta}</div>` : ''}
        </div>
      </div>
    `;
  };

  fetch(chartApi)
    .then((res) => res.json())
    .then((data) => {
      chart = new d3.OrgChart()
        .container('.chart-container')
        .data(data)
        .nodeWidth(() => 260)
        .nodeHeight(() => 190)
        .childrenMargin(() => 40)
        .compactMarginBetween(() => 20)
        .compactMarginPair(() => 80)
        .initialZoom(zoomLevel)
        .nodeContent((d) => renderNode(d.data))
        .render();
    });

  const zoom = (delta) => {
    if (!chart) return;
    zoomLevel = Math.max(0.2, Math.min(2, zoomLevel + delta));
    chart.zoom(zoomLevel);
  };

  const on = (id, handler) => {
    const element = document.getElementById(id);
    if (element) element.onclick = handler;
  };

  on('zoomIn', () => zoom(0.1));
  on('zoomOut', () => zoom(-0.1));
  on('reset', () => {
    if (!chart) return;
    zoomLevel = 0.7;
    chart.zoom(zoomLevel).center();
  });
  on('fit', () => chart && chart.fit());
  on('export', () => {
    if (!chart) return;
    chart.exportImg({
      full: true,
      scale: 2,
      onLoad: (base64) => {
        const a = document.createElement('a');
        a.href = base64;
        a.download = 'org-chart.png';
        a.click();
      }
    });
  });
  on('fullscreen', () => {
    if (!document.fullscreenElement) container.requestFullscreen();
    else document.exitFullscreen();
  });
})();
