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
    body {
      font-family: 'Nunito', system-ui, sans-serif;
      background: var(--bg); color: #2c2416;
      min-height: 100dvh; display: flex; flex-direction: column;
      align-items: center; justify-content: center;
      padding: 1rem; gap: 0.75rem;
    }
    h1 { font-family: 'Baloo 2', sans-serif; font-size: 1.5rem; font-weight: 700; }
    h1 span { color: var(--accent); }
    canvas { display: block; filter: drop-shadow(0 4px 20px rgba(0,0,0,.12)); touch-action: none; }
    #ampm-badge {
      display: inline-flex; align-items: center; gap: 8px;
      background: var(--surface); border: 1px solid rgba(0,0,0,.08);
      border-radius: 100px; padding: 5px 14px;
      font-family: 'Baloo 2', sans-serif; font-size: 13px; font-weight: 500; color: #8a7a62;
    }
    #ampm-badge .dot { width: 8px; height: 8px; border-radius: 50%; background: var(--accent); transition: background .3s; }
    #activity-label { font-size: 1.05rem; font-weight: 600; min-height: 1.4rem; text-align: center; }
    #empty-msg { color: #8a7a62; font-size: .95rem; display: none; text-align: center; }
    #offline { font-size: .7rem; color: #b91c1c; display: none; letter-spacing: .05em; }
  </style>
