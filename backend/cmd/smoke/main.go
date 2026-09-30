// Command smoke runs one real job end to end against a running backend: it deletes any
// existing Subtitle Result for the video first, so the pipeline always runs instead of reusing.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"
)

const (
	defaultBackendURL = "http://127.0.0.1:8080"
	defaultVideoID    = "jNQXAC9IVRw" // "Me at the zoo", 19 seconds
	sourceLanguage    = "en"
	targetLanguage    = "zh"
	pollInterval      = 3 * time.Second
	jobTimeout        = 10 * time.Minute
	previewLines      = 12
)

// client bounds each request so a stalled backend cannot outlast jobTimeout.
var client = &http.Client{Timeout: 30 * time.Second}

type job struct {
	ID           string  `json:"id"`
	Status       string  `json:"status"`
	Stage        string  `json:"stage"`
	ErrorMessage *string `json:"errorMessage"`
}

func main() {
	if err := run(envOr("LSI_BACKEND_URL", defaultBackendURL), envOr("LSI_SMOKE_VIDEO_ID", defaultVideoID)); err != nil {
		fmt.Fprintln(os.Stderr, "smoke failed:", err)
		os.Exit(1)
	}
}

func run(backend string, videoID string) error {
	backend = strings.TrimRight(backend, "/")
	videoURL := "https://www.youtube.com/watch?v=" + url.QueryEscape(videoID)

	query := url.Values{"videoId": {videoID}, "targetLanguage": {targetLanguage}}.Encode()
	if _, err := call(http.MethodDelete, backend+"/subtitle-results?"+query, nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("delete existing subtitle result: %w", err)
	}

	body, _ := json.Marshal(map[string]string{"youtubeUrl": videoURL, "sourceLanguage": sourceLanguage, "targetLanguage": targetLanguage})
	// A reused job answers 200; accept it so the reused check below reports it clearly.
	created, err := call(http.MethodPost, backend+"/jobs", body, http.StatusCreated, http.StatusOK)
	if err != nil {
		return fmt.Errorf("create job: %w", err)
	}
	var submitted struct {
		Job    job  `json:"job"`
		Reused bool `json:"reused"`
	}
	if err := json.Unmarshal(created, &submitted); err != nil {
		return fmt.Errorf("decode created job: %w", err)
	}
	if submitted.Reused {
		return fmt.Errorf("job %s was reused right after deleting the subtitle result", submitted.Job.ID)
	}
	fmt.Printf("job %s: %s %s -> %s\n", submitted.Job.ID, videoURL, sourceLanguage, targetLanguage)

	startedAt := time.Now()
	lastStage := ""
	for {
		data, err := call(http.MethodGet, backend+"/jobs/"+submitted.Job.ID, nil, http.StatusOK)
		if err != nil {
			return fmt.Errorf("poll job: %w", err)
		}
		var polled struct {
			Job job `json:"job"`
		}
		if err := json.Unmarshal(data, &polled); err != nil {
			return fmt.Errorf("decode job: %w", err)
		}
		current := polled.Job
		elapsed := time.Since(startedAt).Round(time.Second)
		if current.Stage != lastStage {
			fmt.Printf("[%s] %s\n", elapsed, current.Stage)
			lastStage = current.Stage
		}
		switch current.Status {
		case "completed":
			return printSubtitles(backend, current.ID)
		case "failed":
			message := "(no error message)"
			if current.ErrorMessage != nil {
				message = *current.ErrorMessage
			}
			return fmt.Errorf("failed at %s: %s", current.Stage, message)
		}
		if elapsed > jobTimeout {
			return fmt.Errorf("still %s after %s", current.Stage, jobTimeout)
		}
		time.Sleep(pollInterval)
	}
}

func printSubtitles(backend string, jobID string) error {
	for _, mode := range []string{"source", "translated", "bilingual"} {
		data, err := call(http.MethodGet, backend+"/subtitle-files/"+jobID+"/"+mode, nil, http.StatusOK)
		if err != nil {
			return fmt.Errorf("fetch %s.vtt: %w", mode, err)
		}
		lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
		fmt.Printf("\n--- %s.vtt (%d bytes) ---\n%s\n", mode, len(data), strings.Join(lines[:min(len(lines), previewLines)], "\n"))
	}
	return nil
}

func call(method string, target string, body []byte, wantStatus ...int) ([]byte, error) {
	request, err := http.NewRequest(method, target, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if !slices.Contains(wantStatus, response.StatusCode) {
		return nil, fmt.Errorf("%s %s: status %d: %s", method, target, response.StatusCode, strings.TrimSpace(string(data)))
	}
	return data, nil
}

func envOr(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
