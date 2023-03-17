package core

import (
	"bufio"
	"fmt"
	"github.com/Macmod/goblob/utils"
	"sort"
	"sync"
)

type ResultsMap struct {
	results map[string]ContainerStats
	mutex   sync.Mutex
}

type AccountResult struct {
	name  string
	stats ContainerStats
}

type ContainerStats struct {
	containerNames map[string]struct{}
	numFiles       int
	contentLength  int64
}

func (rm *ResultsMap) Init() {
	rm.results = make(map[string]ContainerStats)
}

func (rm *ResultsMap) StoreContainerResults(
	account string,
	containerName string,
	numFiles int,
	contentLength int64,
) {
	rm.mutex.Lock()

	if entity, ok := rm.results[account]; ok {
		entity.containerNames[containerName] = utils.Empty
		entity.numFiles += numFiles
		entity.contentLength += contentLength

		rm.results[account] = entity
	} else {
		rm.results[account] = ContainerStats{
			map[string]struct{}{containerName: utils.Empty},
			numFiles,
			contentLength,
		}
	}

	rm.mutex.Unlock()
}

func (rm *ResultsMap) PrintResults() {
	rm.mutex.Lock()

	fmt.Printf("[+] Results:\n")
	if len(rm.results) != 0 {
		var numFiles int = 0
		var numContainers int = 0

		entries := make([]AccountResult, 0, len(rm.results))
		for accountName, containerStats := range rm.results {
			entries = append(entries, AccountResult{accountName, containerStats})
		}

		sort.Slice(entries, func(i, j int) bool {
			return entries[i].stats.numFiles > entries[j].stats.numFiles
		})

		for _, entry := range entries {
			fmt.Printf(
				"%s[+] %s - %d files in %d containers (%s)%s\n",
				Green, entry.name,
				entry.stats.numFiles,
				len(entry.stats.containerNames),
				utils.FormatSize(int64(entry.stats.contentLength)),
				Reset,
			)

			numContainers += len(entry.stats.containerNames)
			numFiles += entry.stats.numFiles
		}

		fmt.Printf(
			"%s[+] Found a total of %d files across %d account(s) and %d containers%s\n",
			Green, numFiles, numContainers, len(rm.results), Reset,
		)
	} else {
		fmt.Printf("%s[-] No files found.%s\n", Red, Reset)
	}

	rm.mutex.Unlock()
}

func ReportResults(writer *bufio.Writer, msgChannel chan Message) {
	for msg := range msgChannel {
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
}
