package app

import "github.com/pkg/errors"

var (
	ErrInvalidCommand = errors.New("invalid command")
	ErrInvalidEvent   = errors.New("invalid event")
)
