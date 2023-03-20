package xml

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type Properties struct {
	LastModified    string `xml:"Last-Modified"`
	Etag            string `xml:"Etag"`
	ContentLength   int64  `xml:"Content-Length"`
	ContentType     string `xml:"Content-Type"`
	ContentEncoding string `xml:"Content-Encoding"`
	ContentLanguage string `xml:"Content-Language"`
	ContentMD5      string `xml:"Content-MD5"`
	CacheControl    string `xml:"Cache-Control"`
	BlobType        string `xml:"BlobType"`
	LeaseStatus     string `xml:"LeaseStatus"`
}

type Blob struct {
	Name       string     `xml:"Name"`
	Url        string     `xml:"Url"`
	Properties Properties `xml:"Properties"`
}

type Blobs struct {
	Blob []Blob `xml:"Blob"`
}

type EnumerationResults struct {
	ServiceEndpoint string `xml:"ServiceEndpoint,attr"`
	ContainerName string `xml:"ContainerName,attr"`
	Blobs         Blobs  `xml:"Blobs"`
	NextMarker    string `xml:"NextMarker"`
}

func (e *EnumerationResults) LoadXML(xmlData []byte) error {
	err := xml.Unmarshal(xmlData, e)
	return err
}

func (e *EnumerationResults) BlobURLs() []string {
	var urls []string
	var blobUrl string

	for _, blob := range e.Blobs.Blob {
		// The logic here is that if no URL can be identified,
		// then it will append an empty blob URL to the list
		// to let the user know that a blob was found
		// but no URL could be extracted
		if blob.Url != "" {
			blobUrl = blob.Url
		} else if blob.Name != "" {
			if strings.HasPrefix(e.ContainerName, "https://") {
				blobUrl = fmt.Sprintf("%s/%s", e.ContainerName, blob.Name)
			} else if strings.HasPrefix(e.ServiceEndpoint, "https://") {
				blobUrl = fmt.Sprintf("%s/%s/%s", e.ServiceEndpoint, e.ContainerName, blob.Name)
			} else {
				blobUrl = ""
			}
		} else {
			blobUrl = ""
		}

		urls = append(urls, blobUrl)
	}

	return urls
}

func (e *EnumerationResults) TotalContentLength() int64 {
	var contentLength int64 = 0

	for _, blob := range e.Blobs.Blob {
		contentLength += blob.Properties.ContentLength
	}

	return contentLength
}
