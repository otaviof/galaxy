package galaxy

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var printer *Printer

func TestPrinterNew(t *testing.T) {
	SetLogLevel("trace")

	dotGalaxy, _ := NewDotGalaxy("../../test/galaxy.yaml")
	g := NewGalaxy(dotGalaxy, NewConfig())
	g.Plan()
	printer = NewPrinter(g.Modified)
}

func TestPrinterTree(t *testing.T) {
	tree := printer.Tree()
	fmt.Println(tree)
}

func TestPrinterTable(t *testing.T) {
	table := printer.Table()
	fmt.Println(table)
	assert.True(t, len(strings.Split(table, "\n")) > 2)
}
