package api

import (
	"errors"
	"net/url"
	"strings"
)

var ErrInvalidYouTubeURL = errors.New("invalid youtube url")

func ParseVideoID(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", ErrInvalidYouTubeURL
	}

	host := strings.ToLower(parsed.Host)
	if host == "www.youtube.com" || host == "youtube.com" {
		videoID := parsed.Query().Get("v")
		if videoID == "" {
			return "", ErrInvalidYouTubeURL
		}
		return videoID, nil
	}

	if host == "youtu.be" {
		videoID := strings.Trim(parsed.Path, "/")
		if videoID == "" || strings.Contains(videoID, "/") {
			return "", ErrInvalidYouTubeURL
		}
		return videoID, nil
	}

	return "", ErrInvalidYouTubeURL
}
