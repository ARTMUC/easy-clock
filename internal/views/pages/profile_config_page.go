package pages

import (
	"fmt"
	"html"
	"io"
	"sort"

	"github.com/a-h/templ"

	"easy-clock/internal/domain"
	"easy-clock/internal/i18n"
)

func ProfileConfigPage(profile *domain.Profile, lang i18n.Lang) templ.Component {
	t := func(k string) string { return i18n.Msg(k, lang) }
	return basePage(profile.Name+" — KidClock", func(w io.Writer) error {
		p := profile
		pid := html.EscapeString(p.ID)
		ring1 := filterRing(p.Activities, 1)
		ring2 := filterRing(p.Activities, 2)

		// ---- header ----
		fmt.Fprint(w, `<div class="space-y-6">`)
		fmt.Fprintf(w, `<div class="bg-white rounded-xl p-4 shadow-sm" style="border-left:4px solid %s">`,
			html.EscapeString(p.Color))
		fmt.Fprint(w, `<div class="flex items-center justify-between">`)
		fmt.Fprintf(w, `<div><h1 class="text-xl font-bold text-gray-800">%s</h1>`, html.EscapeString(p.Name))
		fmt.Fprintf(w, `<p class="text-xs text-gray-500">%d %s</p></div>`,
			len(p.Activities), html.EscapeString(t(i18n.MsgActivities)))
		fmt.Fprintf(w, `<a href="/children/%s" class="btn btn-ghost btn-sm">%s</a>`,
			html.EscapeString(p.ChildID), html.EscapeString(t(i18n.MsgBackToChild)))
		fmt.Fprint(w, `</div></div>`)

		// ---- activity rings ----
		fmt.Fprint(w, `<div class="bg-white rounded-xl p-4 shadow-sm">`)
		fmt.Fprintf(w, `<h2 class="font-semibold text-gray-800 mb-3">%s</h2>`, html.EscapeString(t(i18n.MsgActivities)))
		writeActivityRing(w, t(i18n.MsgRing1), ring1, p.ID, t(i18n.MsgDeleteConfirm))
		writeActivityRing(w, t(i18n.MsgRing2), ring2, p.ID, t(i18n.MsgDeleteConfirm))

		// ---- add activity — tabs ----
		fmt.Fprint(w, `<div class="border-t border-gray-100 pt-4 mt-4">`)
		fmt.Fprintf(w, `<h3 class="font-medium text-gray-700 mb-3">%s</h3>`, html.EscapeString(t(i18n.MsgAddActivity)))

		// tab buttons
		fmt.Fprint(w, `<style>.tab-active{background:#fff;box-shadow:0 1px 3px rgba(0,0,0,.12)}</style>`)
		fmt.Fprintf(w, `<div class="flex gap-1 mb-4 bg-gray-100 rounded-xl p-1">
  <button type="button" id="tab-btn-preset" class="flex-1 py-1.5 rounded-xl text-sm font-medium tab-active" onclick="switchTab('preset')">📋 %s</button>
  <button type="button" id="tab-btn-custom" class="flex-1 py-1.5 rounded-xl text-sm font-medium" onclick="switchTab('custom')">✏️ %s</button>
</div>`, html.EscapeString(t(i18n.MsgPreset)), html.EscapeString(t(i18n.MsgCustom)))

		// ---- preset tab ----
		fmt.Fprint(w, `<div id="tab-preset">`)
		fmt.Fprint(w, `<div id="preset-grid" class="grid grid-cols-3 gap-2 mb-3 min-h-16">
  <p class="text-gray-400 text-sm col-span-3">Loading…</p></div>`)
		fmt.Fprintf(w, `<p id="preset-selected" class="text-sm text-indigo-600 font-medium mb-2 min-h-5"></p>`)
		fmt.Fprintf(w, `<form method="POST" action="/config/profiles/%s/activities" id="preset-form">`, pid)
		fmt.Fprint(w, `<input type="hidden" name="preset_id" id="pf-preset-id">
<input type="hidden" name="emoji"    id="pf-emoji">
<input type="hidden" name="label"    id="pf-label">
<input type="hidden" name="image_path" id="pf-image-path">
<div class="flex gap-2">`)
		fmt.Fprintf(w, `<div class="flex-1"><label>%s</label><input type="number" name="from_hour" min="0" max="23" value="8" required></div>`,
			html.EscapeString(t(i18n.MsgLabelFromHour)))
		fmt.Fprintf(w, `<div class="flex-1"><label>%s</label><input type="number" name="to_hour" min="1" max="24" value="9" required></div>`,
			html.EscapeString(t(i18n.MsgLabelToHour)))
		fmt.Fprintf(w, `<div class="flex-1"><label>%s</label><select name="ring" id="pf-ring"><option value="1">1 (AM)</option><option value="2">2 (PM)</option></select></div>`,
			html.EscapeString(t(i18n.MsgLabelRing)))
		fmt.Fprint(w, `</div>`)
		fmt.Fprintf(w, `<button type="submit" id="preset-submit" disabled class="btn btn-primary w-full mt-3">%s</button>`,
			html.EscapeString(t(i18n.MsgAddActivity)))
		fmt.Fprint(w, `</form></div>`) // end preset tab

		// ---- custom tab ----
		fmt.Fprint(w, `<div id="tab-custom" style="display:none">`)
		fmt.Fprintf(w, `<form method="POST" action="/config/profiles/%s/activities" class="space-y-3">`, pid)
		fmt.Fprint(w, `<div class="flex gap-2">`)
		fmt.Fprintf(w, `<div style="width:5rem;flex-shrink:0"><label>%s</label><input type="text" name="emoji" placeholder="🌙" maxlength="4"></div>`,
			html.EscapeString(t(i18n.MsgLabelEmoji)))
		fmt.Fprintf(w, `<div class="flex-1"><label>%s</label><input type="text" name="label" required></div>`,
			html.EscapeString(t(i18n.MsgLabelLabel)))
		fmt.Fprint(w, `</div>`)
		// image upload
		fmt.Fprintf(w, `<div>
  <label>%s</label>
  <div class="flex items-center gap-2 mt-1">
    <button type="button" class="btn btn-ghost btn-sm" onclick="document.getElementById('file-input').click()">📁 %s</button>
    <span id="upload-status" class="text-sm text-gray-500"></span>
  </div>
  <input type="file" id="file-input" accept="image/*" style="display:none">
  <input type="hidden" name="image_path" id="custom-image-path" required>
</div>`, html.EscapeString(t(i18n.MsgLabelImagePath)), html.EscapeString(t(i18n.MsgUploadImage)))
		fmt.Fprint(w, `<div class="flex gap-2">`)
		fmt.Fprintf(w, `<div class="flex-1"><label>%s</label><input type="number" name="from_hour" min="0" max="23" value="8" required></div>`,
			html.EscapeString(t(i18n.MsgLabelFromHour)))
		fmt.Fprintf(w, `<div class="flex-1"><label>%s</label><input type="number" name="to_hour" min="1" max="24" value="9" required></div>`,
			html.EscapeString(t(i18n.MsgLabelToHour)))
		fmt.Fprintf(w, `<div class="flex-1"><label>%s</label><select name="ring"><option value="1">1 (AM)</option><option value="2">2 (PM)</option></select></div>`,
			html.EscapeString(t(i18n.MsgLabelRing)))
		fmt.Fprint(w, `</div>`)
		fmt.Fprintf(w, `<div><label>%s</label><input type="number" name="sort_order" value="0"></div>`,
			html.EscapeString(t(i18n.MsgLabelSortOrder)))
		fmt.Fprintf(w, `<button type="submit" class="btn btn-primary w-full">%s</button>`,
			html.EscapeString(t(i18n.MsgAddActivity)))
		fmt.Fprint(w, `</form></div>`) // end custom tab

		fmt.Fprint(w, `</div></div></div>`) // end add-activity, card, space-y-6

		// ---- inject JS const ----
		fmt.Fprintf(w, "<script>const PROFILE_ID=%q;</script>\n", p.ID)

		// ---- tab + upload JS (no % fmt issues) ----
		_, err := io.WriteString(w, profileConfigScript)
		return err
	})
}

