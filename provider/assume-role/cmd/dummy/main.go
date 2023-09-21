package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	assumerole "github.com/fuller-inc/actions-aws-assume-role/provider/assume-role"
)

func main() {
	chSignal := make(chan os.Signal, 1)
	signal.Notify(chSignal, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(chSignal)

	h := assumerole.NewDummyHandler()
	s := &http.Server{
		Addr:    ":8080",
		Handler: h,
	}
	http.Handle("/", h)

	chServe := make(chan error, 1)
	go func() {
		defer close(chServe)
		chServe <- s.ListenAndServe()
	}()

	select {
	case err := <-chServe:
		log.Fatal(err)
	case <-chSignal:
	}

	signal.Stop(chSignal)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	s.Close()
	<-chServe
}
