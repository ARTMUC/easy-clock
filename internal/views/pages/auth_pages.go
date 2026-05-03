package pages

import (
	"context"
	"fmt"
	"html"
	"io"

	"github.com/a-h/templ"

	"easy-clock/internal/i18n"
)

func LoginPage(errMsg string, lang i18n.Lang) templ.Component {
	t := func(k string) string { return i18n.Msg(k, lang) }
	return funcComp(func(_ context.Context, w io.Writer) error {
		fmt.Fprintf(w, `<!doctype html><html lang="%s"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="theme-color" content="#4f46e5">
<title>%s</title>
<link rel="stylesheet" href="/static/app.css">
</head>
<body class="bg-gray-50 min-h-screen flex items-center justify-center px-4">
<div class="w-full" style="max-width:24rem">
  <div class="text-center mb-8">
    <span class="text-indigo-600 font-bold text-3xl">KidClock</span>
    <p class="text-gray-500 mt-1 text-sm">%s</p>
  </div>
  <div class="bg-white rounded-2xl p-6 shadow-sm space-y-4">`,
			string(lang),
			html.EscapeString(t(i18n.MsgLoginTitle)),
			html.EscapeString(t(i18n.MsgLoginSubtitle)))
		if errMsg != "" {
			fmt.Fprintf(w, `<div class="bg-red-50 border border-red-200 text-red-700 rounded-xl px-4 py-3 text-sm">%s</div>`,
				html.EscapeString(errMsg))
		}
		fmt.Fprintf(w, `<form method="POST" action="/login" class="space-y-4">
  <div><label>%s</label><input type="email" name="email" required autofocus placeholder="you@example.com"></div>
  <div><label>%s</label><input type="password" name="password" required placeholder="••••••••"></div>
  <button type="submit" class="btn btn-primary w-full">%s</button>
</form>
<p class="text-center text-sm text-gray-500">%s
  <a href="/register" class="text-indigo-600 font-medium hover:underline">%s</a>
</p>`,
			html.EscapeString(t(i18n.MsgLabelEmail)),
			html.EscapeString(t(i18n.MsgLabelPassword)),
			html.EscapeString(t(i18n.MsgBtnLogin)),
			html.EscapeString(t(i18n.MsgNoAccount)),
			html.EscapeString(t(i18n.MsgSignUp)))
		fmt.Fprint(w, `</div></div></body></html>`)
		return nil
	})
}

func RegisterPage(errMsg string, lang i18n.Lang) templ.Component {
	t := func(k string) string { return i18n.Msg(k, lang) }
	return funcComp(func(_ context.Context, w io.Writer) error {
		fmt.Fprintf(w, `<!doctype html><html lang="%s"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="theme-color" content="#4f46e5">
<title>%s</title>
<link rel="stylesheet" href="/static/app.css">
</head>
<body class="bg-gray-50 min-h-screen flex items-center justify-center px-4">
<div class="w-full" style="max-width:24rem">
  <div class="text-center mb-8">
    <span class="text-indigo-600 font-bold text-3xl">KidClock</span>
    <p class="text-gray-500 mt-1 text-sm">%s</p>
  </div>
  <div class="bg-white rounded-2xl p-6 shadow-sm space-y-4">`,
			string(lang),
			html.EscapeString(t(i18n.MsgRegisterTitle)),
			html.EscapeString(t(i18n.MsgRegisterSubtitle)))
		if errMsg != "" {
			fmt.Fprintf(w, `<div class="bg-red-50 border border-red-200 text-red-700 rounded-xl px-4 py-3 text-sm">%s</div>`,
				html.EscapeString(errMsg))
		}
		fmt.Fprintf(w, `<form method="POST" action="/register" class="space-y-4">
  <div><label>%s</label><input type="text" name="name" required autofocus placeholder="%s"></div>
  <div><label>%s</label><input type="email" name="email" required placeholder="you@example.com"></div>
  <div><label>%s</label><input type="password" name="password" required placeholder="••••••••" minlength="8"></div>
  <button type="submit" class="btn btn-primary w-full">%s</button>
</form>
<p class="text-center text-sm text-gray-500">%s
  <a href="/login" class="text-indigo-600 font-medium hover:underline">%s</a>
</p>`,
			html.EscapeString(t(i18n.MsgLabelName)),
			html.EscapeString(t(i18n.MsgPlaceholderName)),
			html.EscapeString(t(i18n.MsgLabelEmail)),
			html.EscapeString(t(i18n.MsgLabelPassword)),
			html.EscapeString(t(i18n.MsgBtnCreateAccount)),
			html.EscapeString(t(i18n.MsgHaveAccount)),
			html.EscapeString(t(i18n.MsgBtnLogin)))
		fmt.Fprint(w, `</div></div></body></html>`)
		return nil
	})
}

func VerifyPage(success bool, msg string, lang i18n.Lang) templ.Component {
	return funcComp(func(_ context.Context, w io.Writer) error {
		color := "#dc2626"
		if success {
			color = "#16a34a"
		}
		fmt.Fprintf(w, `<!doctype html><html lang="%s"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>%s</title>
<link rel="stylesheet" href="/static/app.css">
</head>
<body class="bg-gray-50 min-h-screen flex items-center justify-center px-4">
<div class="w-full text-center" style="max-width:24rem">
  <p style="color:%s;font-size:1rem;font-weight:600">%s</p>
  <a href="/login" class="btn btn-primary mt-4 inline-block">%s</a>
</div></body></html>`,
			string(lang),
			html.EscapeString(i18n.Msg(i18n.MsgVerifyTitle, lang)),
			color,
			html.EscapeString(msg),
			html.EscapeString(i18n.Msg(i18n.MsgBtnLogin, lang)))
		return nil
	})
}

func CheckEmailPage(email string, lang i18n.Lang) templ.Component {
	return funcComp(func(_ context.Context, w io.Writer) error {
		fmt.Fprintf(w, `<!doctype html><html lang="%s"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>%s</title>
<link rel="stylesheet" href="/static/app.css">
</head>
<body class="bg-gray-50 min-h-screen flex items-center justify-center px-4">
<div class="w-full text-center" style="max-width:24rem">
  <span class="text-indigo-600 font-bold text-3xl">KidClock</span>
  <p class="text-gray-700 mt-4">%s <strong>%s</strong>.</p>
</div></body></html>`,
			string(lang),
			html.EscapeString(i18n.Msg(i18n.MsgCheckEmailTitle, lang)),
			html.EscapeString(i18n.Msg(i18n.MsgCheckEmailBody, lang)),
			html.EscapeString(email))
		return nil
	})
}
