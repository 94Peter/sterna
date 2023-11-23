package util

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_idtest(t *testing.T) {
	// //
	ok := IsIdNumber("A123456780")
	fmt.Println(ok)
	assert.True(t, false)
}

func Test_SliceSplit(t *testing.T) {
	a := []int{1, 2, 3, 4, 5, 6}
	fmt.Println(a)
	start := 0
	l := 3
	fmt.Println(a[start : start+l])
	assert.True(t, false)
}

func Test_Byte(t *testing.T) {
	a := uint16(1)
	fmt.Println(uint32(a) << 16)
	assert.True(t, false)
}

func Test_ByteToFloat32(t *testing.T) {
	a := uint16(18000)
	b := uint16(8192)
	ba := make([]byte, 2)
	bb := make([]byte, 2)
	binary.BigEndian.PutUint16(ba, a)
	binary.BigEndian.PutUint16(bb, b)
	fmt.Println(ba, bb)

	bf := []byte{ba[0], ba[1], bb[0], bb[1]}

	r := binary.BigEndian.Uint32(bf)
	n := float64(math.Float32frombits(r))
	fmt.Println(r, n)

	assert.True(t, false)
}

func Test_CheckTimeFormat(t *testing.T) {
	teststr := ""
	err := CheckTimeFormat(teststr)
	assert.Nil(t, err)

}

func Test_ReflactTypeOf(t *testing.T) {
	tst := "string"
	tst2 := 10
	tst3 := 1.2

	fmt.Println(reflect.TypeOf(tst))
	fmt.Println(reflect.TypeOf(tst2).Kind() == reflect.Int)
	fmt.Println(reflect.TypeOf(tst3))
	assert.Nil(t, "show")
}

type test struct {
	Ele int
}

func Test_ReturnExist(t *testing.T) {
	var a string
	a = ReturnExist("a", "").(string)
	assert.Equal(t, a, "a")
	a2 := ReturnExist("a", "b")
	assert.Equal(t, a2, "b")

	b := ReturnExist(20, 0)
	assert.Equal(t, b, 20)
	b2 := ReturnExist(20, 30)
	assert.Equal(t, b2, 30)

	fr := test{Ele: 15}
	frn := test{Ele: 30}
	c := ReturnExist(fr, test{})
	assert.Equal(t, c, fr)
	c2 := ReturnExist(fr, frn)
	assert.Equal(t, c2, frn)

	assert.Nil(t, "out")
}

func Test_regexpMatchString(t *testing.T) {
	no := "abcdd##d"
	ok, err := regexp.MatchString(`^[^!@#$%^&*()_+{}|"?:><,./;']*$`, no)
	fmt.Println(ok)
	fmt.Println(err)
	fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	assert.Nil(t, "out")
}

func TestIsRFC3339(t *testing.T) {
	tests := []struct {
		name     string
		timeStr  string
		expected error
	}{
		{
			name:     "Valid RFC3339 time",
			timeStr:  "2022-01-01T12:00:00Z",
			expected: nil,
		},
		{
			name:     "Valid RFC3339 time",
			timeStr:  "2022-01-01T12:00:00+08:00",
			expected: nil,
		},
		{
			name:     "Invalid RFC3339 time",
			timeStr:  "2022-01-01T12:00:00",
			expected: ErrNotRfc3339,
		},
		{
			name:     "Empty time string",
			timeStr:  "",
			expected: ErrNotRfc3339,
		},
		{
			name:     "Invalid format",
			timeStr:  "2022-01-01",
			expected: ErrNotRfc3339,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := IsRFC3339(test.timeStr)
			if err != test.expected {
				t.Errorf("Expected error: %v, but got: %v", test.expected, err)
			}
		})
	}
}
