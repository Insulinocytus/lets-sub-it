package http

import (
	"errors"
	"fmt"
	stdhttp "net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"

	"lets-sub-it/backend/internal/jobs"
)

var ErrInvalidYouTubeURL = errors.New("invalid youtube url")

var youtubeVideoIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

type Handler struct {
	jobs *jobs.Service
}

func (h *Handler) createJob(c *gin.Context) {
	var req CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(stdhttp.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	youtubeURL := strings.TrimSpace(req.YouTubeURL)
	videoID, err := extractVideoID(youtubeURL)
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	job, err := h.jobs.CreateJob(c.Request.Context(), jobs.CreateJobInput{
		VideoID:        videoID,
		YouTubeURL:     youtubeURL,
		TargetLanguage: strings.TrimSpace(req.TargetLanguage),
	})
	if err != nil {
		status := stdhttp.StatusInternalServerError
		if errors.Is(err, jobs.ErrInvalidCreateJobInput) {
			status = stdhttp.StatusBadRequest
		}

		c.JSON(status, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(stdhttp.StatusAccepted, toJobResponse(job))
}

func (h *Handler) getJob(c *gin.Context) {
	job, err := h.jobs.GetJob(c.Request.Context(), c.Param("id"))
	if err != nil {
		status := stdhttp.StatusInternalServerError
		if errors.Is(err, jobs.ErrJobNotFound) {
			status = stdhttp.StatusNotFound
		}

		c.JSON(status, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(stdhttp.StatusOK, toJobResponse(job))
}

func (h *Handler) getAssetByVideoID(c *gin.Context) {
	asset, err := h.jobs.GetAssetByVideoID(c.Request.Context(), c.Param("videoId"))
	if err != nil {
		status := stdhttp.StatusInternalServerError
		if errors.Is(err, jobs.ErrAssetNotFound) {
			status = stdhttp.StatusNotFound
		}

		c.JSON(status, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(stdhttp.StatusOK, toSubtitleAssetResponse(
		asset,
		buildAssetURL(c, asset.SourceVTTPath),
		buildAssetURL(c, asset.TranslatedVTTPath),
		buildAssetURL(c, asset.BilingualVTTPath),
	))
}

func (h *Handler) getAssetByJobID(c *gin.Context) {
	asset, err := h.jobs.GetAssetByJobID(c.Request.Context(), c.Param("id"))
	if err != nil {
		status := stdhttp.StatusInternalServerError
		if errors.Is(err, jobs.ErrAssetNotFound) {
			status = stdhttp.StatusNotFound
		}

		c.JSON(status, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(stdhttp.StatusOK, toSubtitleAssetResponse(
		asset,
		buildAssetURL(c, asset.SourceVTTPath),
		buildAssetURL(c, asset.TranslatedVTTPath),
		buildAssetURL(c, asset.BilingualVTTPath),
	))
}

func buildAssetURL(c *gin.Context, relativePath string) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s/assets/%s", scheme, c.Request.Host, strings.TrimLeft(relativePath, "/"))
}

func extractVideoID(rawURL string) (string, error) {
	trimmedURL := strings.TrimSpace(rawURL)
	if trimmedURL == "" {
		return "", fmt.Errorf("%w: url is required", ErrInvalidYouTubeURL)
	}

	parsed, err := url.Parse(trimmedURL)
	if err != nil {
		return "", fmt.Errorf("%w: parse url: %v", ErrInvalidYouTubeURL, err)
	}
	if !parsed.IsAbs() {
		return "", fmt.Errorf("%w: url must be absolute", ErrInvalidYouTubeURL)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("%w: scheme must be http or https", ErrInvalidYouTubeURL)
	}

	host := strings.ToLower(parsed.Hostname())
	switch host {
	case "youtube.com", "www.youtube.com", "m.youtube.com":
		if parsed.Path != "/watch" {
			return "", fmt.Errorf("%w: unsupported youtube path", ErrInvalidYouTubeURL)
		}

		videoID := strings.TrimSpace(parsed.Query().Get("v"))
		if !youtubeVideoIDPattern.MatchString(videoID) {
			return "", fmt.Errorf("%w: invalid video id", ErrInvalidYouTubeURL)
		}

		return videoID, nil
	case "youtu.be", "www.youtu.be":
		path := strings.Trim(parsed.Path, "/")
		if path == "" || strings.Contains(path, "/") {
			return "", fmt.Errorf("%w: unsupported youtube path", ErrInvalidYouTubeURL)
		}
		if !youtubeVideoIDPattern.MatchString(path) {
			return "", fmt.Errorf("%w: invalid video id", ErrInvalidYouTubeURL)
		}

		return path, nil
	default:
		return "", fmt.Errorf("%w: unsupported youtube host", ErrInvalidYouTubeURL)
	}
}
