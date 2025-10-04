package service

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"sync"

	"concurrent-data/internal/progress"

	"golang.org/x/sync/errgroup"
)

type Result struct {
	File     string `json:"file"`
	Rows     int    `json:"rows"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
	Progress int    `json:"progress"`
}

var tracker *progress.Tracker
var trackerMu sync.Mutex

func ProcessFiles(path string, workers int) ([]Result, error) {
	files, _ := filepath.Glob(filepath.Join(path, "*.csv"))
	if len(files) == 0 {
		return []Result{}, nil
	}

	results := make([]Result, 0, len(files))
	mu := sync.Mutex{}

	trackerMu.Lock()
	tracker = progress.New(len(files))
	trackerMu.Unlock()

	g := new(errgroup.Group)
	sem := make(chan struct{}, workers)

	for _, f := range files {
		file := f
		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			count, err := processCSV(file)
			percent := tracker.Inc()

			res := Result{
				File:     file,
				Rows:     count,
				Status:   "success",
				Progress: percent,
			}
			if err != nil {
				res.Status = "failed"
				res.Error = err.Error()
			}

			mu.Lock()
			results = append(results, res)
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return results, err
	}
	return results, nil
}

func GetProgress() progress.Tracker {
	trackerMu.Lock()
	defer trackerMu.Unlock()
	if tracker == nil {
		return progress.Tracker{}
	}
	return tracker.Snapshot()
}

func processCSV(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return 0, err
	}
	return len(records), nil
}
