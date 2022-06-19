package cosmosfaucet

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"github.com/ignite-hq/cli/ignite/pkg/openapiconsole"
)

// ServeHTTP 實現 http.Handler 以通過 HTTP 公開 Faucet.Transfer() 的功能。
// request/響應有效載荷與 allinbits 的先前實現兼容/cosmos-faucet.
func (f Faucet) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := mux.NewRouter()

	router.Handle("/", cors.Default().Handler(http.HandlerFunc(f.faucetHandler))).
		Methods(http.MethodPost)

	router.Handle("/info", cors.Default().Handler(http.HandlerFunc(f.faucetInfoHandler))).
		Methods(http.MethodGet)

	router.HandleFunc("/", openapiconsole.Handler("Faucet", "openapi.yml")).
		Methods(http.MethodGet)

	router.HandleFunc("/openapi.yml", f.openAPISpecHandler).
		Methods(http.MethodGet)

	router.ServeHTTP(w, r)
}
