package main

import (
	"log/slog"
	"net/http"
	"os"

	assumerole "github.com/fuller-inc/actions-aws-assume-role/provider/assume-role"
	"github.com/shogo82148/aws-xray-yasdk-go/xray/xrayslog"
	"github.com/shogo82148/ridgenative"
)

var logger *slog.Logger

func init() {
	// initialize the logger
	h1 := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	h2 := xrayslog.NewHandler(h1, "trace_id")
	logger = slog.New(h2)
	slog.SetDefault(logger)
}

func main() {
	h := assumerole.NewHandler()
	http.Handle("/", h)
	ridgenative.ListenAndServe(":8080", nil)
}
