package model

import (
	"time"
)

const (
	DefaultFormat        = FormatVideo
	DefaultQuality       = QualityHigh
	DefaultPageSize      = 50
	DefaultUpdatePeriod  = 6 * time.Hour
	DefaultLogMaxSize    = 50 // megabytes
	DefaultLogMaxAge     = 30 // days
	DefaultLogMaxBackups = 7
	PathRegex            = `^[A-Za-z0-9]+$`
)
