package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type HTTPTranscriber struct {
	baseURL string
	apiKey  string
	timeout time.Duration
	client  *http.Client
}

func NewHTTPTranscriber(baseURL string, apiKey string, timeout time.Duration, client *http.Client) *HTTPTranscriber {
	if client == nil {
		client = http.DefaultClient
	}
	return &HTTPTranscriber{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		timeout: timeout,
		client:  client,
	}
}

func (t *HTTPTranscriber) Transcribe(ctx context.Context, request TranscriptionRequest) error {
	requestCtx := ctx
	if t.timeout > 0 {
		var cancel context.CancelFunc
		requestCtx, cancel = context.WithTimeout(ctx, t.timeout)
		defer cancel()
	}

	segments, err := t.transcribe(requestCtx, request)
	if err != nil {
		return err
	}
	vtt, err := renderSourceVTT(segments)
	if err != nil {
		return err
	}
	if err := ensureSourceDir(request.SourcePath); err != nil {
		return err
	}
	return writeFileAtomic(request.SourcePath, vtt)
}

func (t *HTTPTranscriber) transcribe(ctx context.Context, request TranscriptionRequest) ([]transcriptionSegment, error) {
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.baseURL+"/audio/transcriptions", pipeReader)
	if err != nil {
		_ = pipeReader.CloseWithError(err)
		return nil, fmt.Errorf("create transcription request: %w", err)
	}
	defer pipeReader.Close()
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if t.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+t.apiKey)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		_ = pipeReader.CloseWithError(err)
		return nil, fmt.Errorf("send transcription request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		message := readErrorMessage(resp.Body)
		if t.apiKey != "" {
			message = strings.ReplaceAll(message, t.apiKey, "[redacted]")
		}
		return nil, fmt.Errorf("transcription request failed with status %d: %s", resp.StatusCode, message)
	}

	var response transcriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode transcription response: %w", err)
	}
	if err := validateSegments(response.Segments); err != nil {
		return nil, err
	}
	return response.Segments, nil
}

func writeTranscriptionForm(writer *multipart.Writer, request TranscriptionRequest) error {
	fields := []struct {
		name  string
		value string
	}{
		{name: "model", value: request.Model},
		{name: "response_format", value: "verbose_json"},
	}
	if request.Language != "" {
		fields = append(fields, struct {
			name  string
			value string
		}{name: "language", value: request.Language})
	}
	for _, field := range fields {
		if err := writer.WriteField(field.name, field.value); err != nil {
			return fmt.Errorf("write transcription form field %s: %w", field.name, err)
		}
	}

	file, err := os.Open(request.AudioPath)
	if err != nil {
		return fmt.Errorf("open audio file: %w", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(request.AudioPath))
	if err != nil {
		return fmt.Errorf("create audio form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("write audio form file: %w", err)
	}
	return nil
}

func renderSourceVTT(segments []transcriptionSegment) (string, error) {
	var b strings.Builder
	b.WriteString("WEBVTT\n\n")
	for _, segment := range segments {
		text, err := formatCueText(segment.Text)
		if err != nil {
			return "", err
		}
		b.WriteString(formatTimestamp(*segment.Start))
		b.WriteString(" --> ")
		b.WriteString(formatTimestamp(*segment.End))
		b.WriteString("\n")
		b.WriteString(text)
		b.WriteString("\n\n")
	}
	return b.String(), nil
}

func formatCueText(text string) (string, error) {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}
	if len(cleaned) == 0 {
		return "", fmt.Errorf("transcription segment text is empty")
	}
	for _, line := range cleaned {
		if strings.Contains(line, "-->") {
			return "", fmt.Errorf("transcription segment text contains a timeline marker")
		}
	}
	return strings.Join(cleaned, "\n"), nil
}

func formatTimestamp(seconds float64) string {
	milliseconds := int64(seconds*1000 + 0.5)
	if milliseconds < 0 {
		milliseconds = 0
	}
	hours := milliseconds / 3600000
	minutes := (milliseconds / 60000) % 60
	secs := (milliseconds / 1000) % 60
	millis := milliseconds % 1000
	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, secs, millis)
}

func validateSegments(segments []transcriptionSegment) error {
	if len(segments) == 0 {
		return fmt.Errorf("transcription response contains no segments")
	}
	for i, segment := range segments {
		if segment.Start == nil {
			return fmt.Errorf("transcription segment %d start is required", i+1)
		}
		if segment.End == nil {
			return fmt.Errorf("transcription segment %d end is required", i+1)
		}
		if !isFinite(*segment.Start) || !isFinite(*segment.End) {
			return fmt.Errorf("transcription segment %d has non-finite timestamps", i+1)
		}
		if *segment.Start < 0 || *segment.End < 0 {
			return fmt.Errorf("transcription segment %d has negative timestamps", i+1)
		}
		if *segment.End <= *segment.Start {
			return fmt.Errorf("transcription segment %d end %v must be after start %v", i+1, *segment.End, *segment.Start)
		}
		if i > 0 && *segment.Start < *segments[i-1].Start {
			return fmt.Errorf("transcription segment %d start %v is before previous segment start %v", i+1, *segment.Start, *segments[i-1].Start)
		}
		if strings.TrimSpace(segment.Text) == "" {
			return fmt.Errorf("transcription segment %d text is empty", i+1)
		}
		if secondsToMilliseconds(*segment.End) <= secondsToMilliseconds(*segment.Start) {
			return fmt.Errorf("transcription segment %d timestamps are not positive after rounding", i+1)
		}
	}
	return nil
}

func secondsToMilliseconds(seconds float64) int64 {
	return int64(seconds*1000 + 0.5)
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func writeFileAtomic(path string, content string) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+"-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary source.vtt: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.WriteString(content); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write source.vtt: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close source.vtt: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replace source.vtt: %w", err)
	}
	return nil
}

func readErrorMessage(body io.Reader) string {
	data, err := io.ReadAll(io.LimitReader(body, 4096))
	if err != nil {
		return ""
	}
	var response struct {
		Detail string `json:"detail"`
		Error  struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &response); err == nil {
		if message := strings.TrimSpace(response.Error.Message); message != "" {
			return message
		}
		if detail := strings.TrimSpace(response.Detail); detail != "" {
			return detail
		}
	}
	return strings.TrimSpace(string(data))
}

type transcriptionResponse struct {
	Segments []transcriptionSegment `json:"segments"`
}

type transcriptionSegment struct {
	Start *float64 `json:"start"`
	End   *float64 `json:"end"`
	Text  string   `json:"text"`
}
