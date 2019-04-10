package galaxy

import (
	ldsc "github.com/Eneco/landscaper/pkg/landscaper"
	log "github.com/sirupsen/logrus"
)

// Landscaper represents upstream Landscaper.
type Landscaper struct {
	logger *log.Entry // logger
	opts   *ldsc.Environment
}

// NewLandscaper instance a new Landscaper object.
func NewLandscaper() *Landscaper {
	return &Landscaper{
		logger: log.WithField("type", "landscaper"),
		opts:   &ldsc.Environment{},
	}
}
