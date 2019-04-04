package galaxy

import (
	"testing"

	logrus "github.com/sirupsen/logrus"
)

var landscaper *Landscaper

func TestLandscaperNewLandscaper(t *testing.T) {
	landscaper = NewLandscaper(logrus.New())
}
