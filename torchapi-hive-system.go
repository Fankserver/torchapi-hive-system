package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/fankserver/torchapi-hive-system/src/hive"
	"github.com/fankserver/torchapi-hive-system/src/notification"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func main() {
	system, err := hive.NewSystem([]string{}, "", "", "")
	if err != nil {
		logrus.Fatalln(err.Error())
	}
	hub := notification.NewHub()
	go hub.Run()

	// subscribe to SIGINT signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	router := mux.NewRouter()
	router.HandleFunc("/", func(writer http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(writer, "TorchAPI Hive System")
	}).Methods(http.MethodGet)
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		notification.ServeWs(hub, w, r)
	})
	router.HandleFunc("/api/hive", system.GetHives).Methods(http.MethodGet)
	router.HandleFunc("/api/hive", system.CreateHive).Methods(http.MethodPost)
	router.HandleFunc("/api/hive/{hive_id:[a-z0-9]+}/sector", system.GetSectors).Methods(http.MethodGet)
	router.HandleFunc("/api/hive/{hive_id:[a-z0-9]+}/sector", system.CreateSector).Methods(http.MethodPost)
	router.HandleFunc("/api/hive/{hive_id:[a-z0-9]+}/sector/{sector_id:[a-z0-9]+}", system.DeleteSector).Methods(http.MethodDelete)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logrus.Fatalf("listen: %s\n", err)
		}
	}()

	<-quit
	logrus.Println("shutting down server...")
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxTimeout); err != nil {
		logrus.Fatalf("could not shutdown: %v", err)
	}
	logrus.Println("server gracefully stopped")
}
