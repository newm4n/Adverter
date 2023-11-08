package web

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hyperjumptech/jiffy"
	"github.com/newm4n/Adverter/server/config"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

var (
	// Router instance of gorilla mux.Router
	Router *mux.Router
)

// InitializeRouter initializes Gorilla Mux and all handler, including Database and Mailer connector
func InitializeRouter() {
	log.Info("Initializing server")
	Router = mux.NewRouter()

	Router.HandleFunc("/path/{b64path}/files", ListFiles).Methods(http.MethodGet)
	Router.HandleFunc("/path/{b64path}/directories", ListDirectories).Methods(http.MethodGet)
	Router.HandleFunc("/path/{b64path}/chunk/info", GetChunkInfo).Methods(http.MethodGet)
	Router.HandleFunc("/path/{b64path}/chunk/{chunkno}", GetChunkData).Methods(http.MethodGet)

	Walk()
}

func configureLogging() {
	lLevel := config.Get("server.log.level")
	fmt.Println("Setting log level to ", lLevel)
	switch strings.ToUpper(lLevel) {
	default:
		fmt.Println("Unknown level [", lLevel, "]. Log level set to ERROR")
		log.SetLevel(log.ErrorLevel)
	case "TRACE":
		log.SetLevel(log.TraceLevel)
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	case "FATAL":
		log.SetLevel(log.FatalLevel)
	}
}

// Start this server
func Start() {
	configureLogging()
	log.Infof("Starting Server")
	startTime := time.Now()

	InitializeRouter()

	var wait time.Duration

	graceShut, err := jiffy.DurationOf(config.Get("server.timeout.graceshut"))
	if err != nil {
		panic(err)
	}
	WriteTimeout, err := jiffy.DurationOf(config.Get("server.timeout.write"))
	if err != nil {
		panic(err)
	}
	ReadTimeout, err := jiffy.DurationOf(config.Get("server.timeout.read"))
	if err != nil {
		panic(err)
	}
	IdleTimeout, err := jiffy.DurationOf(config.Get("server.timeout.idle"))
	if err != nil {
		panic(err)
	}

	wait = graceShut

	address := fmt.Sprintf("%s:%s", config.Get("server.host"), config.Get("server.port"))
	log.Info("Server binding to ", address)

	srv := &http.Server{
		Addr: address,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: WriteTimeout,
		ReadTimeout:  ReadTimeout,
		IdleTimeout:  IdleTimeout,
		Handler:      Router, // Pass our instance of gorilla/mux in.
	}
	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	dur := time.Now().Sub(startTime)
	durDesc := jiffy.DescribeDuration(dur, jiffy.NewWant())
	log.Infof("Shutting down. This Hansip been protecting the world for %s", durDesc)
	os.Exit(0)
}

// Walk and show all endpoint that available on this server
func Walk() {
	err := Router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err != nil {
			log.Error(err)
		}
		methods, err := route.GetMethods()
		if err != nil {
			log.Error(err)
		}

		log.Infof("Route : %s [%s]", pathTemplate, strings.Join(methods, ","))
		return nil
	})

	if err != nil {
		log.Error(err)
	}
}
