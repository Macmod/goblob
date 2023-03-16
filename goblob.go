package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
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

	accountsFilename := flag.String(
		"accounts", "", "File with target Azure storage account names to check",
	)
	containersFilename := flag.String(
		"containers", "wordlists/goblob-folder-names.txt",
		"Wordlist file with possible container names for Azure blob storage",
	)
	maxGoroutines := flag.Int(
		"goroutines", 500,
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

	if *verbose > 0 {
		fmt.Println(BANNER)
	}

	// Import input from files
	var accounts []string
	if *accountsFilename != "" {
		accounts = utils.ReadLines(*accountsFilename)
	} else if flag.NArg() > 0 {
		accounts = []string{flag.Arg(0)}
	}

	var containers []string = utils.ReadLines(*containersFilename)

	// Results report
	resultsMap := new(core.ResultsMap)
	resultsMap.Init()

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

	// Container scanner object
	var containerScanner = new(core.ContainerScanner)
	containerScanner.Init(
		httpClient,
		*maxGoroutines,
		outputChannel,
		resultsMap,
		*blobs,
		*verbose,
		*maxpages,
		*invertSearch,
	)
	defer containerScanner.Done()

	var filteredContainers []string
	utils.FilterValidContainers(containers, &filteredContainers, *verbose > 1)

	if flag.NArg() == 0 && *accountsFilename == "" {
		if !*invertSearch {
			scanner := bufio.NewScanner(os.Stdin)

			for scanner.Scan() {
				account := scanner.Text()
				accounts = []string{account}
				if utils.IsValidStorageAccountName(account) {
					containerScanner.ScanList(accounts, filteredContainers)
				}
			}
		} else {
			fmt.Printf(
				"%s[-] The 'invertsearch' flag cannot be used with an input accounts file.%s\n",
				core.Red,
				core.Reset,
			)
		}
	} else {
		var filteredAccounts []string
		utils.FilterValidAccounts(accounts, &filteredAccounts, *verbose > 1)

		containerScanner.ScanList(filteredAccounts, filteredContainers)
	}
}
