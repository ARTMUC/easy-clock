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
}

func ChildConfigPage(d ChildConfigData, lang i18n.Lang) templ.Component {
	t := func(k string) string { return i18n.Msg(k, lang) }
	return basePage(d.Child.Name+" — KidClock", func(w io.Writer) error {
		c := d.Child
		fmt.Fprint(w, `<div class="space-y-6">`)

		// header
		fmt.Fprint(w, `<div class="bg-white rounded-xl p-4 shadow-sm">`)
		fmt.Fprint(w, `<div class="flex items-center justify-between mb-2">`)
		fmt.Fprintf(w, `<div><h1 class="text-xl font-bold text-gray-800">%s</h1><p class="text-xs text-gray-500">%s</p></div>`,
			html.EscapeString(c.Name), html.EscapeString(c.Timezone))
		fmt.Fprintf(w, `<a href="/clock/%s" target="_blank" class="btn btn-ghost btn-sm">%s</a>`,
			html.EscapeString(c.ClockToken), html.EscapeString(t(i18n.MsgClock)))
		fmt.Fprint(w, `</div>`)
		fmt.Fprintf(w, `<form method="POST" action="/config/children/%s/delete" onsubmit="return confirm('%s')">`,
			html.EscapeString(c.ID), html.EscapeString(t(i18n.MsgDeleteChildConfirm)))
		fmt.Fprintf(w, `<button type="submit" class="btn btn-danger btn-sm">%s</button></form>`,
			html.EscapeString(t(i18n.MsgDeleteChild)))
		fmt.Fprint(w, `</div>`)

		// profiles
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

		// schedule
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

		fmt.Fprintf(w, `<a href="/dashboard" class="btn btn-ghost btn-sm">%s</a>`, html.EscapeString(t(i18n.MsgBackDashboard)))
		fmt.Fprint(w, `</div>`)
		return nil
	})
}
