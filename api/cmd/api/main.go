package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/theonlyjohnny/rac/api/internal/api"
	"github.com/theonlyjohnny/rac/api/internal/storage"
)

func main() {
	dao := storage.NewDAO()

	api, err := api.NewAPI(dao)
	if err != nil {
		panic(fmt.Errorf("failed to setup API: %s", err.Error()))
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", getPort()),
		Handler: api.Router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("Failed to shutdown HTTP server cleanly :( -- %s", err.Error())
	}

	select {
	case <-ctx.Done():
		fmt.Println("server shutdown timed out after 5 sec")
	}

	fmt.Println("exiting")
}

func getPort() string {
	if v := os.Getenv("PORT"); v != "" {
		if _, err := strconv.Atoi(v); err != nil {
			return v
		}
	}
	return "8090"
}
