package utils

import (
	"archive/zip"
	"bytes"
)

// CreateZip creates a ZIP archive containing a single file
func CreateZip(fileName string, content []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	
	f, err := w.Create(fileName)
	if err != nil {
		return nil, err
	}
	_, err = f.Write(content)
	if err != nil {
		return nil, err
	}
	
	err = w.Close()
	if err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}