</head>
<body>
  <h1>Kid<span>Clock</span></h1>
  <div id="ampm-badge">
    <span class="dot" id="badge-dot"></span>
    <span id="badge-text">...</span>
  </div>
  <canvas id="clockCanvas" width="360" height="360"></canvas>
  <div id="activity-label"></div>
  <div id="empty-msg"></div>
  <div id="offline">⚠ offline</div>`

const clockScript = `
  <script>
  const canvas = document.getElementById('clockCanvas');
  const ctx    = canvas.getContext('2d');
  const CX = 180, CY = 180, R = 175;
  const W_THICK = 69, W_THIN = 28, GAP = 3, NUM_R = R - 16;

  let activities   = {1: [], 2: []};
  let profileColor = '#e07a3a';
  let timeOffset   = 0;
  let isOnline     = false;
  const CACHE_KEY  = 'kidclock_' + TOKEN;

  // ---- cache ----
  function loadCache() {
    try {
      const d = JSON.parse(localStorage.getItem(CACHE_KEY) || 'null');
      if (!d) return;
      activities   = d.activities   || {1: [], 2: []};
      profileColor = d.profileColor || '#e07a3a';
    } catch(e) {}
  }

  function saveCache(s) {
    const a1 = [], a2 = [];
    (s.AllActivities || []).forEach(a => {
      const item = {emoji: a.Emoji || '⭐', label: a.Label || '', from: a.FromHour, to: a.ToHour};
      if (a.Ring === 1) a1.push(item); else if (a.Ring === 2) a2.push(item);
    });
    activities   = {1: a1, 2: a2};
    profileColor = s.ProfileColor || '#e07a3a';
    try { localStorage.setItem(CACHE_KEY, JSON.stringify({activities, profileColor})); } catch(e) {}

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
      const s = await r.json();
      saveCache(s);
      isOnline = true;
    } catch(e) {
      isOnline = false;
    }
    document.getElementById('offline').style.display = isOnline ? 'none' : '';
    draw();
  }

  // ---- drawing helpers ----
  function hsl(h, s, l, a) { return 'hsla(' + h + ',' + s + '%,' + l + '%,' + (a == null ? 1 : a) + ')'; }
  function hourToAngle(h)   { return (h / 12) * Math.PI * 2 - Math.PI / 2; }

  function getRingParams(h) {
    const isAM = h < 12;
    const w1   = isAM ? W_THICK : W_THIN;
    const w2   = isAM ? W_THIN  : W_THICK;
    const ring1Outer = NUM_R - 13;
    const r1   = ring1Outer - w1 / 2;
    const r2   = ring1Outer - w1 - GAP - w2 / 2;
    const centerR = ring1Outer - w1 - GAP - w2 - GAP;
    return { isAM, w1, w2, r1, r2, centerR };
  }

  function drawSpokes(centerR, spokeOuter, isAM) {
    for (let i = 0; i < 12; i++) {
      const major = i % 3 === 0;
      const angle = hourToAngle(i === 0 ? 12 : i);
      ctx.beginPath();
      ctx.moveTo(CX + spokeOuter * Math.cos(angle), CY + spokeOuter * Math.sin(angle));
      ctx.lineTo(CX + (centerR + 1) * Math.cos(angle), CY + (centerR + 1) * Math.sin(angle));
      ctx.strokeStyle = isAM
        ? (major ? hsl(35, 50, 40, .32) : hsl(35, 40, 50, .18))
        : (major ? hsl(220, 50, 75, .32) : hsl(220, 40, 70, .18));
      ctx.lineWidth = major ? 1 : 0.5;
      ctx.lineCap = 'round'; ctx.stroke();
    }
  }

  function drawRingActivities(list, ringR, ringW, opacity, h) {
    const hues = [200, 35, 130, 280, 350, 175, 55, 320];
    list.forEach((a, i) => {
      let f12 = a.from % 12, t12 = a.to % 12;
      if (t12 === 0 && a.to  !== 0) t12 = 12;
      if (f12 === 0 && a.from !== 0) f12 = 12;
      const startA = hourToAngle(f12);
      let endA = hourToAngle(t12);
      if (endA <= startA) endA += Math.PI * 2;
      const active = h >= a.from && h < a.to;

      ctx.beginPath(); ctx.arc(CX, CY, ringR, startA, endA);
      ctx.strokeStyle = hsl(hues[i % hues.length], 65, 60, opacity * (active ? 1 : .35));
      ctx.lineWidth = ringW; ctx.lineCap = 'butt'; ctx.stroke();

      if (ringW < 10) return;
      const mid = (startA + endA) / 2;
      const ex = CX + ringR * Math.cos(mid);
      const ey = CY + ringR * Math.sin(mid);
      const sz = Math.min(ringW * .68, 44);
      ctx.save();
      ctx.globalAlpha = opacity * (active ? 1 : .35);
      ctx.font = sz + 'px serif';
      ctx.textAlign = 'center'; ctx.textBaseline = 'middle';
      ctx.fillText(a.emoji, ex, ey);
      ctx.restore();
    });
  }

  function draw() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    const d = now();
    const h = d.getHours(), m = d.getMinutes();
    const { isAM, w1, w2, r1, r2, centerR } = getRingParams(h);

    document.getElementById('badge-text').textContent = isAM ? 'AM' : 'PM';
    document.getElementById('badge-dot').style.background = isAM ? profileColor : '#4a7abf';

    // background
    ctx.beginPath(); ctx.arc(CX, CY, R, 0, Math.PI * 2);
    ctx.fillStyle = isAM ? hsl(48, 90, 96) : hsl(228, 40, 10); ctx.fill();
    ctx.beginPath(); ctx.arc(CX, CY, R - 1, 0, Math.PI * 2);
    ctx.strokeStyle = isAM ? hsl(45, 55, 75, .5) : hsl(228, 35, 30, .6);
    ctx.lineWidth = 2; ctx.stroke();

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
      const tx = -Math.sin(angle), ty = Math.cos(angle), off = 7;
      ctx.font = (isAM ? '700' : '400') + ' ' + (isAM ? 14 : 8) + 'px Nunito,sans-serif';
      ctx.fillStyle = isAM ? hsl(35, 60, 28) : hsl(40, 25, 60, .38);
      ctx.fillText(p.am, nx - tx * off, ny - ty * off);
      ctx.font = (isAM ? '400' : '700') + ' ' + (isAM ? 8 : 14) + 'px Nunito,sans-serif';
      ctx.fillStyle = isAM ? hsl(220, 45, 50, .38) : hsl(220, 50, 84);
      ctx.fillText(p.pm, nx + tx * off, ny + ty * off);
    });

    // tick marks
    const tickOuter = NUM_R - 12;
    for (let i = 0; i < 12; i++) {
      const a = hourToAngle(i === 0 ? 12 : i), len = i % 3 === 0 ? 9 : 5;
      ctx.beginPath();
      ctx.moveTo(CX + tickOuter * Math.cos(a), CY + tickOuter * Math.sin(a));
      ctx.lineTo(CX + (tickOuter - len) * Math.cos(a), CY + (tickOuter - len) * Math.sin(a));
      ctx.strokeStyle = isAM ? hsl(35, 30, 55, .45) : hsl(220, 25, 65, .4);
      ctx.lineWidth = i % 3 === 0 ? 2 : 1; ctx.lineCap = 'round'; ctx.stroke();
    }

    // separator + spokes
    const sep0 = tickOuter - 10;
    ctx.beginPath(); ctx.arc(CX, CY, sep0, 0, Math.PI * 2);
    ctx.strokeStyle = isAM ? hsl(0, 0, 0, .07) : hsl(0, 0, 100, .07);
    ctx.lineWidth = 1; ctx.stroke();
    drawSpokes(centerR, sep0, isAM);

    // ring 1 (AM)
    drawRingActivities(activities[1], r1, w1, isAM ? 1 : .4, h);
    const sep1 = r1 - w1 / 2 - 1;
    ctx.beginPath(); ctx.arc(CX, CY, sep1, 0, Math.PI * 2);
    ctx.strokeStyle = isAM ? hsl(0, 0, 0, .07) : hsl(0, 0, 100, .07);
    ctx.lineWidth = .5; ctx.stroke();

    // ring 2 (PM)
    drawRingActivities(activities[2], r2, w2, isAM ? .4 : 1, h);
    const sep2 = r2 - w2 / 2 - 1;
    ctx.beginPath(); ctx.arc(CX, CY, sep2, 0, Math.PI * 2);
    ctx.strokeStyle = isAM ? hsl(0, 0, 0, .07) : hsl(0, 0, 100, .07);
    ctx.lineWidth = .5; ctx.stroke();

    // center
    const cr = Math.max(centerR, 8);
    ctx.beginPath(); ctx.arc(CX, CY, cr, 0, Math.PI * 2);
    ctx.fillStyle = isAM ? hsl(48, 60, 93) : hsl(228, 30, 7); ctx.fill();
    ctx.strokeStyle = isAM ? hsl(45, 40, 68, .4) : hsl(228, 30, 38, .4);
    ctx.lineWidth = 1; ctx.stroke();
    if (cr > 8) {
      ctx.font = Math.min(cr * 1.2, 20) + 'px serif';
      ctx.textAlign = 'center'; ctx.textBaseline = 'middle';
      ctx.fillText(isAM ? '☀️' : '🌙', CX, CY);
    }

    // hands
    const minLen  = NUM_R - 10;
    const hourLen = minLen - (W_THICK + W_THIN + GAP) * .55;
    const h12 = (h % 12) + m / 60;
    const hourA = hourToAngle(h12);
    const minA  = (m / 60) * Math.PI * 2 - Math.PI / 2;
    const handClr = isAM ? hsl(30, 55, 22) : hsl(220, 55, 90);
    const minClr  = isAM ? hsl(30, 45, 38) : hsl(220, 40, 75);

    ctx.beginPath(); ctx.moveTo(CX, CY);
    ctx.lineTo(CX + hourLen * Math.cos(hourA), CY + hourLen * Math.sin(hourA));
    ctx.strokeStyle = handClr; ctx.lineWidth = 4; ctx.lineCap = 'round'; ctx.stroke();

    ctx.beginPath(); ctx.moveTo(CX, CY);
    ctx.lineTo(CX + minLen * Math.cos(minA), CY + minLen * Math.sin(minA));
    ctx.strokeStyle = minClr; ctx.lineWidth = 2.5; ctx.lineCap = 'round'; ctx.stroke();

    ctx.beginPath(); ctx.arc(CX, CY, 3.5, 0, Math.PI * 2);
    ctx.fillStyle = handClr; ctx.fill();
  }

  // ---- init ----
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
