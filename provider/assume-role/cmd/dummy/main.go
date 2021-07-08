package main

import (
	"net/http"

	assumerole "github.com/fuller-inc/actions-aws-assume-role/provider/assume-role"
)

func main() {
	h := assumerole.NewDummyHandler()
	http.Handle("/", h)
	http.ListenAndServe(":8080", nil)
}
