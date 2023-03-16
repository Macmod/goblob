package utils

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"regexp"
)

const (
	ACCOUNT_PATTERN    = "^[a-z0-9]*$"
	CONTAINER_PATTERN     = "^[a-zA-Z][a-zA-Z0-9-]*$"
	BLOB_URL_PATTERN = "<Url>([^<]+)"
)

var Empty struct{}

var REGEXP_BLOB_URL = regexp.MustCompile(BLOB_URL_PATTERN)
var REGEXP_CONTAINER = regexp.MustCompile(CONTAINER_PATTERN)
var REGEXP_ACCOUNT = regexp.MustCompile(ACCOUNT_PATTERN)

func IsValidContainerName(name string) bool {
	if len(name) < 3 || len(name) > 63 {
		return false
	}

	match := REGEXP_CONTAINER.MatchString(name)

	return match
}

func IsValidStorageAccountName(name string) bool {
	if len(name) < 3 || len(name) > 24 {
		return false
	}

	match := REGEXP_ACCOUNT.MatchString(name)

	return match
}

func ReadLines(filename string) []string {
	var results []string

	fileObj, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer fileObj.Close()

	scanner := bufio.NewScanner(fileObj)

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}

	for scanner.Scan() {
		results = append(results, scanner.Text())
	}

	return results
}

func FormatSize(inputSize int64) string {
	var size float64 = float64(inputSize)
	var base float64 = 1024.0
	var idx int

	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	if size == 0 {
		return "0 B"
	}

	idx = int(math.Floor(math.Log(size) / math.Log(base)))
	if idx >= len(units) {
		idx = len(units) - 1
	}

	unit := units[idx]
	return fmt.Sprintf("%.1f %s", size/math.Pow(base, float64(idx)), unit)
}
