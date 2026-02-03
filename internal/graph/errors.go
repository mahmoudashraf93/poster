package graph

import "errors"

var (
	ErrGraphClientNil     = errors.New("graph client is nil")
	ErrGraphAPIStatus     = errors.New("graph api returned non-2xx")
	ErrPollTimeout        = errors.New("poll timed out")
	ErrMediaProcessing    = errors.New("media processing failed")
	ErrMissingID          = errors.New("missing id in response")
	ErrEmptyID            = errors.New("empty id in response")
	ErrUnexpectedIDType   = errors.New("unexpected id type in response")
	ErrMissingAccessToken = errors.New("missing access_token in response")
	ErrMissingTokenData   = errors.New("missing data in response")
	ErrMissingIGAccount   = errors.New("missing instagram_business_account in response")
	ErrMissingIGAccountID = errors.New("missing instagram_business_account id")
)
