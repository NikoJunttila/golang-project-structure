package handlers

import (
	"log"
	"net/http"

	"github.com/nikojunttila/community/db"
)

func GetFooHandler(w http.ResponseWriter, r *http.Request) {
	foo, err := db.Get().GetFoo(r.Context(), 1)
	if err != nil {
		log.Println("err: foo handler, ", err)
		RespondWithError(w, 404, "error with foo")
	}
	RespondWithJson(w, 200, foo)
}
