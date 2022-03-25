package parser

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CsvHeaderParser(t *testing.T) {
	file, err := os.Open("./temp.csv") // For read access.
	assert.Nil(t, err)
	defer file.Close()
	r, err := NewCsvHeaderParser(file, EncodeUtf8)
	fmt.Println(r, err)
	assert.True(t, false)
}

func Test_CsvParser(t *testing.T) {
	file, err := os.Open("./temp.csv") // For read access.
	assert.Nil(t, err)
	defer file.Close()
	r, err := NewCsvParser(file, EncodeUtf8)
	for r.Next() {
		fmt.Println(r.Row().GetVal("mac_address"))
		fmt.Println(r.Row().GetVal("model"))
	}
	//fmt.Println(r, err)
	assert.True(t, false)
}
