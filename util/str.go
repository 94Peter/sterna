package util

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

const (
	SymbolComma           = ","
	SymbolSingleQuotation = "'"
)

func StrAppend(strs ...string) string {
	var buffer bytes.Buffer
	for _, str := range strs {
		buffer.WriteString(str)
	}
	return buffer.String()
}

func JoinStrWithQuotation(separateSymbol string, quotation string, strs ...string) string {
	var buffer bytes.Buffer
	for _, code := range strs {
		buffer.WriteString(quotation)
		buffer.WriteString(code)
		buffer.WriteString(quotation)
		buffer.WriteString(separateSymbol)
	}
	buffer.Truncate(buffer.Len() - 1)
	return buffer.String()
}

func ToStrAry(input interface{}) []string {
	switch dtype := reflect.TypeOf(input).String(); dtype {
	case "string":
		str := input.(string)
		if str != "" {
			return []string{str}
		}
	case "[]string":
		return input.([]string)
	}
	return []string{}
}

func IntToFixStrLen(val int, length int) (string, error) {
	t := strconv.Itoa(val)
	valLen := len(t)
	if valLen > length {
		return "", errors.New(fmt.Sprintf("value %d is too long.", val))
	} else if valLen == length {
		return t, nil
	}

	returnStr := ""
	overLength := length - valLen
	for i := 0; i < overLength; i++ {
		returnStr = StrAppend(returnStr, "0")
	}
	return StrAppend(returnStr, t), nil
}

// 西元轉中華民國
func ADtoROC(adStr, format string) (TW_Date string, err error) {
	TWyear, err := strconv.Atoi(adStr[0:4])
	fmt.Println(TWyear)
	TWyear = TWyear - 1911
	TWmonth, err := strconv.Atoi(adStr[5:7])
	TWday := 1
	if len(adStr) > 7 {
		TWday, err = strconv.Atoi(adStr[8:10])
	}

	switch format {
	case "ch":
		TW_Date = fmt.Sprintf("%d年%d月%d日", TWyear, TWmonth, TWday)
		break
	case "file":
		TW_Date = fmt.Sprintf("%d%d", TWyear, TWmonth)
		break
	case "invoice":
		if TWmonth%2 == 0 {
			TWmonth = TWmonth - 1
		}
		TW_Date = fmt.Sprintf("%d年%d月-%d月", TWyear, TWmonth, TWmonth+1)
		break
	default:
		TW_Date = fmt.Sprintf("%d/%s/%s", TWyear, adStr[5:7], adStr[8:10])
		break
	}
	return
}

func RemoveScriptTag(htmlStr string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return "", err
	}
	removeScript(doc)
	doc, err = body(doc)
	if err != nil {
		return "", err
	}
	buf := bytes.NewBuffer([]byte{})

	if err := html.Render(buf, doc); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func body(doc *html.Node) (*html.Node, error) {
	var body *html.Node
	var crawler func(*html.Node)
	crawler = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "body" {
			body = node
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			crawler(child)
		}
	}
	crawler(doc)
	if body != nil {
		return body, nil
	}
	return nil, errors.New("Missing <body> in the node tree")
}

func removeScript(n *html.Node) {
	// if note is script tag
	if n.Type == html.ElementNode && n.Data == "script" {
		n.Parent.RemoveChild(n)
		return // script tag is gone...
	}
	// traverse DOM
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		removeScript(c)
	}
}
