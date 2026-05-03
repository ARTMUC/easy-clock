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
		ring1 := filterRing(p.Activities, 1)
		ring2 := filterRing(p.Activities, 2)

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

		fmt.Fprint(w, `<div class="bg-white rounded-xl p-4 shadow-sm">`)
		fmt.Fprintf(w, `<h2 class="font-semibold text-gray-800 mb-3">%s</h2>`, html.EscapeString(t(i18n.MsgActivities)))

		writeActivityRing(w, t(i18n.MsgRing1), ring1, p.ID, t(i18n.MsgDeleteConfirm))
		writeActivityRing(w, t(i18n.MsgRing2), ring2, p.ID, t(i18n.MsgDeleteConfirm))

		fmt.Fprint(w, `<div class="border-t border-gray-100 pt-4 mt-4">`)
		fmt.Fprintf(w, `<h3 class="font-medium text-gray-700 mb-3">%s</h3>`, html.EscapeString(t(i18n.MsgAddActivity)))
		fmt.Fprintf(w, `<form method="POST" action="/config/profiles/%s/activities" class="space-y-3">`,
			html.EscapeString(p.ID))
		fmt.Fprint(w, `<div class="flex gap-2">`)
		fmt.Fprintf(w, `<div style="width:5rem;flex-shrink:0"><label>%s</label><input type="text" name="emoji" placeholder="🌙" maxlength="4"></div>`,
			html.EscapeString(t(i18n.MsgLabelEmoji)))
		fmt.Fprintf(w, `<div class="flex-1"><label>%s</label><input type="text" name="label" required></div>`,
			html.EscapeString(t(i18n.MsgLabelLabel)))
		fmt.Fprint(w, `</div>`)
		fmt.Fprintf(w, `<div><label>%s</label><input type="text" name="image_path" placeholder="/static/uploads/sleep.png" required></div>`,
			html.EscapeString(t(i18n.MsgLabelImagePath)))
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
		fmt.Fprint(w, `</form></div></div></div>`)
		return nil
	})
}

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
