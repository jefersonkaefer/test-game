package main

import (
	"log"
	"net/http"

	"game/api/internal/application"
	"game/api/internal/application/controller"
	"game/api/internal/application/repository"
	"game/api/internal/infra/database"
	"game/api/internal/infra/network"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic recovery: %v", err)
		}
	}()

	conn, err := application.DbConn()
	if err != nil {
		log.Fatalf("ERROR occurs: %v", err)
	}

	pg := database.NewPostgres(conn)
	cacheClient := application.CacheConn()
	redis := database.NewRedis(cacheClient)

	clientsRepo := repository.NewClient(pg, redis)
	clientCtrl := controller.NewClient(clientsRepo)
	app := &application.App{
		ClientCtrl: clientCtrl,
	}
	api := network.NewWebServer(app)

	http.HandleFunc("/client", api.NewClient)
	http.HandleFunc("/login", api.Login)

	log.Println("HTTP server started on port :4300")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatalf("An error occurred while starting HTTP: %v", err)
	}
}
