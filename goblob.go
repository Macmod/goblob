package main

import (
	"flag"
	"fmt"
	"strings"
	"context"
	"bufio"
	"regexp"
	"io/ioutil"
	"os"
	"sync"
	"github.com/Macmod/goblob/utils"
)

const (
	Reset  = "\033[0m"
	Red	= "\033[31m"
	Green  = "\033[32m"
)

var REGEXP_NEXT_MARKER = regexp.MustCompile("<NextMarker>([^<]+)")

type Message struct {
	textToStdout string
	textToFile string
}

//var REGEX_ERROR_CODE = regexp.MustCompile("<Code>([^<]+)")

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run goblob.go <targetaccount>")
		os.Exit(1)
	}

	accountsFilename := flag.String(
		"accounts", "", "File with target Azure storage account names to check",
	)
	containersFilename := flag.String(
		"containers", "wordlists/gosab-folder-names.txt",
		"Wordlist file with possible container names for Azure blob storage",
	)
	maxGoroutines := flag.Int(
		"goroutines", 5000,
		"Maximum of concurrent goroutines",
	)
	output := flag.String(
		"output", "",
		"Save found URLs to output file",
	)
	verbose := flag.Int("verbose", 0, "Verbosity level (default=0)")
	blobs := flag.Bool(
		"blobs", false,
		"Show each blob URL in the results instead of their container URLs",
	)

	flag.Parse()

	// Import input from files
	var accounts []string
	if *accountsFilename != "" {
		accounts = utils.ReadLines(*accountsFilename)
	} else {
		accounts = []string{os.Args[1]}
	}

	var containers []string = utils.ReadLines(*containersFilename)

	// Results report
	resultEntities := make(map[string][]string)

	printResults := func(result *map[string][]string) {
		fmt.Println("[+] Results:")
		if len(*result) != 0 {
			numFiles := 0
			for key, value := range *result {
				fmt.Printf("[+] %s - %d files\n", key, len(value))
				numFiles += len(value)
			}
	
			fmt.Printf(
				"%s[+] Found a total of %d files across %d account(s)%s\n",
				Green, numFiles, len(*result), Reset,
			)
		} else {
			fmt.Printf("%s[-] No files found.%s\n", Red, Reset)
		}
	}

	if *verbose > 0 {
		defer printResults(&resultEntities)
	}

	// Requests context and synchronization stuff
	ctx := context.Background()
	semaphore := make(chan struct{}, *maxGoroutines)
	var wg sync.WaitGroup

	var writer *bufio.Writer
	if *output != "" {
		output_file, _ := os.OpenFile(*output, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		defer output_file.Close()

		writer = bufio.NewWriter(output_file)
	} else {
		writer = nil
	}

	// Dedicated goroutine for writing results
	outputChannel := make(chan Message)
	go func(writer *bufio.Writer, msgChannel chan Message) {
		for {
			msg := <-msgChannel
			if msg.textToStdout != "" {
				fmt.Printf(msg.textToStdout)
			}

			if msg.textToFile != "" {
				if writer != nil {
					writer.WriteString(msg.textToFile + "\n")
					writer.Flush()
				}
			}
		}
	}(writer, outputChannel)

	checkAzureBlobs := func(account string, containerName string) {
		defer func() {
			<-semaphore
			wg.Done()
		}()

		containerURL := fmt.Sprintf(
			"https://%s.blob.core.windows.net/%s?restype=container&comp=list&showonly=files",
			account,
			containerName,
		)

		resp, err := utils.Fetch(containerURL, ctx)
		if err != nil {
			if *verbose > 1 {
				fmt.Printf("%s[-] Error while fetching URL: '%s'%s\n", Red, err, Reset)
			}
		} else {
			defer resp.Body.Close()

			resBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				if *verbose > 1 {
					fmt.Printf("%s[-] Error while reading response body: '%s'%s\n", Red, err, Reset)
				}
			}

			statusCode := resp.StatusCode
			if statusCode < 400 {
				if !*blobs {
					outputChannel <- Message{
						fmt.Sprintf("%s[+][%d] %s%s\n", Green, statusCode, containerURL, Reset),
						containerURL,
					}
				}

				blobURLs := utils.GetBlobURLs(resBody)
				resultEntities[account] = append(
					resultEntities[account],
					blobURLs...
				)

				if *blobs {
					for _, blobURL := range blobURLs {
						outputChannel <- Message{
							fmt.Sprintf("%s[+][%d] %s%s\n", Green, statusCode, blobURL, Reset),
							blobURL,
						}	
					}
				}

				markerMatch := REGEXP_NEXT_MARKER.FindSubmatch(resBody)
				for markerMatch != nil && len(markerMatch) > 1 {
					markerCode := markerMatch[1]
					containerURLWithMarker := fmt.Sprintf("%s&marker=%s", containerURL, markerCode)

					resp, err := utils.Fetch(containerURLWithMarker, ctx)
					if err != nil {
						if *verbose > 1 {
							fmt.Printf("%s[-] Error while fetching URL: '%s'%s\n", Red, err, Reset)
						}
					} else {
						defer resp.Body.Close()

						resBody, err := ioutil.ReadAll(resp.Body)
						if err != nil {
							if *verbose > 1 {
								fmt.Printf("%s[-] Error while reading response body: '%s'%s\n", Red, err, Reset)
							}
							break
						} else {
							blobURLs := utils.GetBlobURLs(resBody)
							resultEntities[account] = append(
								resultEntities[account],
								blobURLs...
							)

							if *blobs {
								for _, blobURL := range blobURLs {
									outputChannel <- Message{
										fmt.Sprintf("%s[+][%d] %s%s\n", Green, statusCode, blobURL, Reset),
										blobURL,
									}
								}
							}

							markerMatch = REGEXP_NEXT_MARKER.FindSubmatch(resBody)
						}
					}
				}
			} else {
				if *verbose > 2 {
					fmt.Printf("%s[+][%d] %s%s\n", Red, statusCode, containerURL, Reset)
				}
			}
		}
	}

	// Main loop
	for idx, account := range accounts {
		account = strings.ToLower(account)
		if !utils.IsValidStorageAccountName(account) {
			if *verbose > 0 {
				fmt.Printf("[~][%d] Skipping invalid storage account name '%s'\n", idx, account)
			}
			continue
		}

		if *verbose > 0 {
			fmt.Printf("[~][%d] Searching blob containers in storage account %s\n", idx, account)
		}

		for _, containerName := range containers {
			containerName = strings.ToLower(containerName)
			if !utils.IsValidEntityName(containerName) {
				continue
			}

			wg.Add(1)
			semaphore <- struct{}{}

			go checkAzureBlobs(account, containerName)
		}
	}

	wg.Wait()
}
