package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/fission/fission"
	"github.com/fission/fission/environments/envsidecar"
)

func dumpStackTrace() {
	debug.PrintStack()
}

// Usage: fetcher <shared volume path>
func main() {
	// register signal handler for dumping stack trace.
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Received SIGTERM : Dumping stack trace")
		dumpStackTrace()
		os.Exit(1)
	}()

	shareVolumeDir := flag.Arg(0)
	if _, err := os.Stat(shareVolumeDir); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(shareVolumeDir, os.ModeDir|0700)
			if err != nil {
				log.Fatalf("Error creating directory: %v", err)
			}
		}
	}

	f, err := envsidecar.MakeEnvSidecar("/", fission.SharedVolumeSecrets, fission.SharedVolumeConfigmaps)
	if err != nil {
		log.Fatalf("Error making fetcher: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/fetch", f.FetchHandler)
	mux.HandleFunc("/upload", f.UploadHandler)
	mux.HandleFunc("/version", f.VersionHandler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Println("Fetcher ready to receive requests")
	http.ListenAndServe(":8000", mux)
}
