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
	f, err := os.OpenFile("dummy.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.SetOutput(f)

	if err := serve(); err != nil {
		log.Fatal(err)
	}
}

func serve() error {
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
		return err
	case <-chSignal:
	}

	signal.Stop(chSignal)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		return err
	}
	s.Close()
	<-chServe

	return nil
}
