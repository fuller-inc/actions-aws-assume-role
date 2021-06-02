package main

import (
	"net/http"

	assumerole "github.com/fuller-inc/actions-aws-assume-role/provider/assume-role"
	"github.com/fuller-inc/ridgenative"
)

func main() {
	h := assumerole.NewHandler()
	http.Handle("/", h)
	ridgenative.ListenAndServe(":8080", nil)
}
