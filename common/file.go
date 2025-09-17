package common

import (
	"bufio"
	"fmt"
	"os"
	"path"
)

func CacheDir() (string, error) {
	p, err := os.UserCacheDir()
	if err != nil {
		return p, err
	}
	p = path.Join(p, "goorphans")
	return p, os.MkdirAll(p, 0o700)
}

func ReadFileLines(name string) ([]string, error) {
	var lines []string
	var file *os.File
	if name == "-" {
		file = os.Stdin
	} else {
		var err error
		file, err = os.Open(name)
		if err != nil {
			return lines, err
		}
		defer file.Close()
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func WriteFileLines(name string, lines []string) error {
	var file *os.File
	if name == "-" {
		file = os.Stdout
	} else {
		var err error
		file, err = os.Create(name)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := fmt.Fprintln(writer, line)
		if err != nil {
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	return nil
}
