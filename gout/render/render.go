package render

import "net/http"

type Render interface {
	// Render writes data with custom ContentType.
	Render(http.ResponseWriter) error
	// WriteContentType writes custom ContentType.
	WriteContentType(w http.ResponseWriter)
}

var (
	_ Render = JSON{}
	_ Render = HTML{}
	_ Render = Raw{}
	_ Render = Text{}
	_ Render = XML{}
)

func writeContentType(w http.ResponseWriter, value []string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}
