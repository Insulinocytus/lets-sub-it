package api

import "testing"

func TestParseVideoIDSupportsWatchURL(t *testing.T) {
	videoID, err := ParseVideoID("https://www.youtube.com/watch?v=abc123")
	if err != nil {
		t.Fatalf("ParseVideoID() error = %v", err)
	}
	if videoID != "abc123" {
		t.Fatalf("videoID = %q", videoID)
	}
}

func TestParseVideoIDSupportsShortURL(t *testing.T) {
	videoID, err := ParseVideoID("https://youtu.be/abc123")
	if err != nil {
		t.Fatalf("ParseVideoID() error = %v", err)
	}
	if videoID != "abc123" {
		t.Fatalf("videoID = %q", videoID)
	}
}

func TestParseVideoIDRejectsUnsupportedURL(t *testing.T) {
	if _, err := ParseVideoID("https://example.com/watch?v=abc123"); err == nil {
		t.Fatal("ParseVideoID() expected error")
	}
}

func TestParseVideoIDRejectsUnsupportedScheme(t *testing.T) {
	if _, err := ParseVideoID("ftp://www.youtube.com/watch?v=abc123"); err == nil {
		t.Fatal("ParseVideoID() expected error")
	}
}

func TestParseVideoIDRejectsNonWatchYouTubeURL(t *testing.T) {
	if _, err := ParseVideoID("https://www.youtube.com/embed/abc123?v=abc123"); err == nil {
		t.Fatal("ParseVideoID() expected error")
	}
}
