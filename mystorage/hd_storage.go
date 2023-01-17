package mystorage

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/94peter/sterna/util"
)

func NewHdStorage(path string) Storage {
	return &hd{Path: path}
}

type hd struct {
	Path string
}

func (hd *hd) getAbsFilePath(filePath string) string {
	absFilePath, _ := filepath.Abs(hd.Path + filePath)
	return absFilePath
}

func (hd *hd) Save(fp string, file []byte) (string, error) {
	absFilePath := hd.getAbsFilePath(fp)
	err := hd.mkdir(absFilePath)
	if err != nil {
		return "", err
	}
	return absFilePath, ioutil.WriteFile(absFilePath, file, 0644)
}

func (hd *hd) SaveByReader(fp string, reader io.Reader) (string, error) {
	absFilePath := hd.getAbsFilePath(fp)
	err := hd.mkdir(absFilePath)
	if err != nil {
		return "", err
	}
	f, err := os.OpenFile(absFilePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = io.Copy(f, reader)
	if err != nil {
		return "", err
	}
	return absFilePath, nil
}

func (hd *hd) Delete(filePath string) error {
	absFilePath := hd.getAbsFilePath(filePath)
	exist, err := fileExist(absFilePath)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New("file not exist: " + absFilePath)
	}
	return os.Remove(absFilePath)
}

func (hd *hd) Get(fp string) ([]byte, error) {
	absFilePath := hd.getAbsFilePath(fp)
	return ioutil.ReadFile(absFilePath)
}

func (hd *hd) mkdir(absPath string) error {
	dir := filepath.Dir(absPath)
	exist, _ := fileExist(dir)

	if !exist {
		err := os.MkdirAll(dir, 0766)
		if err != nil {
			return err
		}
	}
	return nil
}

func (hd *hd) FileExist(fp string) bool {
	absFilePath := hd.getAbsFilePath(fp)
	exist, _ := fileExist(absFilePath)
	return exist
}

func (hd *hd) List(dir string) []string {
	absDir := hd.getAbsFilePath(dir)
	files, err := ioutil.ReadDir(absDir)
	if err != nil {
		return nil
	}
	var result []string
	for _, f := range files {
		if f.IsDir() {
			result = append(result, util.StrAppend(dir, f.Name(), "/"))
		} else {
			result = append(result, util.StrAppend(dir, f.Name()))
		}

	}
	return result
}

func fileExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
