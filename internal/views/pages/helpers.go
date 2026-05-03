package pages

import (
	"context"
	"io"

	"github.com/a-h/templ"

	"easy-clock/internal/views/layout"
)

// context is imported via funcComp — re-export so sub-files can use it inline.
var _ = context.Background

type funcComp func(context.Context, io.Writer) error

func (f funcComp) Render(ctx context.Context, w io.Writer) error { return f(ctx, w) }

func basePage(title string, body func(w io.Writer) error) templ.Component {
	return funcComp(func(ctx context.Context, w io.Writer) error {
		inner := funcComp(func(_ context.Context, w io.Writer) error { return body(w) })
		return layout.Base(title).Render(templ.WithChildren(ctx, inner), w)
	})
}
