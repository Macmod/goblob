package core

import (
	"fmt"
	"sync"
	"bytes"
	"io"
	"net/http"
	"github.com/Macmod/goblob/xml"
)

type Message struct {
	textToStdout string
	textToFile   string
}

type ContainerScanner struct {
	HttpClient *http.Client
	WG *sync.WaitGroup
	Semaphore chan struct{}
	OutputChannel chan Message
	ResultsMap ResultsMap
	BlobsOnly bool
	VerboseMode int
	MaxPages int
}

func (cs ContainerScanner) Scan(account string, containerName string) {
	defer func() {
		<-cs.Semaphore
		cs.WG.Done()
	}()

	containerURL := fmt.Sprintf(
		"https://%s.blob.core.windows.net/%s?restype=container",
		account,
		containerName,
	)

	checkResp, err := cs.HttpClient.Get(containerURL)
	if err != nil {
		if cs.VerboseMode > 1 {
			fmt.Printf("%s[-] Error while fetching URL: '%s'%s\n", Red, err, Reset)
		}
	} else {
		defer checkResp.Body.Close()

		checkStatusCode := checkResp.StatusCode
		if checkStatusCode < 400 {
			cs.ResultsMap.SaveContainerResults(
				account,
				containerName,
				0, 0,
			)

			if !cs.BlobsOnly {
				cs.OutputChannel <- Message{
					fmt.Sprintf("%s[+][C=%d] %s%s\n", Green, checkStatusCode, containerURL, Reset),
					containerURL,
				}
			}

			markerCode := "FirstPage"
			page := 1
			for markerCode != "" && (cs.MaxPages == -1 || page <= cs.MaxPages) {
				var containerURLWithMarker string
				if page == 1 {
					containerURLWithMarker = fmt.Sprintf("%s&comp=list&showonly=files", containerURL)
				} else {
					containerURLWithMarker = fmt.Sprintf("%s&comp=list&showonly=files&marker=%s", containerURL, markerCode)
				}

				resp, err := cs.HttpClient.Get(containerURLWithMarker)
				if err != nil {
					if cs.VerboseMode > 1 {
						fmt.Printf("%s[-] Error while fetching URL: '%s'%s\n", Red, err, Reset)
					}
				} else {
					statusCode := resp.StatusCode
					defer resp.Body.Close()

					resBuf := new(bytes.Buffer)
					_, err = io.Copy(resBuf, resp.Body)
					if err != nil {
						if cs.VerboseMode > 1 {
							fmt.Printf("%s[-] Error while reading response body: '%s'%s\n", Red, err, Reset)
						}
						break
					}

					if statusCode < 400 {
						resultsPage := new(xml.EnumerationResults)
						resultsPage.LoadXML(resBuf.Bytes())

						blobURLs := resultsPage.BlobURLs()
						cs.ResultsMap.SaveContainerResults(
							account,
							containerName,
							len(blobURLs),
							resultsPage.TotalContentLength(),
						)

						if cs.BlobsOnly {
							for _, blobURL := range blobURLs {
								cs.OutputChannel <- Message{
									fmt.Sprintf("%s[+] %s%s\n", Green, blobURL, Reset),
									blobURL,
								}
							}
						}

						markerCode = resultsPage.NextMarker
					} else {
						if cs.VerboseMode > 1 {
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
			if cs.VerboseMode > 2 {
				fmt.Printf("%s[+][C=%d] %s%s\n", Red, checkStatusCode, containerURL, Reset)
			}
		}
	}
}