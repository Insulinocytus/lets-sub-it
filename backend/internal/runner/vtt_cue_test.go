package runner

import (
	"strings"
	"testing"
)

func TestParseWebVTTCuesParsesMultipleCuesAndMultilineText(t *testing.T) {
	input := "WEBVTT\n\n00:00:00.000 --> 00:00:01.000\nline one\nline two\n\n00:00:01.000 --> 00:00:02.000\nsecond\n"

	cues, err := parseWebVTTCues(input)
	if err != nil {
		t.Fatalf("parseWebVTTCues() error = %v", err)
	}

	if len(cues) != 2 {
		t.Fatalf("len(cues) = %d, want 2", len(cues))
	}
	if cues[0].TimeLine != "00:00:00.000 --> 00:00:01.000" {
		t.Fatalf("TimeLine = %q", cues[0].TimeLine)
	}
	if got := cueText(cues[0]); got != "line one\nline two" {
		t.Fatalf("cueText = %q", got)
	}
}

func TestParseWebVTTCuesRejectsInvalidHeader(t *testing.T) {
	_, err := parseWebVTTCues("not vtt\n\n00:00:00.000 --> 00:00:01.000\ntext\n")
	if err == nil {
		t.Fatal("parseWebVTTCues() error = nil, want error")
	}
}

func TestRenderTranslatedVTTUsesSourceTimeline(t *testing.T) {
	cues := []Cue{
		{TimeLine: "00:00:00.000 --> 00:00:01.000", TextLines: []string{"hello"}},
		{TimeLine: "00:00:01.000 --> 00:00:02.000", TextLines: []string{"world"}},
	}

	got, err := renderTranslatedVTT(cues, []string{"你好", "世界"})
	if err != nil {
		t.Fatalf("renderTranslatedVTT() error = %v", err)
	}

	if !strings.Contains(got, "00:00:00.000 --> 00:00:01.000\n你好") {
		t.Fatalf("translated VTT = %q", got)
	}
	if strings.Contains(got, "hello") {
		t.Fatalf("translated VTT = %q, should not include source text", got)
	}
}

func TestRenderBilingualVTTCombinesSourceAndTranslation(t *testing.T) {
	cues := []Cue{
		{TimeLine: "00:00:00.000 --> 00:00:01.000", TextLines: []string{"hello"}},
	}

	got, err := renderBilingualVTT(cues, []string{"你好"})
	if err != nil {
		t.Fatalf("renderBilingualVTT() error = %v", err)
	}

	want := "00:00:00.000 --> 00:00:01.000\nhello\n你好"
	if !strings.Contains(got, want) {
		t.Fatalf("bilingual VTT = %q, want containing %q", got, want)
	}
}

func TestRenderTranslatedVTTRejectsMismatchedCounts(t *testing.T) {
	_, err := renderTranslatedVTT([]Cue{{TimeLine: "00:00:00.000 --> 00:00:01.000", TextLines: []string{"hello"}}}, nil)
	if err == nil {
		t.Fatal("renderTranslatedVTT() error = nil, want error")
	}
}

func TestRenderVTTRejectsInvalidTranslations(t *testing.T) {
	cues := []Cue{
		{TimeLine: "00:00:00.000 --> 00:00:01.000", TextLines: []string{"hello"}},
	}
	tests := []struct {
		name        string
		translation string
	}{
		{name: "empty after trim", translation: " \n\t "},
		{name: "blank line", translation: "first\n\nsecond"},
		{name: "timeline marker", translation: "bad --> marker"},
	}

	for _, tt := range tests {
		t.Run(tt.name+"/translated", func(t *testing.T) {
			_, err := renderTranslatedVTT(cues, []string{tt.translation})
			if err == nil {
				t.Fatal("renderTranslatedVTT() error = nil, want error")
			}
		})

		t.Run(tt.name+"/bilingual", func(t *testing.T) {
			_, err := renderBilingualVTT(cues, []string{tt.translation})
			if err == nil {
				t.Fatal("renderBilingualVTT() error = nil, want error")
			}
		})
	}
}
