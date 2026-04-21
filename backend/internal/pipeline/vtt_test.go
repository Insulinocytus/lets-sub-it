package pipeline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildBilingualVTTKeepsCueTiming(t *testing.T) {
	segments := loadFixtureSegments(t)

	got, err := BuildBilingualVTT(segments)
	if err != nil {
		t.Fatalf("BuildBilingualVTT returned error: %v", err)
	}
	want := "WEBVTT\n\n1\n00:00:01.000 --> 00:00:03.000\nHello world\n你好，世界\n\n2\n00:00:03.000 --> 00:00:05.000\nHow are you\n你好吗\n\n"

	if got != want {
		t.Fatalf("unexpected VTT output:\n%s", got)
	}
}

func TestBuildTranslatedVTTUsesTranslatedText(t *testing.T) {
	segments := loadFixtureSegments(t)

	got, err := BuildTranslatedVTT(segments[:1])
	if err != nil {
		t.Fatalf("BuildTranslatedVTT returned error: %v", err)
	}
	want := "WEBVTT\n\n1\n00:00:01.000 --> 00:00:03.000\n你好，世界\n\n"

	if got != want {
		t.Fatalf("unexpected translated VTT output:\n%s", got)
	}
}

func TestBuildBilingualVTTSanitizesCueText(t *testing.T) {
	segments := []Segment{
		{
			Start:          "00:00:01.000",
			End:            "00:00:03.000",
			SourceText:     "Hello\r\n\r\nworld",
			TranslatedText: "你好\r\n \r\n世界",
		},
	}

	got, err := BuildBilingualVTT(segments)
	if err != nil {
		t.Fatalf("BuildBilingualVTT returned error: %v", err)
	}

	want := "WEBVTT\n\n1\n00:00:01.000 --> 00:00:03.000\nHello\nworld\n你好\n世界\n\n"
	if got != want {
		t.Fatalf("unexpected sanitized VTT output:\n%s", got)
	}

	if segments[0].SourceText != "Hello\r\n\r\nworld" {
		t.Fatalf("BuildBilingualVTT mutated SourceText: %q", segments[0].SourceText)
	}
	if segments[0].TranslatedText != "你好\r\n \r\n世界" {
		t.Fatalf("BuildBilingualVTT mutated TranslatedText: %q", segments[0].TranslatedText)
	}
}

func TestBuildTranslatedVTTValidatesSegments(t *testing.T) {
	segments := []Segment{
		{
			Start:          " ",
			End:            "00:00:03.000",
			TranslatedText: "你好",
		},
	}

	_, err := BuildTranslatedVTT(segments)
	if err == nil {
		t.Fatal("BuildTranslatedVTT returned nil error for invalid segment")
	}
	if !strings.Contains(err.Error(), "start timestamp is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func loadFixtureSegments(t *testing.T) []Segment {
	t.Helper()

	path := filepath.Join("..", "..", "testdata", "transcript_segments.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) returned error: %v", path, err)
	}

	var segments []Segment
	if err := json.Unmarshal(data, &segments); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	return segments
}
