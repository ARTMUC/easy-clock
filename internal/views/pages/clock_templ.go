package pages

import (
	"context"
	"fmt"
	"io"

	"github.com/a-h/templ"

	"easy-clock/internal/i18n"
)

type clockPage struct {
	token   string
	noSched string
}

const clockHead = `<!doctype html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>KidClock</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link href="https://fonts.googleapis.com/css2?family=Nunito:wght@400;600;700&family=Baloo+2:wght@500;700&display=swap" rel="stylesheet">
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    :root { --accent: #e07a3a; --bg: #f5f0e8; --surface: #fffdf8; }
    html, body { height: 100dvh; overflow: hidden; }
    body {
      font-family: 'Nunito', system-ui, sans-serif;
      background: var(--bg); color: #2c2416;
      display: flex; flex-direction: column;
    }
    h1 { font-family: 'Baloo 2', sans-serif; font-size: 1.3rem; font-weight: 700; }
    h1 span { color: var(--accent); }
    #ampm-badge {
      display: inline-flex; align-items: center; gap: 6px;
      background: var(--surface); border: 1px solid rgba(0,0,0,.08);
      border-radius: 100px; padding: 3px 10px;
      font-family: 'Baloo 2', sans-serif; font-size: 12px; font-weight: 500; color: #8a7a62;
    }
    #ampm-badge .dot { width: 7px; height: 7px; border-radius: 50%; background: var(--accent); transition: background .3s; }
    main { flex: 1; min-height: 0; position: relative; }
    #logo-overlay {
      position: absolute; top: 0.55rem; left: 0.75rem; z-index: 10;
      display: flex; align-items: center; gap: 0.5rem; pointer-events: none;
    }
    canvas { display: block; touch-action: none; width: 100%; height: 100%; }
    footer { flex-shrink: 0; padding: 0.4rem 0.75rem 0.65rem; }
    #timeline-bar { position: relative; height: 48px; background: rgba(0,0,0,.06); border-radius: 10px; overflow: hidden; }
    #timeline-labels { display: flex; justify-content: space-between; font-size: 11px; font-weight: 600; color: #6a5a42; padding: 4px 0 0; }
    #timeline-labels span { width: calc(100% / 24); text-align: center; flex-shrink: 0; }
    #activity-label { font-size: 1rem; font-weight: 600; text-align: center; padding-top: 0.3rem; min-height: 1.3rem; }
    #empty-msg { color: #8a7a62; font-size: .9rem; display: none; text-align: center; }
    #offline { font-size: .65rem; color: #b91c1c; display: none; letter-spacing: .05em; text-align: center; }
  </style>
</head>
<body>
  <main>
    <canvas id="clockCanvas"></canvas>
    <div id="logo-overlay">
      <h1>Kid<span>Clock</span></h1>
      <div id="ampm-badge">
        <span class="dot" id="badge-dot"></span>
        <span id="badge-text">...</span>
      </div>
    </div>
  </main>
  <footer>
    <div id="timeline-bar"></div>
    <div id="timeline-labels"><span>0</span><span>1</span><span>2</span><span>3</span><span>4</span><span>5</span><span>6</span><span>7</span><span>8</span><span>9</span><span>10</span><span>11</span><span>12</span><span>13</span><span>14</span><span>15</span><span>16</span><span>17</span><span>18</span><span>19</span><span>20</span><span>21</span><span>22</span><span>23</span><span>24</span></div>
    <div id="activity-label"></div>
    <div id="empty-msg"></div>
    <div id="offline">⚠ offline</div>
  </footer>`

