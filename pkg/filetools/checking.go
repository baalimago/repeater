package filetools

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// CheckAmLines by parsing the file and counting the amount of lines using
// a scanner. Will return error if the file fails to open or if the scanner fails somehow
// On error, am will be -1
func CheckAmLines(filePath string) (int, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return -1, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()
	// Solution joinked from: https://stackoverflow.com/questions/24562942/golang-how-do-i-determine-the-number-of-lines-in-a-file-efficiently
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := f.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}
