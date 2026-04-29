package runner

import (
	"fmt"
	"strings"
)

type Cue struct {
	TimeLine  string
	TextLines []string
}

func parseWebVTTCues(content string) ([]Cue, error) {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	normalized = strings.TrimSpace(normalized)
	if normalized == "" {
		return nil, fmt.Errorf("empty vtt")
	}
	blocks := strings.Split(normalized, "\n\n")
	if len(blocks) == 0 || strings.TrimSpace(blocks[0]) != "WEBVTT" {
		return nil, fmt.Errorf("vtt must start with WEBVTT")
	}

	var cues []Cue
	for _, block := range blocks[1:] {
		lines := nonEmptyLines(block)
		if len(lines) < 2 {
			return nil, fmt.Errorf("invalid cue block")
		}
		if !strings.Contains(lines[0], "-->") {
			return nil, fmt.Errorf("invalid cue timeline %q", lines[0])
		}
		cues = append(cues, Cue{TimeLine: lines[0], TextLines: lines[1:]})
	}
	if len(cues) == 0 {
		return nil, fmt.Errorf("vtt has no cues")
	}
	return cues, nil
}

func nonEmptyLines(block string) []string {
	rawLines := strings.Split(strings.TrimSpace(block), "\n")
	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func cueText(cue Cue) string {
	return strings.Join(cue.TextLines, "\n")
}

func renderTranslatedVTT(cues []Cue, translations []string) (string, error) {
	if len(cues) != len(translations) {
		return "", fmt.Errorf("translation count %d does not match cue count %d", len(translations), len(cues))
	}
	var b strings.Builder
	b.WriteString("WEBVTT\n\n")
	for i, cue := range cues {
		b.WriteString(cue.TimeLine)
		b.WriteString("\n")
		b.WriteString(strings.TrimSpace(translations[i]))
		b.WriteString("\n\n")
	}
	return b.String(), nil
}

func renderBilingualVTT(cues []Cue, translations []string) (string, error) {
	if len(cues) != len(translations) {
		return "", fmt.Errorf("translation count %d does not match cue count %d", len(translations), len(cues))
	}
	var b strings.Builder
	b.WriteString("WEBVTT\n\n")
	for i, cue := range cues {
		b.WriteString(cue.TimeLine)
		b.WriteString("\n")
		b.WriteString(cueText(cue))
		b.WriteString("\n")
		b.WriteString(strings.TrimSpace(translations[i]))
		b.WriteString("\n\n")
	}
	return b.String(), nil
}
