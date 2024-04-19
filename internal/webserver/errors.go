package webserver

import (
	"github.com/danielmichaels/tawny/assets/static/view/pages"
	"github.com/danielmichaels/tawny/internal/render"
	chirender "github.com/go-chi/render"
	"net/http"
)

func (app *Application) notFound(w http.ResponseWriter, r *http.Request) {
	accept := r.Header.Get("Accept")
	switch accept {
	default:
		_ = render.Render(r.Context(), w, http.StatusNotFound, pages.NotFoundErrorPage())
	}
}

func (app *Application) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	accept := r.Header.Get("Accept")
	switch accept {
	default:
		_ = render.Render(
			r.Context(),
			w,
			http.StatusInternalServerError,
			pages.InternalServerErrorPage(),
		)
	}
}
func (app *Application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Error().Err(err).Send()
	chirender.HTML(w, r, "<h2>ERROR</h2>")
}
