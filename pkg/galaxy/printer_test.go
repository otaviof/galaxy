package galaxy

import (
	"testing"
)

var printer *Printer

func TestPrinterNew(t *testing.T) {
	printer = NewPrinter(populatedContext(t))
}

func TestPrinterTree(t *testing.T) {
	printer.Tree()
}
