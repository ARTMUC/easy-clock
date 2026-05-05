package pages

import (
	"fmt"
	"html"
	"io"

	"github.com/a-h/templ"

	"easy-clock/internal/domain"
	"easy-clock/internal/i18n"
)

type ChildConfigData struct {
	Child       *domain.Child
	Profiles    []domain.Profile
	Assignments []domain.DayAssignment
	Events      []domain.Event
}

func ChildConfigPage(d ChildConfigData, lang i18n.Lang) templ.Component {
	t := func(k string) string { return i18n.Msg(k, lang) }
	return basePage(d.Child.Name+" — KidClock", func(w io.Writer) error {
		c := d.Child
		fmt.Fprint(w, `<div class="space-y-6">`)

		// ---- header ----
		fmt.Fprint(w, `<div class="bg-white rounded-xl p-4 shadow-sm">`)
		fmt.Fprint(w, `<div class="flex items-center justify-between mb-3">`)

		// avatar + name
		fmt.Fprint(w, `<div class="flex items-center gap-3">`)
		if c.AvatarPath != "" {
			fmt.Fprintf(w, `<img src="%s" class="w-12 h-12 rounded-full object-cover">`, html.EscapeString(c.AvatarPath))
		} else {
			fmt.Fprint(w, `<div class="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center text-2xl">👤</div>`)
		}
		fmt.Fprintf(w, `<div><h1 class="text-xl font-bold text-gray-800">%s</h1><p class="text-xs text-gray-500">%s</p></div>`,
			html.EscapeString(c.Name), html.EscapeString(c.Timezone))
		fmt.Fprint(w, `</div>`)

		fmt.Fprintf(w, `<a href="/clock/%s" target="_blank" class="btn btn-ghost btn-sm">%s</a>`,
			html.EscapeString(c.ClockToken), html.EscapeString(t(i18n.MsgClock)))
		fmt.Fprint(w, `</div>`)

		// avatar upload
		fmt.Fprint(w, `<div class="flex items-center gap-2 mb-3">`)
		fmt.Fprintf(w, `<button type="button" class="btn btn-ghost btn-sm" onclick="document.getElementById('avatar-input').click()">📁 %s</button>`,
			html.EscapeString(t(i18n.MsgUploadAvatar)))
		fmt.Fprint(w, `<span id="avatar-status" class="text-xs text-gray-500"></span>`)
		fmt.Fprintf(w, `<input type="file" id="avatar-input" accept="image/*" style="display:none" data-child="%s">`,
			html.EscapeString(c.ID))
		fmt.Fprint(w, `</div>`)

		// delete child
		fmt.Fprintf(w, `<form method="POST" action="/config/children/%s/delete" onsubmit="return confirm('%s')">`,
			html.EscapeString(c.ID), html.EscapeString(t(i18n.MsgDeleteChildConfirm)))
		fmt.Fprintf(w, `<button type="submit" class="btn btn-danger btn-sm">%s</button></form>`,
			html.EscapeString(t(i18n.MsgDeleteChild)))
		fmt.Fprint(w, `</div>`)

		// ---- profiles ----
		fmt.Fprint(w, `<div class="bg-white rounded-xl p-4 shadow-sm">`)
		fmt.Fprintf(w, `<h2 class="font-semibold text-gray-800 mb-3">%s</h2>`, html.EscapeString(t(i18n.MsgProfiles)))
		if len(d.Profiles) == 0 {
			fmt.Fprintf(w, `<p class="text-sm text-gray-500 mb-3">%s</p>`, html.EscapeString(t(i18n.MsgNoProfiles)))
		} else {
			fmt.Fprint(w, `<div class="space-y-2 mb-4">`)
			for _, p := range d.Profiles {
				isDefault := p.ID == c.DefaultProfileID
				fmt.Fprint(w, `<div class="flex items-center justify-between border border-gray-100 rounded-xl px-3 py-2">`)
				fmt.Fprint(w, `<div class="flex items-center gap-2">`)
				fmt.Fprintf(w, `<span style="width:12px;height:12px;border-radius:50%%;background:%s;display:inline-block"></span>`,
					html.EscapeString(p.Color))
				fmt.Fprintf(w, `<a href="/profiles/%s" class="font-medium text-gray-800 hover:text-indigo-600">%s</a>`,
					html.EscapeString(p.ID), html.EscapeString(p.Name))
				if isDefault {
					fmt.Fprintf(w, `<span class="text-xs text-gray-400">%s</span>`, html.EscapeString(t(i18n.MsgDefaultProfile)))
				}
				fmt.Fprint(w, `</div><div class="flex gap-2">`)
				if !isDefault {
					fmt.Fprintf(w, `<form method="POST" action="/config/children/%s/default-profile">`, html.EscapeString(c.ID))
					fmt.Fprintf(w, `<input type="hidden" name="profile_id" value="%s">`, html.EscapeString(p.ID))
					fmt.Fprintf(w, `<button type="submit" class="btn btn-ghost btn-sm">%s</button></form>`,
						html.EscapeString(t(i18n.MsgSetDefault)))
				}
				fmt.Fprintf(w, `<form method="POST" action="/config/profiles/%s/delete" onsubmit="return confirm('%s')">`,
					html.EscapeString(p.ID), html.EscapeString(t(i18n.MsgDeleteConfirm)))
				fmt.Fprintf(w, `<input type="hidden" name="child_id" value="%s">`, html.EscapeString(c.ID))
				fmt.Fprint(w, `<button type="submit" class="btn btn-danger btn-sm">×</button></form>`)
				fmt.Fprint(w, `</div></div>`)
			}
			fmt.Fprint(w, `</div>`)
		}
		fmt.Fprintf(w, `<form method="POST" action="/config/children/%s/profiles" class="space-y-2">`, html.EscapeString(c.ID))
		fmt.Fprint(w, `<div class="flex gap-2">`)
		fmt.Fprintf(w, `<div class="flex-1"><input type="text" name="name" placeholder="%s" required></div>`,
			html.EscapeString(t(i18n.MsgProfileName)))
		fmt.Fprint(w, `<input type="color" name="color" value="#6366f1" style="width:3rem;flex-shrink:0">`)
		fmt.Fprint(w, `</div>`)
		fmt.Fprintf(w, `<button type="submit" class="btn btn-primary w-full">%s</button>`,
			html.EscapeString(t(i18n.MsgAddProfile)))
		fmt.Fprint(w, `</form></div>`)

		// ---- schedule ----
		fmt.Fprint(w, `<div class="bg-white rounded-xl p-4 shadow-sm">`)
		fmt.Fprintf(w, `<h2 class="font-semibold text-gray-800 mb-3">%s</h2>`, html.EscapeString(t(i18n.MsgWeeklySchedule)))

		assigned := make(map[int]string)
		for _, a := range d.Assignments {
			assigned[a.DayOfWeek] = a.ProfileID
		}

		fmt.Fprint(w, `<div class="space-y-2">`)
		for day := 0; day < 7; day++ {
			currentProfileID := assigned[day]
			fmt.Fprintf(w, `<form method="POST" action="/config/children/%s/schedule/%d" class="flex items-center gap-2">`,
				html.EscapeString(c.ID), day)
			fmt.Fprintf(w, `<span class="text-sm text-gray-600 w-8 shrink-0">%s</span>`,
				html.EscapeString(i18n.DayName(day, lang)))
			fmt.Fprint(w, `<select name="profile_id" class="flex-1">`)
			noneSelected := ""
			if currentProfileID == "" {
				noneSelected = ` selected`
			}
			fmt.Fprintf(w, `<option value=""%s>%s</option>`, noneSelected, html.EscapeString(t(i18n.MsgDayNone)))
			for _, p := range d.Profiles {
				sel := ""
				if p.ID == currentProfileID {
					sel = ` selected`
				}
				fmt.Fprintf(w, `<option value="%s"%s>%s</option>`,
					html.EscapeString(p.ID), sel, html.EscapeString(p.Name))
			}
			fmt.Fprint(w, `</select>`)
			fmt.Fprintf(w, `<button type="submit" class="btn btn-primary btn-sm shrink-0">%s</button>`,
				html.EscapeString(t(i18n.MsgSave)))
			fmt.Fprint(w, `</form>`)
		}
		fmt.Fprint(w, `</div></div>`)

		// ---- events ----
		fmt.Fprint(w, `<div class="bg-white rounded-xl p-4 shadow-sm">`)
		fmt.Fprintf(w, `<h2 class="font-semibold text-gray-800 mb-3">%s</h2>`, html.EscapeString(t(i18n.MsgEvents)))

		if len(d.Events) == 0 {
			fmt.Fprintf(w, `<p class="text-sm text-gray-500 mb-3">%s</p>`, html.EscapeString(t(i18n.MsgNoEvents)))
		} else {
			fmt.Fprint(w, `<div class="space-y-1 mb-4">`)
			for _, e := range d.Events {
				fmt.Fprint(w, `<div class="flex items-center justify-between bg-gray-50 rounded-xl px-3 py-2">`)
				label := e.Label
				if e.Emoji != "" {
					label = e.Emoji + " " + label
				}
				fmt.Fprintf(w, `<span class="text-sm text-gray-800">%s</span>`, html.EscapeString(label))
				fmt.Fprintf(w, `<span class="text-xs text-gray-400 mx-2">%s %s–%s</span>`,
					html.EscapeString(e.Date.Format("2006-01-02")),
					html.EscapeString(e.FromTime[:5]),
					html.EscapeString(e.ToTime[:5]))
				fmt.Fprintf(w, `<form method="POST" action="/config/events/%s/delete" onsubmit="return confirm('%s')">`,
					html.EscapeString(e.ID), html.EscapeString(t(i18n.MsgDeleteConfirm)))
				fmt.Fprintf(w, `<input type="hidden" name="child_id" value="%s">`, html.EscapeString(c.ID))
				fmt.Fprint(w, `<button type="submit" class="btn btn-danger btn-sm">×</button></form>`)
				fmt.Fprint(w, `</div>`)
			}
			fmt.Fprint(w, `</div>`)
		}

		// add event form
		fmt.Fprintf(w, `<form method="POST" action="/config/children/%s/events" class="space-y-3">`, html.EscapeString(c.ID))
		fmt.Fprint(w, `<div class="flex gap-2">`)
		fmt.Fprintf(w, `<div class="flex-1"><label>%s</label><input type="date" name="date" required></div>`,
			html.EscapeString(t(i18n.MsgLabelDate)))
		fmt.Fprintf(w, `<div class="flex-1"><label>%s</label><input type="time" name="from_time" required></div>`,
			html.EscapeString(t(i18n.MsgLabelFromTime)))
		fmt.Fprintf(w, `<div class="flex-1"><label>%s</label><input type="time" name="to_time" required></div>`,
			html.EscapeString(t(i18n.MsgLabelToTime)))
		fmt.Fprint(w, `</div>`)
		fmt.Fprint(w, `<div class="flex gap-2">`)
		fmt.Fprintf(w, `<div style="width:5rem;flex-shrink:0"><label>%s</label><input type="text" name="emoji" maxlength="4"></div>`,
			html.EscapeString(t(i18n.MsgLabelEmoji)))
		fmt.Fprintf(w, `<div class="flex-1"><label>%s</label><input type="text" name="label" required></div>`,
			html.EscapeString(t(i18n.MsgLabelLabel)))
		fmt.Fprint(w, `</div>`)
		fmt.Fprintf(w, `<div><label>%s</label><select name="profile_id">`,
			html.EscapeString(t(i18n.MsgLabelOptionalProfile)))
		fmt.Fprintf(w, `<option value="">— %s —</option>`, html.EscapeString(t(i18n.MsgDayNone)))
		for _, p := range d.Profiles {
			fmt.Fprintf(w, `<option value="%s">%s</option>`, html.EscapeString(p.ID), html.EscapeString(p.Name))
		}
		fmt.Fprint(w, `</select></div>`)
		fmt.Fprintf(w, `<button type="submit" class="btn btn-primary w-full">%s</button>`,
			html.EscapeString(t(i18n.MsgAddEvent)))
		fmt.Fprint(w, `</form></div>`)

		fmt.Fprintf(w, `<a href="/dashboard" class="btn btn-ghost btn-sm">%s</a>`, html.EscapeString(t(i18n.MsgBackDashboard)))
		fmt.Fprint(w, `</div>`)

		_, err := io.WriteString(w, childConfigScript)
		return err
	})
}

const childConfigScript = `<script>
document.getElementById('avatar-input').addEventListener('change', function() {
  var file = this.files[0];
  if (!file) return;
  var childID = this.getAttribute('data-child');
  var status = document.getElementById('avatar-status');
  status.textContent = 'Uploading…';
  var fd = new FormData();
  fd.append('file', file);
  fetch('/config/upload', {method: 'POST', body: fd})
    .then(function(r) {
      if (!r.ok) throw new Error('HTTP ' + r.status);
      return r.json();
    })
    .then(function(d) {
      return fetch('/config/children/' + childID + '/avatar', {
        method: 'POST',
        headers: {'Content-Type': 'application/x-www-form-urlencoded'},
        body: 'avatar_path=' + encodeURIComponent(d.image_path)
      });
    })
    .then(function(r) {
      if (!r.ok) throw new Error('HTTP ' + r.status);
      status.style.color = '#16a34a';
      status.textContent = '✓ ' + file.name;
      setTimeout(function() { window.location.reload(); }, 800);
    })
    .catch(function() {
      status.style.color = '#b91c1c';
      status.textContent = '✗ Upload failed';
    });
});
</script>`
