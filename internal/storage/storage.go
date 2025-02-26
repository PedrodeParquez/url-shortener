package storage

import (
	"errors"
)

var (
	ErrURLNotFound  = errors.New("url not found")
	ErrURLExists    = errors.New("url exists")
	ErrDBConnection = errors.New("failed to connect to database")
)
