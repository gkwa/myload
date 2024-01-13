package myload

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	LogFormat     string `long:"log-format" choice:"text" choice:"json" default:"text" description:"Log format"`
	Verbose       []bool `short:"v" long:"verbose" description:"Show verbose debug information, each -v bumps log level"`
	logLevel      slog.Level
	DataPathRaw   string `long:"data-raw" description:"Path to full data raw" required:"true"`
	DataPathDaily string `long:"data-daily" description:"Path to daily summary" required:"true"`
	Port          string `short:"p" long:"port" description:"Listen for request on this port" default:"8001"`
}

func Execute() int {
	if err := parseFlags(); err != nil {
		fmt.Println(err)
		return 1
	}

	if err := setLogLevel(); err != nil {
		return 1
	}

	if err := setupLogger(); err != nil {
		return 1
	}

	if err := run(); err != nil {
		slog.Error("run failed", "error", err)
		return 1
	}

	return 0
}

func parseFlags() error {
	_, err := flags.Parse(&opts)
	if err != nil {
		return fmt.Errorf("error parsing flags: %v", err)
	}
	return nil
}

func run() error {
	slog.Info("Starting server", "port", opts.Port)
	slog.Info("Serving raw data", "path", opts.DataPathRaw)
	slog.Info("Serving daily data", "path", opts.DataPathDaily)

	if err := checkDataPath(opts.DataPathRaw); err != nil {
		return fmt.Errorf("data path not found: %s", opts.DataPathRaw)
	}

	if err := checkDataPath(opts.DataPathDaily); err != nil {
		return fmt.Errorf("calling checkDataPath for %s error: %v", opts.DataPathDaily, err)
	}

	http.HandleFunc("/data/json/raw", func(w http.ResponseWriter, r *http.Request) {
		serveJSONFile(w, opts.DataPathRaw)
	})

	http.HandleFunc("/data/json/daily", func(w http.ResponseWriter, r *http.Request) {
		serveJSONFile(w, opts.DataPathDaily)
	})

	err := http.ListenAndServe(fmt.Sprintf(":%s", opts.Port), nil)
	if err != nil {
		return fmt.Errorf("ListenAndServe: %s", err)
	}
	return nil
}

func serveJSONFile(w http.ResponseWriter, path string) {
	slog.Info("Serving JSON file", "timestamp", time.Now(), "path", path)
	file, err := os.Open(path)
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
}

func checkDataPath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("data path not found: %s", path)
	}
	return nil
}
