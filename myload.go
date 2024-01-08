package myload

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	LogFormat string `long:"log-format" choice:"text" choice:"json" default:"text" description:"Log format"`
	Verbose   []bool `short:"v" long:"verbose" description:"Show verbose debug information, each -v bumps log level"`
	logLevel  slog.Level
	DataPath  string `short:"d" long:"data-path" description:"Path to the JSON file" required:"true"`
	Port      string `default:"8001" short:"p" long:"port" description:"Listen for request on this port"`
}

func Execute() int {
	if err := parseFlags(); err != nil {
		return 1
	}

	if err := setLogLevel(); err != nil {
		return 1
	}

	if err := setupLogger(); err != nil {
		return 1
	}

	data := opts.DataPath
	if err := run(data); err != nil {
		slog.Error("run failed", "error", err)
		return 1
	}

	return 0
}

func parseFlags() error {
	_, err := flags.Parse(&opts)
	return err
}

func run(dataPath string) error {
	http.HandleFunc("/data/json", func(w http.ResponseWriter, r *http.Request) {
		file, err := os.Open(dataPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error opening JSON file: %s", err), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		w.Header().Set("Content-Type", "application/json")

		_, err = io.Copy(w, file)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error serving JSON file: %s", err), http.StatusInternalServerError)
		}
	})

	err := http.ListenAndServe(opts.Port, nil)
	if err != nil {
		return fmt.Errorf("ListenAndServe: %s", err)
	}
	return nil
}
