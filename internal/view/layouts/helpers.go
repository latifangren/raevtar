package layouts

import (
	"context"
	"io"

	"github.com/a-h/templ"
)

func RawHTML(s string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte(s))
		return err
	})
}
