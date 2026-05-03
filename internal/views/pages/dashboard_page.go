package pages

import (
	"fmt"
	"html"
	"io"

	"github.com/a-h/templ"

	"easy-clock/internal/domain"
	"easy-clock/internal/i18n"
)

func DashboardPage(children []domain.Child, lang i18n.Lang) templ.Component {
	t := func(k string) string { return i18n.Msg(k, lang) }
	return basePage(t(i18n.MsgChildren)+" — KidClock", func(w io.Writer) error {
		fmt.Fprint(w, `<div class="space-y-6">`)
		fmt.Fprintf(w, `<h1 class="text-xl font-bold text-gray-800">%s</h1>`, html.EscapeString(t(i18n.MsgChildren)))

		if len(children) == 0 {
			fmt.Fprintf(w, `<p class="text-gray-500 text-sm">%s</p>`, html.EscapeString(t(i18n.MsgNoChildren)))
		} else {
			fmt.Fprint(w, `<div class="space-y-3">`)
			for _, c := range children {
				fmt.Fprint(w, `<div class="bg-white rounded-xl p-4 shadow-sm flex items-center justify-between">`)
				fmt.Fprintf(w, `<div><p class="font-semibold text-gray-800">%s</p><p class="text-xs text-gray-500 mt-1">%s</p></div>`,
					html.EscapeString(c.Name), html.EscapeString(c.Timezone))
				fmt.Fprint(w, `<div class="flex gap-2">`)
				fmt.Fprintf(w, `<a href="/clock/%s" target="_blank" class="btn btn-ghost btn-sm">%s</a>`,
					html.EscapeString(c.ClockToken), html.EscapeString(t(i18n.MsgClock)))
				fmt.Fprintf(w, `<a href="/children/%s" class="btn btn-primary btn-sm">%s</a>`,
					html.EscapeString(c.ID), html.EscapeString(t(i18n.MsgConfigure)))
				fmt.Fprint(w, `</div></div>`)
			}
			fmt.Fprint(w, `</div>`)
		}

		fmt.Fprint(w, `<div class="bg-white rounded-xl p-5 shadow-sm">`)
		fmt.Fprintf(w, `<h2 class="font-semibold text-gray-800 mb-4">%s</h2>`, html.EscapeString(t(i18n.MsgAddChild)))
		fmt.Fprint(w, `<form method="POST" action="/config/children" class="space-y-3">`)
		fmt.Fprintf(w, `<div><label>%s</label><input type="text" name="name" placeholder="%s" required></div>`,
			html.EscapeString(t(i18n.MsgLabelName)), html.EscapeString(t(i18n.MsgPlaceholderName)))
		fmt.Fprintf(w, `<div><label>%s</label><input type="text" name="timezone" placeholder="%s" required></div>`,
			html.EscapeString(t(i18n.MsgLabelTimezone)), html.EscapeString(t(i18n.MsgPlaceholderTZ)))
		fmt.Fprintf(w, `<button type="submit" class="btn btn-primary w-full">%s</button>`,
			html.EscapeString(t(i18n.MsgAddChild)))
		fmt.Fprint(w, `</form></div></div>`)
		return nil
	})
}
