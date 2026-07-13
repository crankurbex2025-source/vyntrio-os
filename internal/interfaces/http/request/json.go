package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const defaultMaxJSONBodyBytes = 8 * 1024

// ErrInvalidJSON indicates the request body is not valid JSON for the endpoint.
var ErrInvalidJSON = errors.New("invalid json request body")

// ErrInvalidContentType indicates the request Content-Type is not application/json.
var ErrInvalidContentType = errors.New("invalid content type")

// ErrBodyTooLarge indicates the request body exceeds the allowed size.
var ErrBodyTooLarge = errors.New("request body too large")

// DecodeStrictJSON reads and decodes one JSON object with unknown-field rejection.
func DecodeStrictJSON(r *http.Request, maxBytes int64, dst any) error {
	if maxBytes <= 0 {
		maxBytes = defaultMaxJSONBodyBytes
	}

	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		return ErrInvalidContentType
	}
	mediaType, _, err := parseMediaType(contentType)
	if err != nil || mediaType != "application/json" {
		return ErrInvalidContentType
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxBytes+1))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}
	if int64(len(body)) > maxBytes {
		return ErrBodyTooLarge
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return ErrInvalidJSON
		}
		return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	return nil
}

func parseMediaType(value string) (string, string, error) {
	parts := strings.SplitN(strings.TrimSpace(value), ";", 2)
	mediaType := strings.ToLower(strings.TrimSpace(parts[0]))
	if len(parts) == 1 {
		return mediaType, "", nil
	}
	return mediaType, strings.TrimSpace(parts[1]), nil
}
