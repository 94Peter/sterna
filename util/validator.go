// Package util provides a variety of handy functions while developing backends
package util

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
	"unicode"

	"github.com/asaskevich/govalidator"
)

// InitValidator is used to initialize the validator.
// Call this funtion before using any validator.
func InitValidator() {
	govalidator.TagMap["required"] = govalidator.Validator(required)
	return
}

// CheckParam can help to verify if the input is valid or not.
//
// Input:
//  s      : input map for validation.
//  target : A map indicates that what validation should be applied to the input.
// Output:
//  bool   : Return true if valid, otherwise a false is returned.
//  error  : the output of the detail of the error.
func CheckParam(param map[string]interface{}, target map[string][]string) (bool, map[string]string) {
	var res bool = true
	var detail string
	var errOutput map[string]string
	errOutput = make(map[string]string)
	for key, value := range target {
		if input, ok := param[key]; ok {
			res, detail = checkParamWithArray(input, value)
			if !res {
				errOutput[key] = detail
			}
		}
	}
	if len(errOutput) > 0 {
		return false, errOutput
	}
	return true, errOutput
}

func CheckRequiredAndParam(param map[string]interface{}, requireFields []string, target map[string][]string) (bool, map[string]string) {
	var ok bool
	errOutput := make(map[string]string)
	for _, field := range requireFields {
		if _, ok = param[field]; !ok {
			errOutput[field] = "not exist"
		}
	}
	if len(errOutput) > 0 {
		return false, errOutput
	}
	return CheckParam(param, target)
}

// checkParamWithArray helps CheckParam to verify the inputs recursively.
// It should not be called by other funcions.
// Currently supported tags:  DNS,Alphanumeric,Numeric,Alpha,Email.
func checkParamWithArray(param interface{}, target []string) (bool, string) {
	var res bool = true
	var detail string
	var subString string
	//The last one
	if input, ok := param.(string); ok {
		for i := 0; i < len(target); i++ {
			switch test := target[i]; test {
			case "MAC":
				res = govalidator.IsMAC(input)
				break
			case "DNS":
				res = govalidator.IsDNSName(input)
				break
			case "Alphanumeric":
				res = govalidator.IsAlphanumeric(input)
				break
			case "Numeric":
				res = govalidator.IsNumeric(input)
				break
			case "Alpha":
				res = govalidator.IsAlpha(input)
				break
			case "Email":
				res = govalidator.IsEmail(input)
				break
			case "Required":
				res = required(input)
				break
			case "MongoID":
				res = govalidator.IsMongoID(input)
				break
			case "Bool":
				res = isBool(input)
				break
			default:
				subString = fmt.Sprintf("%s %s\n", "No such validator: ", target[i])
				res = false
				break
			}
			if !res {
				subString = "[Check " + input + " with " + target[i] + " failed]"
				detail += subString
			}
		}
	} else if input, ok := param.([]string); ok { //Recursively disemble the []string
		if len(input) <= 0 {
			return true, detail
		}
		pop := input[0]
		res, subString = checkParamWithArray(pop, target)
		detail += subString
		input = input[1:]
		res, subString = checkParamWithArray(input, target)
		detail += subString
	}
	if len(detail) == 0 {
		return true, detail
	} else {
		return false, detail
	}

}

// IsInt can help to verify if the input is Int and the value is within the range or not.
//
// If the bound array is not 2, it will skip the ranging test
func IsInt(param string, bound []int) (int, error) {
	//Check if the content type is match with the target
	var res bool = false
	res = govalidator.IsInt(param)
	if !res {
		return -1, errors.New("Input is not a Integer")
	}
	num, err := strconv.Atoi(param)
	if err != nil {
		return -1, err
	}
	if bound == nil {
		return num, nil
	}
	if len(bound) == 2 {
		res = govalidator.InRangeInt(num, bound[0], bound[1])
		if !res {
			return num, errors.New("Input is out of range")
		}
	}
	return num, nil
}

// IsFloat64 can help to verify if the input is Float64 and the value is within the range or not.
//
// If the bound array is not 2, it will skip the ranging test
func IsFloat64(param string, bound []float64) (float64, error) {
	//Check if the content type is match with the target
	var res bool = false
	res = govalidator.IsFloat(param)
	if !res {
		return -1, errors.New("Input is not a Float")
	}
	num, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return -1, err
	}
	if len(bound) == 2 {
		res = govalidator.InRange(num, bound[0], bound[1])
		if !res {
			return num, errors.New("Input is out of range")
		}
	}
	return num, nil
}

//IsBool transform strings of true and false into bool.
func IsBool(str string) (bool, error) {
	return strconv.ParseBool(str)
}

// CheckStruct can help to verify the content of the struct with tags.
// It accepts recursive structs.
//
// Remember to Add valid:"[tags]" to the struct.
func CheckStruct(s interface{}) (bool, error) {
	result, err := govalidator.ValidateStruct(s)
	return result, err
	//BUG(Kevin Xu): haha
}

//IsStrInList can help to see if input is one of the target or not.
//a non-empty error will be returned if input does match any of the target.
func IsStrInList(input string, target ...string) bool {
	for _, paramName := range target {
		if input == paramName {
			return true
		}
	}
	return false

}

