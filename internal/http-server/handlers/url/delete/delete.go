package delete

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"
)

type Response struct {
	resp.Response
}

type UrlDeleter interface {
	DeleteURL(url string) (bool, error)
}

func New(log *slog.Logger, urlDeleter UrlDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")

			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		res, err := urlDeleter.DeleteURL(alias)
		if errors.Is(err, storage.ErrURLNotDeleted) {
			log.Info("url not deleted", "alias", alias)

			render.JSON(w, r, resp.Error("url not deleted"))

			return
		}

		if err != nil {
			log.Error("failed to delete url", sl.Err(err))

			render.JSON(w, r, resp.Error("internal server error"))

			return
		}

		log.Info("deleted url", slog.Bool("url", res))

		render.JSON(w, r, Response{
			Response: resp.OK(),
		})
	}
}
