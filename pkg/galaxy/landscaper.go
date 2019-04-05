package galaxy

import (
	ldsc "github.com/Eneco/landscaper/pkg/landscaper"
	"github.com/sirupsen/logrus"
)

// Landscaper represents upstream Landscaper.
type Landscaper struct {
	log  *logrus.Logger // logger
	opts *ldsc.Environment
}

// NewLandscaper instance a new Landscaper object.
func NewLandscaper(log *logrus.Logger) *Landscaper {
	return &Landscaper{
		log:  log,
		opts: &ldsc.Environment{},
	}
}
