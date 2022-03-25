package parser

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"strings"

	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
)

type Encode string

const (
	EncodeBig5 = Encode("big5")
	EncodeUtf8 = Encode("utf8")
)

type Parser interface {
	Next() bool
	Row() Row
	Err() error
}

type Row interface {
	GetVal(header string) (string, error)
}

func NewCsvHeaderParser(reader io.Reader, encode Encode) ([]string, error) {
	scanner := bufio.NewScanner(reader)

	scanTxt := ""
	var header []string
	if scanner.Scan() {
		scanTxt = scanner.Text()
		if encode == EncodeBig5 {
			big5ToUTF8 := traditionalchinese.Big5.NewDecoder()
			scanTxt, _, _ = transform.String(big5ToUTF8, scanTxt)
		}
		trimmedBytes := bytes.Trim([]byte(scanTxt), "\xef\xbb\xbf")
		scanTxt = strings.ReplaceAll(string(trimmedBytes), "\"", "")
		header = strings.Split(scanTxt, ",")
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return header, nil
}

func NewCsvParser(reader io.Reader, encode Encode) (Parser, error) {
	scanner := bufio.NewScanner(reader)
	result := &basicParser{
		headerMap: make(map[string]int),
		encode:    encode,
		Scanner:   scanner,
	}

	if scanner.Scan() {
		scanTxt := scanner.Text()
		if encode == EncodeBig5 {
			big5ToUTF8 := traditionalchinese.Big5.NewDecoder()
			scanTxt, _, _ = transform.String(big5ToUTF8, scanTxt)
		}
		trimmedBytes := bytes.Trim([]byte(scanTxt), "\xef\xbb\xbf")
		scanTxt = strings.ReplaceAll(string(trimmedBytes), "\"", "")

		result.header = strings.Split(scanTxt, ",")
		for i, v := range result.header {
			result.headerMap[v] = i
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

type basicParser struct {
	headerMap map[string]int
	header    []string

	data []string
	err  error

	encode Encode
	file   *os.File
	*bufio.Scanner
}

func (r *basicParser) Row() Row {
	return r
}

func (r *basicParser) Err() error {
	return r.err
}

func (r *basicParser) Next() bool {
	if r.Scan() {
		scanTxt := r.Text()
		if r.encode == EncodeBig5 {
			big5ToUTF8 := traditionalchinese.Big5.NewDecoder()
			scanTxt, _, _ = transform.String(big5ToUTF8, scanTxt)
		}
		r.data = strings.Split(scanTxt, ",")
	} else {
		return false
	}
	r.err = r.Err()

	return true
}

func (r *basicParser) GetVal(name string) (string, error) {
	if len(r.headerMap) == 0 || len(r.data) == 0 {
		return "", nil
	}
	var col int
	var ok bool

	if col, ok = r.headerMap[name]; !ok {
		return "", errors.New("not found " + name)
	}
	return r.data[col], nil
}