const clockScript = `
  <script>
  const canvas = document.getElementById('clockCanvas');
  const ctx    = canvas.getContext('2d');
  let CX, CY, R, NUM_R, ARC_OUTER, ARC_INNER;

  function sizeCanvas() {
    canvas.width  = canvas.offsetWidth  || 360;
    canvas.height = canvas.offsetHeight || 360;
    CX = canvas.width  / 2;
    CY = canvas.height / 2;
    R  = Math.min(CX, CY) - 4;
    NUM_R     = R - 16;
    ARC_OUTER = R - 26;
    ARC_INNER = Math.round(R * 0.12);
  }

  let activities        = [];
  let profileColor      = '#e07a3a';
  let timeOffset        = 0;   // clock drift vs server (ms)
  let localOffsetMin    = 0;   // child's timezone UTC offset in minutes
  let isOnline          = false;
  const CACHE_KEY  = 'kidclock_' + TOKEN;

  // ---- timeline (full 24h bar below clock) ----
  function updateTimeline(h_now) {
    const bar = document.getElementById('timeline-bar');
    if (!bar) return;
    const hues = [200, 35, 130, 280, 350, 175, 55, 320];
    let html = '';
    activities.forEach(function(a, idx) {
      const left  = (a.from / 24) * 100;
      const width = Math.max((a.to - a.from) / 24 * 100, 0.4);
      const hue   = hues[idx % hues.length];
      const past  = a.to <= h_now;
      const active = a.from <= h_now && a.to > h_now;
      const alpha = active ? 0.92 : (past ? 0.42 : 0.62);
      html += '<div style="position:absolute;left:'+left+'%;width:'+width+'%;height:100%;' +
              'background:hsla('+hue+',65%,58%,'+alpha+');' +
              'display:flex;align-items:center;justify-content:center;font-size:45px;overflow:hidden;">' +
              a.emoji+'</div>';
    });
    // current time marker
    const ml = (h_now / 24) * 100;
    html += '<div style="position:absolute;left:'+ml+'%;top:0;bottom:0;width:3px;' +
            'background:'+profileColor+';z-index:9;margin-left:-1px;"></div>';
    // ±6h boundary dashes
    [-3, +9].forEach(function(off) {
      const pct = ((h_now + off) / 24 * 100 + 100) % 100;
      html += '<div style="position:absolute;left:'+pct+'%;top:0;bottom:0;width:0;' +
              'border-left:2px dashed rgba(0,0,0,.38);z-index:8;"></div>';
    });
    bar.innerHTML = html;
  }

  // ---- activity arcs (±6 h window, wraps across midnight) ----
  function drawActivityArcs(h_now) {
    const h_start = h_now - 3, h_end = h_now + 9;
    const hues = [200, 35, 130, 280, 350, 175, 55, 320];
    const arcR = (ARC_INNER + ARC_OUTER) / 2; // ~117

    function drawSlice(a, idx, offset) {
      let f = a.from + offset * 24, t = a.to + offset * 24;
      if (f >= h_end || t <= h_start) return;
      f = Math.max(f, h_start);
      t = Math.min(t, h_end);

      const a1 = hourToAngle(f);
      let   a2 = hourToAngle(t);
      if (a2 <= a1) a2 += Math.PI * 2;

      const hue    = hues[idx % hues.length];
      const past   = t <= h_now;
      const active = f <= h_now && t > h_now;
      const alpha  = active ? 0.88 : (past ? 0.38 : 0.60);

      ctx.beginPath();
      ctx.arc(CX, CY, ARC_OUTER, a1, a2);
      ctx.arc(CX, CY, ARC_INNER, a2, a1, true);
      ctx.closePath();
      ctx.fillStyle = hsl(hue, 65, 58, alpha);
      if (active) { ctx.shadowBlur = 14; ctx.shadowColor = hsl(hue, 80, 60); }
      ctx.fill();
      ctx.shadowBlur = 0;

      const eA = hourToAngle((f + t) / 2);
      ctx.save();
      ctx.globalAlpha = past ? 1 : 0.82;
      ctx.font = '78px serif';
      ctx.textAlign = 'center'; ctx.textBaseline = 'middle';
      ctx.fillText(a.emoji, CX + arcR * Math.cos(eA), CY + arcR * Math.sin(eA));
      ctx.restore();
    }

    activities.forEach(function(a, idx) {
      drawSlice(a, idx,  0); // dzisiaj
      drawSlice(a, idx, +1); // jutro (okno przekracza północ)
      drawSlice(a, idx, -1); // wczoraj (okno zaczyna się przed północą)
    });
  }

  // ---- cache ----
  function loadCache() {
    try {
      const d = JSON.parse(localStorage.getItem(CACHE_KEY) || 'null');
      if (!d) return;
      activities     = Array.isArray(d.activities) ? d.activities : [];
      profileColor   = d.profileColor   || '#e07a3a';
      localOffsetMin = d.localOffsetMin || 0;
    } catch(e) {}
  }

  function saveCache(s) {
    activities = (s.AllActivities || []).map(a => ({
      emoji: a.Emoji || '⭐', label: a.Label || '',
      from: a.FromHour, to: a.ToHour
    }));
    profileColor   = s.ProfileColor || '#e07a3a';
    localOffsetMin = (s.LocalOffsetMinutes !== undefined) ? s.LocalOffsetMinutes : localOffsetMin;
    try { localStorage.setItem(CACHE_KEY, JSON.stringify({activities, profileColor, localOffsetMin})); } catch(e) {}

    const label = document.getElementById('activity-label');
    const empty = document.getElementById('empty-msg');
    if (s.Empty) {
      label.textContent = '';
      empty.textContent = NO_SCHED;
      empty.style.display = '';
    } else {
      const a = s.ActiveActivity;
      label.textContent = a ? (a.Emoji || '') + ' ' + (a.Label || '') : '';
      empty.style.display = 'none';
    }
  }

  // ---- time sync ----
  async function syncTime() {
    try {
      const r = await fetch('/api/time', {cache: 'no-store'});
      if (!r.ok) return;
      const d = await r.json();
      timeOffset = d.millis - Date.now();
    } catch(e) {}
  }

  function now() { return new Date(Date.now() + timeOffset); }

  // ---- data refresh ----
  async function refresh() {
    try {
      const r = await fetch('/api/clock/' + TOKEN, {cache: 'no-store'});
      if (!r.ok) throw new Error(r.status);
      saveCache(await r.json());
      isOnline = true;
    } catch(e) { isOnline = false; }
    document.getElementById('offline').style.display = isOnline ? 'none' : '';
    draw();
  }

  // ---- draw helpers ----
  function hsl(h, s, l, a) { return 'hsla('+h+','+s+'%,'+l+'%,'+(a==null?1:a)+')'; }
  function hourToAngle(h)  { return (h / 12) * Math.PI * 2 - Math.PI / 2; }

  // ---- main draw ----
  function draw() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    const d    = now();
    const utcH = d.getUTCHours() + d.getUTCMinutes() / 60;
    const h_now = ((utcH + localOffsetMin / 60) % 24 + 24) % 24;
    const h     = Math.floor(h_now), m = Math.round((h_now - h) * 60);
    const isAM  = h < 12;

    document.getElementById('badge-text').textContent = isAM ? 'AM' : 'PM';
    document.getElementById('badge-dot').style.background = isAM ? profileColor : '#4a7abf';

    // background
    ctx.beginPath(); ctx.arc(CX, CY, R, 0, Math.PI * 2);
    ctx.fillStyle = isAM ? hsl(48,90,96) : hsl(228,40,10); ctx.fill();
    ctx.beginPath(); ctx.arc(CX, CY, R - 1, 0, Math.PI * 2);
    ctx.strokeStyle = isAM ? hsl(45,55,75,.5) : hsl(228,35,30,.6);
    ctx.lineWidth = 2; ctx.stroke();

    updateTimeline(h_now);
    drawActivityArcs(h_now);

    // ±6h boundary: dashed radial line + scissors at clock edge
    [h_now - 3, h_now + 9].forEach(function(hBound) {
      const a   = hourToAngle(hBound);
      const cos = Math.cos(a), sin = Math.sin(a);
      ctx.save();
      ctx.setLineDash([5, 4]);
      ctx.beginPath();
      ctx.moveTo(CX + ARC_INNER * cos, CY + ARC_INNER * sin);
      ctx.lineTo(CX + (NUM_R - 2) * cos, CY + (NUM_R - 2) * sin);
      ctx.strokeStyle = isAM ? 'rgba(0,0,0,.30)' : 'rgba(255,255,255,.25)';
      ctx.lineWidth = 1.5; ctx.lineCap = 'butt'; ctx.stroke();
      ctx.restore();
      ctx.save();
      const sx = CX + (R - 9) * cos, sy = CY + (R - 9) * sin;
      ctx.translate(sx, sy);
      ctx.rotate(Math.PI);
      ctx.font = '90px serif';
      ctx.textAlign = 'center'; ctx.textBaseline = 'middle';
      ctx.globalAlpha = 0.9;
      ctx.fillStyle = '#dc2020';
      ctx.fillText('✂', 0, 0);
      ctx.restore();
    });

    // dual hour numbers
    ctx.textAlign = 'center'; ctx.textBaseline = 'middle';
    const nums = [
      {h12:12,am:'12',pm:'0'},{h12:1,am:'1',pm:'13'},{h12:2,am:'2',pm:'14'},
      {h12:3,am:'3',pm:'15'},{h12:4,am:'4',pm:'16'},{h12:5,am:'5',pm:'17'},
      {h12:6,am:'6',pm:'18'},{h12:7,am:'7',pm:'19'},{h12:8,am:'8',pm:'20'},
      {h12:9,am:'9',pm:'21'},{h12:10,am:'10',pm:'22'},{h12:11,am:'11',pm:'23'}
    ];
    nums.forEach(p => {
      const angle = hourToAngle(p.h12);
      const nx = CX + NUM_R * Math.cos(angle), ny = CY + NUM_R * Math.sin(angle);
      const numSz = Math.round(R * 0.09), off = Math.round(R * 0.055);
      const tx = -Math.sin(angle), ty = Math.cos(angle);
      ctx.font = '700 ' + numSz + 'px Nunito,sans-serif';
      ctx.fillStyle = isAM ? hsl(35,60,28,.9) : hsl(35,45,40,.75);
      ctx.fillText(p.am, nx - tx*off, ny - ty*off);
      ctx.font = '700 ' + numSz + 'px Nunito,sans-serif';
      ctx.fillStyle = isAM ? hsl(220,50,50,.75) : hsl(220,55,80,.9);
      ctx.fillText(p.pm, nx + tx*off, ny + ty*off);
    });

    // tick marks (on top of spiral)
    const tickOuter = NUM_R - 12;
    for (let i = 0; i < 12; i++) {
      const a = hourToAngle(i === 0 ? 12 : i), len = i % 3 === 0 ? 9 : 5;
      ctx.beginPath();
      ctx.moveTo(CX + tickOuter*Math.cos(a), CY + tickOuter*Math.sin(a));
      ctx.lineTo(CX + (tickOuter-len)*Math.cos(a), CY + (tickOuter-len)*Math.sin(a));
      ctx.strokeStyle = isAM ? hsl(35,30,55,.45) : hsl(220,25,65,.4);
      ctx.lineWidth = i%3===0 ? 2 : 1; ctx.lineCap = 'round'; ctx.stroke();
    }

    // center disc
    ctx.beginPath(); ctx.arc(CX, CY, ARC_INNER, 0, Math.PI * 2);
    ctx.fillStyle   = isAM ? hsl(48,60,93) : hsl(228,30,7); ctx.fill();
    ctx.strokeStyle = isAM ? hsl(45,40,68,.4) : hsl(228,30,38,.4);
    ctx.lineWidth = 1; ctx.stroke();

    // hands
    const minLen  = tickOuter - 2;
    const hourLen = minLen * 0.62;
    const h12     = (h % 12) + m / 60;
    const hourA   = hourToAngle(h12);
    const minA    = (m / 60) * Math.PI * 2 - Math.PI / 2;
    const handClr = isAM ? hsl(30,55,22) : hsl(220,55,90);
    const minClr  = isAM ? hsl(30,45,38) : hsl(220,40,75);

    ctx.beginPath(); ctx.moveTo(CX, CY);
    ctx.lineTo(CX + hourLen*Math.cos(hourA), CY + hourLen*Math.sin(hourA));
    ctx.strokeStyle = handClr; ctx.lineWidth = 4; ctx.lineCap = 'round'; ctx.stroke();

    ctx.beginPath(); ctx.moveTo(CX, CY);
    ctx.lineTo(CX + minLen*Math.cos(minA), CY + minLen*Math.sin(minA));
    ctx.strokeStyle = minClr; ctx.lineWidth = 2.5; ctx.lineCap = 'round'; ctx.stroke();

    ctx.beginPath(); ctx.arc(CX, CY, 3.5, 0, Math.PI * 2);
    ctx.fillStyle = handClr; ctx.fill();
  }

  // ---- init ----
  sizeCanvas();
  window.addEventListener('resize', function() { sizeCanvas(); draw(); });
  loadCache();
  draw();
  syncTime().then(refresh);
  setInterval(draw, 1000);
  setInterval(refresh, 60000);
  setInterval(syncTime, 60000);
  </script>
</body>
</html>`

func (p clockPage) Render(_ context.Context, w io.Writer) error {
	if _, err := io.WriteString(w, clockHead); err != nil {
		return err
	}
	// inject Go values as JS consts
	if _, err := fmt.Fprintf(w, "\n  <script>const TOKEN=%q; const NO_SCHED=%q;</script>", p.token, p.noSched); err != nil {
		return err
	}
	_, err := io.WriteString(w, clockScript)
	return err
}

func ClockPage(token string, lang i18n.Lang) templ.Component {
	return clockPage{token: token, noSched: i18n.Msg(i18n.MsgNoSchedule, lang)}
}
