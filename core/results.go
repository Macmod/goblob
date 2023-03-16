package core

import (
	"bufio"
	"fmt"
	"github.com/Macmod/goblob/utils"
	"sort"
)

type ResultsMap map[string]ContainerStats

type AccountResult struct {
	name  string
	stats ContainerStats
}

type ContainerStats struct {
	containerNames map[string]struct{}
	numFiles       int
	contentLength  int64
}

func (r ResultsMap) SaveContainerResults(
	account string,
	containerName string,
	numFiles int,
	contentLength int64,
) {
	if entity, ok := r[account]; ok {
		entity.containerNames[containerName] = utils.Empty
		entity.numFiles += numFiles
		entity.contentLength += contentLength

		r[account] = entity
	} else {
		r[account] = ContainerStats{
			map[string]struct{}{containerName: utils.Empty},
			numFiles,
			contentLength,
		}
	}
}

func (result ResultsMap) PrintResults() {
	fmt.Printf("[+] Results:\n")
	if len(result) != 0 {
		var numFiles int = 0
		var numContainers int = 0

		entries := make([]AccountResult, 0, len(result))
		for accountName, containerStats := range result {
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
			Green, numFiles, numContainers, len(result), Reset,
		)
	} else {
		fmt.Printf("%s[-] No files found.%s\n", Red, Reset)
	}
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
