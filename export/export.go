package export

import (
	"io"

	"github.com/94peter/sterna/export/pdf"
)

type ExportConf interface {
	GetPdfConf() *pdf.PdfConf
}

type PDF interface {
	Pdf(w io.Writer) error
}

type MyExportConf struct {
	*pdf.PdfConf `yaml:"pdf"`
}

func (c *MyExportConf) GetPdfConf() *pdf.PdfConf {
	return c.PdfConf
}
