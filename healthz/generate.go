package healthz

import (
	"encoding/json"
	"net/http"
)

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.3.0 -config cfg.yaml api.yaml

type dependencyCheck func() (Dependency, bool)

func HealthzHandler(h *http.ServeMux, checks ...dependencyCheck) {
	h.Handle(
		"/status",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			deps := []Dependency{}
			status := "OK"

			for _, dCheck := range checks {
				d, ok := dCheck()
				if !ok {
					status = "KO"
				}

				deps = append(deps, d)
			}

			json.NewEncoder(w).Encode(Status{
				Dependencies: deps,
				Status:       status,
			})
		}))
}
