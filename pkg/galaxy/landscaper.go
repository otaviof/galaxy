package galaxy

import (
	ldsc "github.com/Eneco/landscaper/pkg/landscaper"
	"github.com/sirupsen/logrus"
)

type Landscaper struct {
	log  *logrus.Logger // logger
	opts *ldsc.Environment
}

func NewLandscaper(log *logrus.Logger) *Landscaper {
	return &Landscaper{
		log:  log,
		opts: &ldsc.Environment{},
	}
}
