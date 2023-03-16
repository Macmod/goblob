package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"net/http"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Macmod/goblob/core"
	"github.com/Macmod/goblob/utils"
)

func main() {
	const BANNER = `
                  888      888          888      
                  888      888          888      
                  888      888          888      
 .d88b.   .d88b.  88888b.  888  .d88b.  88888b.  
d88P"88b d88""88b 888 "88b 888 d88""88b 888 "88b 
888  888 888  888 888  888 888 888  888 888  888 
Y88b 888 Y88..88P 888 d88P 888 Y88..88P 888 d88P 
 "Y88888  "Y88P"  88888P"  888  "Y88P"  88888P"  
     888                                         
Y8b d88P                                         
 "Y88P"                                          
`
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run goblob.go <targetaccount>")
		os.Exit(1)
	}

	accountsFilename := flag.String(
		"accounts", "", "File with target Azure storage account names to check",
	)
	containersFilename := flag.String(
		"containers", "wordlists/goblob-folder-names.txt",
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
	verbose := flag.Int("verbose", 1, "Verbosity level (default=1")
	blobs := flag.Bool(
		"blobs", false,
		"Show each blob URL in the results instead of their container URLs",
	)
	invertSearch := flag.Bool(
		"invertsearch", false,
		"Enumerate accounts for each container instead of containers for each account",
	)
	maxpages := flag.Int(
		"maxpages", 20,
		"Maximum of container pages to traverse looking for blobs",
	)
	timeout := flag.Int(
		"timeout", 90,
		"Timeout for HTTP requests (seconds)",
	)
	maxIdleConns := flag.Int(
		"maxidleconns", 100,
		"Maximum of idle connections",
	)
	maxIdleConnsPerHost := flag.Int(
		"maxidleconnsperhost", 10,
		"Maximum of idle connections per host",
	)
	maxConnsPerHost := flag.Int(
		"maxconnsperhost", 0,
		"Maximum of connections per host",
	)
	skipSSL := flag.Bool("skipssl", false, "Skip SSL verification")

	flag.Parse()

	fmt.Println(BANNER)

	// Import input from files
	var accounts []string
	var filteredAccounts []string

	if *accountsFilename != "" {
		accounts = utils.ReadLines(*accountsFilename)
	} else {
		accounts = []string{os.Args[1]}	
	}

	var containers []string = utils.ReadLines(*containersFilename)
	var filteredContainers []string

	// Results report
	resultsMap := make(core.ResultsMap)

	if *verbose > 0 {
		sigChannel := make(chan os.Signal, 1)
		signal.Notify(sigChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

		go func() {
			sig := <-sigChannel
			fmt.Printf(
				"%s[-] Signal detected (%s). Printing partial results...%s\n",
				core.Red,
				sig,
				core.Reset,
			)

			resultsMap.PrintResults()
			os.Exit(1)
		}()

		defer resultsMap.PrintResults()
	}

	// Setting up output file writer
	var writer *bufio.Writer
	if *output != "" {
		output_file, _ := os.OpenFile(*output, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		defer output_file.Close()

		writer = bufio.NewWriter(output_file)
	} else {
		writer = nil
	}

	// Dedicated goroutine for writing results
	outputChannel := make(chan core.Message)
	go core.ReportResults(writer, outputChannel)

	// HTTP client parameters
	var transport = http.Transport{
		DisableKeepAlives:   false,
		DisableCompression:  true,
		MaxIdleConns:        *maxIdleConns,
		MaxIdleConnsPerHost: *maxIdleConnsPerHost,
		MaxConnsPerHost:     *maxConnsPerHost,
	}

	if *skipSSL {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	var httpClient = &http.Client{
		Timeout:   time.Second * time.Duration(*timeout),
		Transport: &transport,
	}

	// Synchronization stuff
	semaphore := make(chan struct{}, *maxGoroutines)
	var wg sync.WaitGroup

	// Container scanner object
	var containerScanner = core.ContainerScanner{
		httpClient,
		&wg,
		semaphore,
		outputChannel,
		resultsMap,
		*blobs,
		*verbose,
		*maxpages,
	}

	// Filtering out invalid values from accounts and containers
	removedAccounts := 0
	removedContainers := 0

	if *verbose > 0 {
		fmt.Printf("[~] Filtering out invalid accounts and containers\n")
	}

	for idx, account := range accounts {
		account = strings.Replace(strings.ToLower(account), ".blob.core.windows.net", "", -1)
		if utils.IsValidStorageAccountName(account) {
			filteredAccounts = append(filteredAccounts, account)
		} else {
			if *verbose > 1 {
				fmt.Printf("[~][%d] Skipping invalid storage account name '%s'\n", idx, account)
			}
		}
	}

	for idx, containerName := range containers {
		containerName = strings.ToLower(containerName)
		if utils.IsValidContainerName(containerName) {
			filteredContainers = append(filteredContainers, containerName)
		} else {
			if *verbose > 1 {
				fmt.Printf("[~][%d] Skipping invalid storage account name '%s'\n", idx, containerName)
			}
		}
	}

	if *verbose > 0 {
		fmt.Printf(
			"[~] Ignored %d invalid accounts and %d invalid containers from input\n",
			removedAccounts,
			removedContainers,
		)
	}

	// Main loop
	if !*invertSearch {
		for idx, account := range filteredAccounts {
			if *verbose > 0 {
				fmt.Printf("[~][%d] Searching blob containers in storage account '%s'\n", idx, account)
			}

			for _, containerName := range filteredContainers {
				wg.Add(1)
				semaphore <- struct{}{}

				go containerScanner.Scan(account, containerName)
			}
		}
	} else {
		for idx, containerName := range filteredContainers {
			if *verbose > 0 {
				fmt.Printf("[~][%d] Searching blob containers named '%s'\n", idx, containerName)
			}

			for _, account := range filteredAccounts {
				wg.Add(1)
				semaphore <- struct{}{}

				go containerScanner.Scan(account, containerName)
			}
		}
	}

	wg.Wait()

	close(outputChannel)
}
