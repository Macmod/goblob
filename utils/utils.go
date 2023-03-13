package utils

import (
    "regexp"
    "fmt"
    "os"
    "bufio"
	"net/http"
	"context"
)

const ENTITY_PATTERN = "^[a-zA-Z][a-zA-Z0-9-]*$"
const ENTITY_URL_PATTERN = "<Url>([^<]+)"

var REGEXP_ENTITY_URL = regexp.MustCompile(ENTITY_URL_PATTERN)

func Fetch(url string, ctx context.Context) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	
	return resp, nil
}

func IsValidEntityName(entityName string) bool {
    if len(entityName) < 3 || len(entityName) > 63 {
        return false
    }
    
    match, err := regexp.MatchString(ENTITY_PATTERN, entityName)
    if err != nil {
        fmt.Println("[-] Error:", err)
        return false
    }
    
    return match
}

func IsValidStorageAccountName(name string) bool {
    if len(name) < 3 || len(name) > 24 {
        return false
    }
    
    match_allowed, _ := regexp.MatchString("^[a-z0-9]*$", name)
    match_first, _ := regexp.MatchString("^[a-z0-9]", name)
    match_consecutive_dash, _ := regexp.MatchString("-{2,}", name)
    
    if !match_allowed || !match_first || match_consecutive_dash {
        return false
    }
    
    return true
}

func GetBlobURLs(containerXML []byte) []string {
	var matches []string

	urlsMatches := REGEXP_ENTITY_URL.FindAllSubmatch(containerXML, -1)
	for _, urlMatches := range urlsMatches {
		matches = append(
			matches,
			string(urlMatches[1]),
		)
	}

	return matches
}
func ReadLines(filename string) []string {
    var results []string
    
    file_obj, err := os.Open(filename)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    defer file_obj.Close()
    
    scanner := bufio.NewScanner(file_obj)
    
    if err := scanner.Err(); err != nil {
        fmt.Println(err)
    }
    
    for scanner.Scan() {
        results = append(results, scanner.Text())
    }
    
    return results
}
