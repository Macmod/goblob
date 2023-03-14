package utils

import (
    "regexp"
    "fmt"
    "os"
    "bufio"
    "net/http"
)

const (
    ACCOUNT_PATTERN = "^[a-z0-9]*$"
    ENTITY_PATTERN = "^[a-zA-Z][a-zA-Z0-9-]*$"
    ENTITY_URL_PATTERN = "<Url>([^<]+)"
)

var REGEXP_ENTITY_URL = regexp.MustCompile(ENTITY_URL_PATTERN)
var HttpClient = &http.Client{
    Transport: &http.Transport{
        DisableKeepAlives: false,
    },
}

func IsValidEntityName(entityName string) bool {
    if len(entityName) < 3 || len(entityName) > 63 {
        return false
    }

    match, _ := regexp.MatchString(ENTITY_PATTERN, entityName)

    return match
}

func IsValidStorageAccountName(name string) bool {
    if len(name) < 3 || len(name) > 24 {
        return false
    }

    match, _ := regexp.MatchString(ACCOUNT_PATTERN, name)

    return match
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
