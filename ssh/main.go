package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	sshTunnel "github.com/shellhub-io/shellhub/ssh/pkg/tunnel"
	"github.com/shellhub-io/shellhub/ssh/server"
	"github.com/shellhub-io/shellhub/ssh/server/handler"
	"github.com/shellhub-io/shellhub/ssh/web"
	"github.com/shellhub-io/shellhub/ssh/web/pkg/cache"
	log "github.com/sirupsen/logrus"
)

func init() {
	loglevel.SetLogLevel()
}

func main() {
	// Populates configuration based on environment variables prefixed with 'SSH_'
	var opts server.Options
	if err := envconfig.Process("ssh", &opts); err != nil {
		log.WithError(err).Fatal("Failed to load environment variables")
	}

	if err := cache.ConnectRedis(opts.RedisURI); err != nil {
		log.WithError(err).Fatal("Failed to connect to redis")
	}

	tunnel := sshTunnel.NewTunnel("/ssh/connection", "/ssh/revdial")

	router := tunnel.GetRouter()
	router.HandleFunc("/sessions/{uid}/close", func(response http.ResponseWriter, request *http.Request) {
		exit := func(response http.ResponseWriter, status int, err error) {
			log.WithError(err).WithFields(log.Fields{
				"status": status,
			}).Error("failed to close the session")

			http.Error(response, err.Error(), status)
		}

		vars := mux.Vars(request)
		decoder := json.NewDecoder(request.Body)
		var closeRequest struct {
			Device string `json:"device"`
		}

		if err := decoder.Decode(&closeRequest); err != nil {
			exit(response, http.StatusBadRequest, err)

			return
		}

		conn, err := tunnel.Dial(context.Background(), closeRequest.Device)
		if err != nil {
			exit(response, http.StatusInternalServerError, err)

			return
		}

		request, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/ssh/close/%s", vars["uid"]), nil)
		if err := request.Write(conn); err != nil {
			exit(response, http.StatusInternalServerError, err)

			return
		}
	})

	// TODO: add `/ws/ssh` route to OpenAPI repository.
	router.Handle("/ws/ssh", web.HandlerRestoreSession(web.RestoreSession, handler.WebSession)).
		Methods(http.MethodGet)
	router.HandleFunc("/ws/ssh", web.HandlerCreateSession(web.CreateSession)).
		Methods(http.MethodPost)

	go http.ListenAndServe(":8080", router) // nolint:errcheck

	log.Fatal(server.NewServer(&opts, tunnel.Tunnel).ListenAndServe())
}
