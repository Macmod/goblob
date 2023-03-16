package core

import (
	"bytes"
	"fmt"
	"github.com/Macmod/goblob/xml"
	"io"
	"net/http"
	"sync"
)

type Message struct {
	textToStdout string
	textToFile   string
}

type ContainerScanner struct {
	httpClient    *http.Client
	wg            *sync.WaitGroup
	semaphore     chan struct{}
	outputChannel chan Message
	resultsMap    *ResultsMap
	blobsOnly     bool
	verboseMode   int
	maxPages      int
	invertSearch bool
}

func (cs *ContainerScanner) Init(
	httpClient *http.Client,
	maxGoroutines int,
	outputChannel chan Message,
	resultsMap *ResultsMap,
	blobsOnly bool,
	verboseMode int,
	maxPages int,
	invertSearch bool,
) {
	cs.httpClient = httpClient
	cs.outputChannel = outputChannel
	cs.resultsMap = resultsMap
	cs.blobsOnly = blobsOnly
	cs.verboseMode = verboseMode
	cs.maxPages = maxPages

	cs.invertSearch = invertSearch
	cs.semaphore = make(chan struct{}, maxGoroutines)
	cs.wg = new(sync.WaitGroup)
}

func (cs *ContainerScanner) Done() {
	cs.wg.Wait()
	close(cs.outputChannel)
}

func (cs *ContainerScanner) ScanContainer(account string, containerName string, done chan struct{}) {
	defer func() {
		<-cs.semaphore
		cs.wg.Done()
		done <- struct{}{}
	}()

	containerURL := fmt.Sprintf(
		"https://%s.blob.core.windows.net/%s?restype=container",
		account,
		containerName,
	)

	checkResp, err := cs.httpClient.Get(containerURL)
	if err != nil {
		if cs.verboseMode > 1 {
			fmt.Printf("%s[-] Error while fetching URL: '%s'%s\n", Red, err, Reset)
		}
	} else {
		defer checkResp.Body.Close()

		checkStatusCode := checkResp.StatusCode
		if checkStatusCode < 400 {
			cs.resultsMap.StoreContainerResults(
				account,
				containerName,
				0, 0,
			)

			if !cs.blobsOnly {
				cs.outputChannel <- Message{
					fmt.Sprintf("%s[+][C=%d] %s%s\n", Green, checkStatusCode, containerURL, Reset),
					containerURL,
				}
			}

			markerCode := "FirstPage"
			page := 1
			for markerCode != "" && (cs.maxPages == -1 || page <= cs.maxPages) {
				if cs.verboseMode > 0 {
					fmt.Printf(
						"[~] Analyzing container '%s' in account '%s' (page %d)\n",
						containerName,
						account,
						page,
					)
				}

				var containerURLWithMarker string
				if page == 1 {
					containerURLWithMarker = fmt.Sprintf("%s&comp=list&showonly=files", containerURL)
				} else {
					containerURLWithMarker = fmt.Sprintf("%s&comp=list&showonly=files&marker=%s", containerURL, markerCode)
				}

				resp, err := cs.httpClient.Get(containerURLWithMarker)
				if err != nil {
					if cs.verboseMode > 1 {
						fmt.Printf("%s[-] Error while fetching URL: '%s'%s\n", Red, err, Reset)
					}
				} else {
					statusCode := resp.StatusCode
					defer resp.Body.Close()

					resBuf := new(bytes.Buffer)
					_, err = io.Copy(resBuf, resp.Body)
					if err != nil {
						if cs.verboseMode > 1 {
							fmt.Printf("%s[-] Error while reading response body: '%s'%s\n", Red, err, Reset)
						}
						break
					}

					if statusCode < 400 {
						resultsPage := new(xml.EnumerationResults)
						resultsPage.LoadXML(resBuf.Bytes())

						blobURLs := resultsPage.BlobURLs()
						cs.resultsMap.StoreContainerResults(
							account,
							containerName,
							len(blobURLs),
							resultsPage.TotalContentLength(),
						)

						if cs.blobsOnly {
							for _, blobURL := range blobURLs {
								cs.outputChannel <- Message{
									fmt.Sprintf("%s[+] %s%s\n", Green, blobURL, Reset),
									blobURL,
								}
							}
						}

						markerCode = resultsPage.NextMarker
					} else {
						if cs.verboseMode > 1 {
							fmt.Printf(
								"%s[-] Error while accessing %s: '%s'%s\n",
								Red, containerURLWithMarker, err, Reset,
							)
						}
						break
					}
				}

				page += 1
			}
		} else {
			if cs.verboseMode > 2 {
				fmt.Printf("%s[+][C=%d] %s%s\n", Red, checkStatusCode, containerURL, Reset)
			}
		}
	}
}

func (cs *ContainerScanner) runDirectScan(accounts[] string, containerNames []string) {
	nContainers := len(containerNames)
	doneChan := make(chan struct{})
 
	for idx, account := range accounts {
		if cs.verboseMode > 0 {
			fmt.Printf(
				"[~][%d] Searching %d containers in account '%s'\n",
				idx,
				nContainers,
				account,
			)
		}

		go func(idx int, account string) {
			for i := 0; i < len(containerNames); i++ {
				<-doneChan
			}

			fmt.Printf("[~][%d] Finished searching account '%s'\n", idx, account)
		}(idx, account)

		for _, containerName := range containerNames {
			cs.wg.Add(1)
			cs.semaphore <- struct{}{}

			go cs.ScanContainer(account, containerName, doneChan)
		}
	}
}

func (cs *ContainerScanner) runInverseScan(accounts []string, containerNames []string) {
	nAccounts := len(accounts)
	doneChan := make(chan struct{})

	for idx, containerName := range containerNames {
		if cs.verboseMode > 0 {
			fmt.Printf(
				"[~][%d] Searching %d accounts for containers named '%s' \n",
				idx,
				nAccounts,
				containerName,
			)
		}

		go func(idx int, containerName string) {
			for i := 0; i < len(accounts); i++ {
				<-doneChan
			}

			fmt.Printf("[~][%d] Finished searching containers named '%s'\n", idx, containerName)
		}(idx, containerName)

		for _, account := range accounts {
			cs.wg.Add(1)
			cs.semaphore <- struct{}{}

			go cs.ScanContainer(account, containerName, doneChan)
		}
	}
}

func (cs *ContainerScanner) ScanList(accounts []string, containerNames []string) {
	if !cs.invertSearch {
		cs.runDirectScan(accounts, containerNames)
	} else {
		cs.runInverseScan(accounts, containerNames)
	}
}

func (cs *ContainerScanner) ScanInput(accounts []string, containerNames []string) {
	if !cs.invertSearch {
		cs.runDirectScan(accounts, containerNames)
	} else {
		cs.runInverseScan(accounts, containerNames)
	}
}