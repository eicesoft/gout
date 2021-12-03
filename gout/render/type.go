package render

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

var (
	jsonContentType  = []string{"application/json; charset=utf-8"}
	htmlContentType  = []string{"text/html; charset=utf-8"}
	plainContentType = []string{"text/plain; charset=utf-8"}
	xmlContentType   = []string{"application/xml; charset=utf-8"}
)

type JSON struct {
	Data interface{}
}

func (r JSON) Render(w http.ResponseWriter) (err error) {
	if err = WriteJSON(w, r.Data); err != nil {
		panic(err)
	}
	return
}

func WriteJSON(w http.ResponseWriter, obj interface{}) error {
	writeContentType(w, jsonContentType)
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = w.Write(jsonBytes)
	return err
}

// WriteContentType (JSON) writes JSON ContentType.
func (r JSON) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonContentType)
}

type HTML struct {
	Data string
}

func (r HTML) Render(w http.ResponseWriter) (err error) {
	r.WriteContentType(w)
	_, err = w.Write([]byte(r.Data))
	return err
}

func (r HTML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, htmlContentType)
}

type Raw struct {
	Data        []byte
	ContentType string
}

func (r Raw) Render(w http.ResponseWriter) (err error) {
	r.WriteContentType(w)
	_, err = w.Write(r.Data)
	return
}

func (r Raw) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, []string{r.ContentType})
}

type Text struct {
	Data string
}

func (r Text) Render(w http.ResponseWriter) (err error) {
	r.WriteContentType(w)
	_, err = w.Write([]byte(r.Data))
	return
}

func (r Text) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, plainContentType)
}

type XML struct {
	Data interface{}
}

func (r XML) Render(w http.ResponseWriter) (err error) {
	r.WriteContentType(w)
	err = xml.NewEncoder(w).Encode(r.Data)
	return
}

func (r XML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, xmlContentType)
}
