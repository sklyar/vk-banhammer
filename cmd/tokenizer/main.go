package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
)

type Config struct {
	Port        int    `long:"port" env:"PORT" description:"Port" default:"8080"`
	ClientID    string `long:"client-id" env:"CLIENT_ID" description:"Application ID" required:"true"`
	GroupID     string `long:"group-id" env:"GROUP_ID" description:"Identifier of the community for which the token will be issued" required:"true"`
	RedirectURI string `long:"redirect-uri" env:"REDIRECT_URI" description:"Address to which the user will be redirected after authorization" required:"true"`
	Scope       string `long:"scope" env:"SCOPE" description:"Bitmask of the application's access settings, which should be checked during authorization and the missing ones should be requested" required:"true"`
}

func main() {
	var cfg Config
	if _, err := flags.Parse(&cfg); err != nil {
		log.Printf("failed to parse flags: %v\n", err)
		return
	}

	app := &App{Config: cfg}
	mux := http.NewServeMux()
	mux.HandleFunc("/authorize", app.Authorize)
	mux.HandleFunc("/callback", app.Callback)

	server := &http.Server{
		Addr:              ":" + strconv.Itoa(cfg.Port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		<-ctx.Done()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("failed to shutdown server: %v\n", err)
		}
	}()

	fmt.Printf("Server is listening on %d port\n", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Printf("failed to start server: %v\n", err)
	}
}

type App struct {
	Config
}

func (app *App) Authorize(w http.ResponseWriter, r *http.Request) {
	rawURL := "https://oauth.vk.com/authorize?display=page&response_type=token&v=5.131"

	u, err := url.Parse(rawURL)
	if err != nil {
		log.Printf("failed to parse url: %v\n", err)
		return
	}

	q := u.Query()
	q.Set("client_id", app.ClientID)
	q.Set("group_ids", app.GroupID)
	q.Set("redirect_uri", app.RedirectURI)
	q.Set("scope", app.Scope)
	u.RawQuery = q.Encode()

	rawURL = u.String()

	http.Redirect(w, r, rawURL, http.StatusFound)
}

func (app *App) Callback(_ http.ResponseWriter, _ *http.Request) {}
