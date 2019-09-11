package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func Append(name, content string) error {
	if f, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		return err
	} else {
		defer func() { _ = f.Close() }()
		if _, err := f.WriteString(fmt.Sprintf("\n%s\n", content)); err != nil {
			return err
		}
	}
	return nil
}

func Write(dir, file, content string) error {
	if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, file)), 0777); err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(dir, file), []byte(strings.TrimSpace(content)), 0666)
}
