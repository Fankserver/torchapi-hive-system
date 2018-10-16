package main

import (
	"net/http"

	"github.com/fankserver/torchapi-hive-system/src/hive"
	"github.com/fankserver/torchapi-hive-system/src/notification"
	"github.com/sirupsen/logrus"
)

func main() {
	system, err := hive.NewSystem([]string{}, "", "", "")
	if err != nil {
		logrus.Fatalln(err.Error())
	}
	hub := notification.NewHub()
	go hub.Run()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		notification.ServeWs(hub, w, r)
	})
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		logrus.Fatal("ListenAndServe: ", err)
	}
}
