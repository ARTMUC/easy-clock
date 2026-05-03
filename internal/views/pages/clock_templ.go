package pages

import (
	"context"
	"fmt"
	"html"
	"io"

	"github.com/a-h/templ"
)

type clockPage struct{ token string }

func (p clockPage) Render(_ context.Context, w io.Writer) error {
	t := html.EscapeString(p.token)
	_, err := fmt.Fprintf(w, `<!doctype html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>KidClock</title>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body { background: #0f172a; color: #f1f5f9; font-family: system-ui, sans-serif;
           display: flex; flex-direction: column; align-items: center;
           justify-content: center; min-height: 100dvh; padding: 1rem; }
    #time { font-size: clamp(3rem, 20vw, 8rem); font-weight: 700; letter-spacing: -0.02em; }
    #activity { margin-top: 1rem; font-size: clamp(1.5rem, 8vw, 3rem); text-align: center; }
    #activity-img { max-height: 40vh; border-radius: 1rem; margin-top: 1rem; display: none; }
    #all { display: flex; gap: 0.75rem; flex-wrap: wrap; justify-content: center; margin-top: 2rem; }
    .chip { background: #1e293b; border-radius: 999px; padding: 0.35rem 0.9rem; font-size: 0.9rem; }
    .chip.active { background: #6366f1; font-weight: 600; }
    #empty { color: #64748b; font-size: 1.25rem; }
  </style>
</head>
<body>
  <div id="time">--:--</div>
  <div id="activity"></div>
  <img id="activity-img" src="" alt="">
  <div id="all"></div>
  <div id="empty" style="display:none">No schedule for now.</div>
  <script>
    const TOKEN = %q;
    async function refresh() {
      try {
        const r = await fetch('/api/clock/' + TOKEN);
        if (!r.ok) return;
        const s = await r.json();
        if (s.Empty) {
          document.getElementById('empty').style.display = '';
          document.getElementById('activity').textContent = '';
          document.getElementById('activity-img').style.display = 'none';
          document.getElementById('all').innerHTML = '';
          return;
        }
        document.getElementById('empty').style.display = 'none';
        const a = s.ActiveActivity;
        if (a) {
          document.getElementById('activity').textContent = (a.Emoji || '') + ' ' + (a.Label || '');
          const img = document.getElementById('activity-img');
          if (a.ImagePath) { img.src = a.ImagePath; img.style.display = ''; }
          else { img.style.display = 'none'; }
        } else {
          document.getElementById('activity').textContent = '';
          document.getElementById('activity-img').style.display = 'none';
        }
        const all = document.getElementById('all');
        all.innerHTML = '';
        (s.AllActivities || []).forEach(act => {
          const d = document.createElement('div');
          d.className = 'chip' + (a && act.ID === a.ID ? ' active' : '');
          d.textContent = (act.Emoji || '') + ' ' + act.Label;
          all.appendChild(d);
        });
      } catch(e) { console.error(e); }
    }
    function tick() {
      const now = new Date();
      const h = String(now.getHours()).padStart(2,'0');
      const m = String(now.getMinutes()).padStart(2,'0');
      document.getElementById('time').textContent = h + ':' + m;
    }
    tick(); setInterval(tick, 10000);
    refresh(); setInterval(refresh, 30000);
  </script>
</body>
</html>`, t)
	return err
}

func ClockPage(token string) templ.Component {
	return clockPage{token: token}
}
