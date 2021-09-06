package export

import (
	"io"
	"sterna/export/pdf"
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
