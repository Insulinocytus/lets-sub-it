package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type HTTPTranscriber struct {
	baseURL      string
	timeout      time.Duration
	pollInterval time.Duration
	client       *http.Client
}

func NewHTTPTranscriber(baseURL string, timeout time.Duration, pollInterval time.Duration, client *http.Client) *HTTPTranscriber {
	if client == nil {
		client = http.DefaultClient
	}
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}
	return &HTTPTranscriber{
		baseURL:      strings.TrimRight(baseURL, "/"),
		timeout:      timeout,
		pollInterval: pollInterval,
		client:       client,
	}
}

func (t *HTTPTranscriber) Transcribe(ctx context.Context, request TranscriptionRequest) (resultErr error) {
	requestCtx := ctx
	var cancel context.CancelFunc
	if t.timeout > 0 {
		requestCtx, cancel = context.WithTimeout(ctx, t.timeout)
		defer cancel()
	}
	remoteID := ""
	defer func() {
		if remoteID == "" {
			return
		}
		cleanupCtx, cleanupCancel := context.WithTimeout(context.WithoutCancel(requestCtx), 10*time.Second)
		defer cleanupCancel()
		if cleanupErr := t.delete(cleanupCtx, remoteID); cleanupErr != nil && resultErr == nil {
			resultErr = cleanupErr
		}
	}()

	status, err := t.upload(requestCtx, request)
	if err != nil {
		return err
	}
	remoteID = status.ID

	for {
		if err := reportTranscriptionProgress(request, status.ProgressText); err != nil {
			return err
		}

		switch status.Status {
		case "completed":
			return t.downloadVTT(requestCtx, status.ID, request.SourcePath)
		case "failed":
			if status.ErrorMessage != "" {
				return fmt.Errorf("transcription failed: %s", status.ErrorMessage)
			}
			return fmt.Errorf("transcription failed")
		case "queued", "running":
			if err := sleepContext(requestCtx, t.pollInterval); err != nil {
				return fmt.Errorf("wait before polling transcription status: %w", err)
			}
			status, err = t.poll(requestCtx, status.ID)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid transcription status %q", status.Status)
		}
	}
}

func (t *HTTPTranscriber) delete(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, t.transcriptionURL(id), nil)
	if err != nil {
		return fmt.Errorf("create transcription delete request: %w", err)
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("delete transcription: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("delete transcription failed with status %d: %s", resp.StatusCode, readErrorMessage(resp.Body))
	}
	return nil
}

func (t *HTTPTranscriber) upload(ctx context.Context, request TranscriptionRequest) (transcriptionStatus, error) {
	pipeReader, pipeWriter := io.Pipe()
	writer := multipart.NewWriter(pipeWriter)
	go func() {
		err := writeTranscriptionForm(writer, request)
		if closeErr := writer.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			_ = pipeWriter.CloseWithError(err)
			return
		}
		_ = pipeWriter.Close()
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.baseURL+"/transcriptions", pipeReader)
	if err != nil {
		_ = pipeReader.Close()
		return transcriptionStatus{}, fmt.Errorf("create transcription request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return t.doStatusRequest(req)
}

func writeTranscriptionForm(writer *multipart.Writer, request TranscriptionRequest) error {
	for _, field := range []struct {
		name  string
		value string
	}{
		{name: "model", value: request.Model},
		{name: "computeType", value: request.ComputeType},
		{name: "language", value: request.Language},
		{name: "jobId", value: request.JobID},
	} {
		if err := writer.WriteField(field.name, field.value); err != nil {
			return fmt.Errorf("write transcription form field %s: %w", field.name, err)
		}
	}

	file, err := os.Open(request.AudioPath)
	if err != nil {
		return fmt.Errorf("open audio file: %w", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("audio", filepath.Base(request.AudioPath))
	if err != nil {
		return fmt.Errorf("create audio form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("write audio form file: %w", err)
	}
	return nil
}

func (t *HTTPTranscriber) poll(ctx context.Context, id string) (transcriptionStatus, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.transcriptionURL(id), nil)
	if err != nil {
		return transcriptionStatus{}, fmt.Errorf("create transcription status request: %w", err)
	}
	return t.doStatusRequest(req)
}

func (t *HTTPTranscriber) downloadVTT(ctx context.Context, id string, sourcePath string) error {
	if err := ensureSourceDir(sourcePath); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.transcriptionURL(id)+"/vtt", nil)
	if err != nil {
		return fmt.Errorf("create transcription VTT request: %w", err)
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("download transcription VTT: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("download transcription VTT failed with status %d: %s", resp.StatusCode, readErrorMessage(resp.Body))
	}

	tmp, err := os.CreateTemp(filepath.Dir(sourcePath), filepath.Base(sourcePath)+"-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary source.vtt: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	_, copyErr := io.Copy(tmp, resp.Body)
	closeErr := tmp.Close()
	if copyErr != nil {
		return fmt.Errorf("write source.vtt: %w", copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close source.vtt: %w", closeErr)
	}
	if err := ensureValidSourceVTT(tmpPath); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, sourcePath); err != nil {
		return fmt.Errorf("replace source.vtt: %w", err)
	}
	return nil
}

func (t *HTTPTranscriber) doStatusRequest(req *http.Request) (transcriptionStatus, error) {
	resp, err := t.client.Do(req)
	if err != nil {
		return transcriptionStatus{}, fmt.Errorf("send transcription request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return transcriptionStatus{}, fmt.Errorf("transcription request failed with status %d: %s", resp.StatusCode, readErrorMessage(resp.Body))
	}

	var transcriptionResp transcriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&transcriptionResp); err != nil {
		return transcriptionStatus{}, fmt.Errorf("decode transcription response: %w", err)
	}
	status := transcriptionResp.Transcription
	if status.ID == "" {
		return transcriptionStatus{}, fmt.Errorf("transcription response transcription.id is required")
	}
	if status.Status == "" {
		return transcriptionStatus{}, fmt.Errorf("transcription response transcription.status is required")
	}
	return status, nil
}

func (t *HTTPTranscriber) transcriptionURL(id string) string {
	return t.baseURL + "/transcriptions/" + url.PathEscape(id)
}

func reportTranscriptionProgress(request TranscriptionRequest, status string) error {
	if request.OnProgress == nil || status == "" {
		return nil
	}
	if err := request.OnProgress(status); err != nil {
		return fmt.Errorf("report transcription progress: %w", err)
	}
	return nil
}

func readErrorMessage(body io.Reader) string {
	data, err := io.ReadAll(io.LimitReader(body, 4096))
	if err != nil {
		return ""
	}
	var response struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &response); err == nil && response.Error.Message != "" {
		return response.Error.Message
	}
	return strings.TrimSpace(string(data))
}

type transcriptionResponse struct {
	Transcription transcriptionStatus `json:"transcription"`
}

type transcriptionStatus struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	ProgressText string `json:"progressText"`
	ErrorMessage string `json:"errorMessage"`
}
