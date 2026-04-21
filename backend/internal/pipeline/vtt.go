package pipeline

import (
	"errors"
	"fmt"
	"strings"
)

func BuildTranslatedVTT(segments []Segment) (string, error) {
	return buildVTT(segments, false)
}

func BuildBilingualVTT(segments []Segment) (string, error) {
	return buildVTT(segments, true)
}

func buildVTT(segments []Segment, bilingual bool) (string, error) {
	var b strings.Builder
	b.WriteString("WEBVTT\n\n")

	for i, segment := range segments {
		sanitized := Segment{
			Start:          strings.TrimSpace(segment.Start),
			End:            strings.TrimSpace(segment.End),
			SourceText:     sanitizeCueText(segment.SourceText),
			TranslatedText: sanitizeCueText(segment.TranslatedText),
		}

		if err := validateSegment(sanitized, bilingual); err != nil {
			return "", fmt.Errorf("segment %d: %w", i+1, err)
		}

		if bilingual {
			b.WriteString(fmt.Sprintf("%d\n%s --> %s\n%s\n%s\n\n", i+1, sanitized.Start, sanitized.End, sanitized.SourceText, sanitized.TranslatedText))
			continue
		}

		b.WriteString(fmt.Sprintf("%d\n%s --> %s\n%s\n\n", i+1, sanitized.Start, sanitized.End, sanitized.TranslatedText))
	}

	return b.String(), nil
}

func sanitizeCueText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	lines := strings.Split(text, "\n")
	sanitized := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		sanitized = append(sanitized, line)
	}

	return strings.Join(sanitized, "\n")
}

func validateSegment(segment Segment, bilingual bool) error {
	if segment.Start == "" {
		return errors.New("start timestamp is required")
	}
	if segment.End == "" {
		return errors.New("end timestamp is required")
	}
	if bilingual && segment.SourceText == "" {
		return errors.New("source text is required")
	}
	if segment.TranslatedText == "" {
		return errors.New("translated text is required")
	}

	return nil
}
