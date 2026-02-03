package upload

import "errors"

var (
	ErrUploadFailed     = errors.New("upload failed")
	ErrUploadMissingURL = errors.New("upload failed: missing url")
	ErrInvalidURLScheme = errors.New("invalid url scheme")
	ErrUploadInvalidURL = errors.New("invalid url")
)
