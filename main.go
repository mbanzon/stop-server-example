package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	server, errCh := setupServer()

	// wait for the error (if any)
	go func() {
		err := <-errCh
		if err != nil {
			if err == http.ErrServerClosed {
				fmt.Println("Server closed")
				return
			}
			fmt.Println("Error from HTTP server: ", err)
		}
	}()

	done := setupClosedownHandling(server, errCh)

	// wait for the shutdown to complete
	<-done
	fmt.Println("Program done...")
}

func setupClosedownHandling(server *http.Server, errCh chan error) chan struct{} {
	// setup signal handler to gracefully shutdown the program
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// make channel to wait for the shutdown signal
	done := make(chan struct{})

	// wait for shutdown signal and shutdown the server
	go func() {
		<-sig
		fmt.Println("Shutting down server...")
		err := server.Shutdown(context.Background())
		if err != nil {
			errCh <- err
		}
		done <- struct{}{}
	}()

	return done
}

func setupServer() (*http.Server, chan error) {
	errCh := make(chan error, 1)

	server := &http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(5 * time.Second)
			w.Write([]byte("Slept 5 seconds! Hello, World!"))
		}),
	}

	go func() {
		errCh <- server.ListenAndServe()
	}()

	return server, errCh
}
