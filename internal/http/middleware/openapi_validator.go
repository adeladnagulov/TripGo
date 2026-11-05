package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/adeladnagulov/TripGo/internal/http/domainerror"
	"github.com/adeladnagulov/TripGo/internal/http/helper/response"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func OpenAPIRequestValidator(swagger *openapi3.T) (func(http.Handler) http.Handler, error) {
	router, err := gorillamux.NewRouter(swagger)
	if err != nil {
		return nil, fmt.Errorf("failed create OpenAPI router: %w", err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route, pathParams, err := router.FindRoute(r)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			requestValidationInput := &openapi3filter.RequestValidationInput{
				Request:    r,
				PathParams: pathParams,
				Route:      route,
			}

			if err := openapi3filter.ValidateRequest(r.Context(), requestValidationInput); err != nil {
				ctx := context.WithValue(r.Context(), domainerror.RoutePatternKey, route.Path)
				r = r.WithContext(ctx)
				response.HandleAPIError(w, r, err)
				return
			}
			next.ServeHTTP(w, r)
		})
	}, nil
}
