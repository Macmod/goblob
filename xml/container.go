package container

import (
	"encoding/xml"
)

type Properties struct {
	LastModified   string `xml:"Last-Modified"`
	Etag           string `xml:"Etag"`
	ContentLength  int64    `xml:"Content-Length"`
	ContentType    string `xml:"Content-Type"`
	ContentEncoding string `xml:"Content-Encoding"`
	ContentLanguage string `xml:"Content-Language"`
	ContentMD5     string `xml:"Content-MD5"`
	CacheControl   string `xml:"Cache-Control"`
	BlobType       string `xml:"BlobType"`
	LeaseStatus    string `xml:"LeaseStatus"`
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
	ContainerName string `xml:"ContainerName,attr"`
	Blobs         Blobs  `xml:"Blobs"`
	NextMarker    string `xml:"NextMarker"`
}

func (e *EnumerationResults) LoadXML(xmlData []byte) {
	err := xml.Unmarshal(xmlData, e)
	if err != nil {
		panic(err)
	}
}

func (e *EnumerationResults) BlobURLs() []string {
	var urls []string
	
	for _, blob := range e.Blobs.Blob {
		urls = append(urls, blob.Url)
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