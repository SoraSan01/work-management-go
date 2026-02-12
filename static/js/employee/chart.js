let chart;
let zoomLevel = 0.7; // define initial zoom

// Fetch org chart data
fetch('/employee/teams/chart/api/orgchart')
  .then(res => res.json())
  .then(data => {
    chart = new d3.OrgChart()
      .container('.chart-container')
      .data(data)
      .nodeWidth(() => 250)
      .nodeHeight(() => 175)
      .childrenMargin(() => 40)
      .compactMarginBetween(() => 15)
      .compactMarginPair(() => 80)
      .initialZoom(zoomLevel)
      .nodeContent(d => `
        <div class="h-full w-full p-3">
          <div class="
            h-full w-full
            bg-white
            rounded-2xl
            shadow-md
            border border-gray-200
            flex flex-col items-center justify-center
            text-center
            hover:shadow-lg
            transition
          ">
            <!-- Avatar -->
            <div class="
              w-14 h-14
              rounded-full
              bg-indigo-100
              text-indigo-600
              flex items-center justify-center
              font-semibold text-lg
              mb-3
            ">
              ${d.data.name?.charAt(0) ?? "U"}
            </div>
            <!-- Name -->
            <div class="font-semibold text-gray-900 text-sm">${d.data.name}</div>
            <!-- Position -->
            <div class="mt-2 px-2 py-0.5 text-[10px] rounded-full bg-gray-100 text-gray-600">
              ${d.data.positionName}
            </div>
          </div>
        </div>
      `)
      .render();
  });

// Zoom helper
const zoom = delta => {
  if (!chart) return;
  zoomLevel = Math.max(0.2, Math.min(2, zoomLevel + delta));
  chart.zoom(zoomLevel);
};

// Toolbar events
document.getElementById('zoomIn').onclick  = () => zoom(0.1);
document.getElementById('zoomOut').onclick = () => zoom(-0.1);
document.getElementById('reset').onclick = () => {
  zoomLevel = 0.7;
  chart.zoom(zoomLevel).center();
};
document.getElementById('fit').onclick = () => chart.fit();

// Export PNG
document.getElementById('export').onclick = () => {
  chart.exportImg({
    full: true,
    scale: 2,
    onLoad: base64 => {
      const a = document.createElement('a');
      a.href = base64;
      a.download = 'org-chart.png';
      a.click();
    }
  });
};

// Fullscreen
document.getElementById('fullscreen').onclick = () => {
  const el = document.querySelector('.chart-container');
  if (!document.fullscreenElement) el.requestFullscreen();
  else document.exitFullscreen();
};