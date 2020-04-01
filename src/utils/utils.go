package utils

import (
	"os"
)

func CreateDir(path string) (err error) {
	if _, err = os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateFile(filename string) (err error) {
	if _, err = os.Stat(filename); os.IsNotExist(err) {
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		f.Close()
	}
	return err
}