const profileConfigScript = `<script>
function switchTab(name) {
  ['preset','custom'].forEach(function(n) {
    document.getElementById('tab-'+n).style.display = n === name ? '' : 'none';
    document.getElementById('tab-btn-'+n).classList.toggle('tab-active', n === name);
  });
}

var presetsLoaded = false;
function loadPresets() {
  if (presetsLoaded) return;
  presetsLoaded = true;
  var grid = document.getElementById('preset-grid');
  fetch('/api/preset-activities')
    .then(function(r) { return r.json(); })
    .then(function(presets) {
      grid.innerHTML = '';
      if (!presets || presets.length === 0) {
        grid.innerHTML = '<p class="text-gray-400 text-sm col-span-3">No presets available.</p>';
        return;
      }
      presets.forEach(function(p) {
        var btn = document.createElement('button');
        btn.type = 'button';
        btn.className = 'flex flex-col items-center gap-1 p-2 bg-gray-50 rounded-xl text-sm border border-transparent hover:border-indigo-300 hover:bg-indigo-50 transition-colors';
        var emoji = document.createElement('span');
        emoji.className = 'text-2xl';
        emoji.textContent = p.Emoji || '';
        var label = document.createElement('span');
        label.className = 'text-xs text-gray-600 truncate w-full text-center';
        label.textContent = p.Label || '';
        btn.appendChild(emoji);
        btn.appendChild(label);
        btn.addEventListener('click', function() { selectPreset(p, btn); });
        grid.appendChild(btn);
      });
    })
    .catch(function() {
      grid.innerHTML = '<p class="text-red-500 text-sm col-span-3">Failed to load presets.</p>';
    });
}

function selectPreset(p, btn) {
  document.getElementById('pf-preset-id').value  = p.ID        || '';
  document.getElementById('pf-emoji').value       = p.Emoji     || '';
  document.getElementById('pf-label').value       = p.Label     || '';
  document.getElementById('pf-image-path').value  = p.ImagePath || '';
  document.getElementById('pf-ring').value        = p.Ring      || 1;
  document.getElementById('preset-selected').textContent = (p.Emoji || '') + ' ' + (p.Label || '');
  document.getElementById('preset-submit').disabled = false;
  document.querySelectorAll('#preset-grid button').forEach(function(b) {
    b.classList.remove('border-indigo-400','bg-indigo-50');
  });
  btn.classList.add('border-indigo-400','bg-indigo-50');
}

document.getElementById('file-input').addEventListener('change', function() {
  var file = this.files[0];
  if (!file) return;
  var status = document.getElementById('upload-status');
  status.style.color = '';
  status.textContent = 'Uploading…';
  var fd = new FormData();
  fd.append('file', file);
  fetch('/config/upload', {method: 'POST', body: fd})
    .then(function(r) {
      if (!r.ok) throw new Error('HTTP ' + r.status);
      return r.json();
    })
    .then(function(d) {
      document.getElementById('custom-image-path').value = d.image_path;
      status.style.color = '#16a34a';
      status.textContent = '✓ ' + file.name;
    })
    .catch(function() {
      status.style.color = '#b91c1c';
      status.textContent = '✗ Upload failed';
    });
});

loadPresets();
</script>`

