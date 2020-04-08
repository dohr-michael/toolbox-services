package lib

import (
	"encoding/json"
	"github.com/go-chi/render"
	"net/http"
)

type httpResponse struct {
	Data   interface{}
	Status int
}

func HttpResponse(data interface{}, status int) render.Renderer {
	return &httpResponse{data, status}
}

func (h *httpResponse) Render(w http.ResponseWriter, r *http.Request) error {
	if h.Status != 0 {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(h.Status)
	}
	return nil
}

func (h *httpResponse) MarshalJSON() ([]byte, error) { return json.Marshal(h.Data) }

type httpResponseWithId struct {
	Id     interface{} `json:"id,omitempty"`
	Data   interface{} `json:"data"`
	Status int         `json:"-"`
}

func HttpResponseWithId(id interface{}, data interface{}, status int) render.Renderer {
	return &httpResponseWithId{id, data, status}
}

func (h *httpResponseWithId) Render(w http.ResponseWriter, r *http.Request) error {
	if h.Status != 0 {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(h.Status)
	}
	return nil
}

func toErrorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

type HttpError interface {
	error
	render.Renderer
}

type HttpResponseError struct {
	Err            error `json:"-"`
	HTTPStatusCode int   `json:"-"`

	StatusText string `json:"status"`
	AppCode    int64  `json:"code,omitempty"`
	ErrorText  string `json:"error,omitempty"`
}

func (e *HttpResponseError) Error() string { return e.Err.Error() }
func (e *HttpResponseError) Render(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if e.HTTPStatusCode == 0 {
		render.Status(r, 500)
	} else {
		render.Status(r, e.HTTPStatusCode)
	}
	return nil
}

func HttpNotFound(err error) HttpError {
	return &HttpResponseError{
		Err:            err,
		HTTPStatusCode: http.StatusNotFound,
		StatusText:     http.StatusText(http.StatusNotFound),
		ErrorText:      toErrorString(err),
	}
}

func HttpBadRequest(err error) HttpError {
	return &HttpResponseError{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     http.StatusText(http.StatusBadRequest),
		ErrorText:      toErrorString(err),
	}
}

func HttpUnauthorized(err error) HttpError {
	return &HttpResponseError{
		Err:            err,
		HTTPStatusCode: http.StatusUnauthorized,
		StatusText:     http.StatusText(http.StatusUnauthorized),
		ErrorText:      toErrorString(err),
	}
}

func HttpRenderError(err error) HttpError {
	return &HttpResponseError{
		Err:            err,
		HTTPStatusCode: http.StatusUnprocessableEntity,
		StatusText:     http.StatusText(http.StatusUnprocessableEntity),
		ErrorText:      toErrorString(err),
	}
}

func ToHttpError(err error) HttpError {
	// TODO manage properly this status in function of the error.
	status := http.StatusInternalServerError
	return &HttpResponseError{
		Err:            err,
		HTTPStatusCode: status,
		StatusText:     http.StatusText(status),
		ErrorText:      toErrorString(err),
	}
}
