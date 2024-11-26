package cors

import (
	"net/http"

	"github.com/rs/cors"
)

type handlerWrapper struct {
	handler http.HandlerFunc
}

func NewHandlerWrapper(handler http.HandlerFunc) *handlerWrapper {
	return &handlerWrapper{handler: handler}
}

func (h *handlerWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler(w, r)
}

func (h *handlerWrapper) HttpHandlerFunc() http.HandlerFunc {
	return h.handler
}

func newCorsAllowAll(allowedOrigins []string) *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
}

func CORSMiddleware(allowedOrigins []string, next http.HandlerFunc) http.HandlerFunc {
	wrapper := NewHandlerWrapper(next)
	h := newCorsAllowAll(allowedOrigins).Handler(wrapper)
	return h.ServeHTTP
}