//isMail can help to verify if the input is Email format or not
func IsMail(param string) (bool, error) {
	res := govalidator.IsEmail(param)
	return res, errors.New("Nothing")
}

//Check if str is empty or not
func required(str string) bool {
	if len(str) > 0 {
		return true
	}
	return false
}

//Check bool
func isBool(str string) bool {
	_, err := strconv.ParseBool(str)
	if err != nil {
		return false
	}
	return true
}

func IsAlpha(str string) bool {
	return govalidator.IsAlpha(str)
}

func IsValidPwd(str string) (bool, error) {
	if len(str) < 6 {
		return false, errors.New("Minimum 6 characters in length")
	}
	hasLower, hasUpper, hasPunct, hasNumber := false, false, false, false
	for _, char := range str {
		if unicode.IsLower(char) && char != ' ' {
			hasLower = true
		} else if unicode.IsUpper(char) && char != ' ' {
			hasUpper = true
		} else if unicode.IsPunct(char) {
			hasPunct = true
		} else if unicode.IsNumber(char) {
			hasNumber = true
		}
	}
	if !hasLower {
		return false, errors.New("Must contain Lowercase Letters")
	}
	if !hasUpper {
		return false, errors.New("Must contain Uppercase Letters")
	}
	if !hasPunct {
		return false, errors.New("Must contain Symbol")
	}
	if !hasNumber {
		return false, errors.New("Must contain Number")
	}
	return true, nil
}

var (
	codeMap = map[byte][]int{
		'A': {1, 0}, 'B': {1, 1}, 'C': {1, 2},
		'D': {1, 3}, 'E': {1, 4}, 'F': {1, 5},
		'G': {1, 6}, 'H': {1, 7}, 'I': {1, 8},
		'J': {1, 9}, 'K': {1, 10}, 'L': {1, 11},
		'M': {1, 12}, 'N': {1, 13}, 'O': {1, 14},
		'P': {1, 15}, 'Q': {1, 16}, 'R': {1, 17},
		'S': {1, 18}, 'T': {1, 19}, 'U': {1, 20},
		'V': {1, 21}, 'W': {1, 22}, 'X': {1, 23},
		'Y': {1, 24}, 'Z': {1, 25},
	}

	multiplier = []int{1, 9, 8, 7, 6, 5, 4, 3, 2, 1, 1}
)

// 是否為身份證字號
func IsIdNumber(id string) bool {
	if m, _ := regexp.MatchString(`^[A-Z][12]\d{8}$`, id); !m {
		return false
	}
	a, ok := codeMap[id[0]]
	if !ok {
		return false
	}

	sum := 0

	for i := 0; i < 2; i++ {
		sum += a[i] * multiplier[i]
	}

	l := len(id)
	for i := 1; i < l; i++ {
		sum += int(id[i]-'0') * multiplier[i+1]
	}

	return sum%10 == 0
}

// 驗證統一編號
func IsVATnumber(num string) bool {
	mul := []int{1, 2, 1, 2, 1, 2, 4, 1}
	total := 0
	minus := false
	intsum := func(a int) int {
		for a >= 10 {
			a = (a / 10) + (a % 10)
		}
		return a
	}
	for ind, ele := range num {
		val, _ := strconv.Atoi(string(ele))
		total += intsum(val * mul[ind])
		if ind == 6 && val == 7 {
			minus = true
		}
	}

	if (total % 10) == 0 {
		return true
	} else if minus && (total%10) == 1 {
		return true
	} else {
		return false
	}
}

//CheckTimeFormat is use to check if the string is correct time format or not
func CheckTimeFormat(tim string) error {
	timTime, err := time.Parse(time.RFC3339, tim)
	if err != nil {
		return err
	}
	timStr := timTime.Format(time.RFC3339)

	if timStr != tim {
		return errors.New("unexpected value but correct format of time")
	}
	return nil
}

func IsMobileNum(number string) bool {
	if len(number) != 10 {
		return false
	}
	if (number[0] != '0') && (number[1] != '9') {
		return false
	}
	// ^ start of string, \d digit, {10} for ten, $ end of string
	if state, _ := regexp.MatchString(`^\d{10}$`, number); !state {
		return false
	}
	return true
}

func IsHomeNum(number string) bool {
	length := len(number)
	if !(length > 8 && length < 11) {
		return false
	}
	if number[0] != '0' {
		return false
	}
	pattern := fmt.Sprintf(`^\d{%s}$`, strconv.Itoa(length))
	if state, _ := regexp.MatchString(pattern, number); !state {
		return false
	}
	return true
}

func IsLegalPhoneNumber(num string) (bool, error) {
	if len(num) != 13 && len(num) != 12 && len(num) != 11 && len(num) != 10 {
		return false, errors.New("Number' format is like 0935120080,02XXXXXXXX,+8862XXXXXXXX,+886935120080")
	}
	if len(num) == 10 {
		if m, _ := regexp.MatchString(`^0[2-9]\d{8}$`, num); !m {
			return false, errors.New("Number' format is like 0935120080,02XXXXXXXX")
		}
		return true, nil
	}
	if x, _ := regexp.MatchString(`^\+(\d{1,3})[2-9](\d{8})$`, num); !x {
		return false, errors.New("Number' format is like +8862XXXXXXXX,+886935120080")
	}
	return true, nil
}