func filterRing(activities []domain.Activity, ring int) []domain.Activity {
	var out []domain.Activity
	for _, a := range activities {
		if a.Ring == ring {
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].FromHour < out[j].FromHour })
	return out
}

func writeActivityRing(w io.Writer, title string, activities []domain.Activity, profileID, deleteConfirm string) {
	fmt.Fprintf(w, `<h3 class="text-sm font-medium text-gray-600 mb-2 mt-3">%s</h3>`, html.EscapeString(title))
	if len(activities) == 0 {
		return
	}
	fmt.Fprint(w, `<div class="space-y-1 mb-2">`)
	for _, a := range activities {
		fmt.Fprint(w, `<div class="flex items-center justify-between bg-gray-50 rounded-xl px-3 py-2">`)
		fmt.Fprintf(w, `<span class="text-sm">%s %s <span class="text-gray-400 text-xs">%02d:00–%02d:00</span></span>`,
			html.EscapeString(a.Emoji), html.EscapeString(a.Label), a.FromHour, a.ToHour)
		fmt.Fprintf(w, `<form method="POST" action="/config/activities/%s/delete" onsubmit="return confirm('%s')">`,
			html.EscapeString(a.ID), html.EscapeString(deleteConfirm))
		fmt.Fprintf(w, `<input type="hidden" name="profile_id" value="%s">`, html.EscapeString(profileID))
		fmt.Fprint(w, `<button type="submit" class="btn btn-danger btn-sm">×</button></form>`)
		fmt.Fprint(w, `</div>`)
	}
	fmt.Fprint(w, `</div>`)
}
