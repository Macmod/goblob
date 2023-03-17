# Goblob 🫐

![](https://img.shields.io/github/go-mod/go-version/Macmod/goblob) ![](https://img.shields.io/github/languages/code-size/Macmod/goblob) ![](https://img.shields.io/github/license/Macmod/goblob) ![](https://img.shields.io/github/actions/workflow/status/Macmod/goblob/release.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/Macmod/goblob)](https://goreportcard.com/report/github.com/Macmod/goblob)

Goblob is a lightweight and fast enumeration tool designed to aid in the discovery of sensitive information exposed publicy in Azure blobs, which can be useful for various research purposes such as vulnerability assessments, penetration testing, and reconnaissance.

*Warning*. Goblob will issue individual goroutines for each container name to check in each storage account, only limited by the maximum number of concurrent goroutines specified in the `-goroutines` flag. This implementation can exhaust bandwidth pretty quickly in most cases with the default wordlist, or potentially cost you a lot of money if you're using the tool in a cloud environment. Make sure you understand what you are doing before running the tool.

# Installation
`go install github.com/Macmod/goblob@latest`

# Usage

To use goblob simply run the following command:

```bash
$ ./goblob <storageaccountname>
```

Where `<storageaccountname>` is the target storage account to enumerate public Azure blob storage URLs on.

You can also specify a list of storage account names to check:
```bash
$ ./goblob -accounts accounts.txt
```

By default, the tool will use a list of common Azure Blob Storage container names to construct potential URLs. However, you can also specify a custom list of container names using the `-containers` option. For example:

```bash
$ ./goblob -accounts accounts.txt -containers wordlists/goblob-folder-names.txt
```

The tool also supports outputting the results to a file using the `-output` option:
```bash
$ ./goblob -accounts accounts.txt -containers wordlists/goblob-folder-names.txt -output results.txt
```

If you want to provide accounts to test via `stdin` you can also omit `-accounts` (or the account name) entirely:

```bash
$ cat accounts.txt | ./goblob
```

## Wordlists

Goblob comes bundled with two basic wordlists that can be used with the `-containers` option:

- [wordlists/goblob-folder-names.txt](wordlists/goblob-folder-names.txt) (default) - Adaptation from koaj's [aws-s3-bucket-wordlist](https://github.com/koaj/aws-s3-bucket-wordlist/blob/master/list.txt) - a wordlist containing generic folder names that are likely to be used as container names.
- [wordlists/goblob-folder-names.small.txt](wordlists/goblob-folder-names.small.txt) - Subset of the default wordlist containing only words that have been found as container names in a real experiment with over 35k distinct storage accounts + words from the default wordlist that are part of the NLTK corpus.

## Optional Flags

Goblob provides several flags that can be tuned in order to improve the enumeration process:

- `-goroutines=N` - Maximum number of concurrent goroutines to allow (default: `5000`).
- `-blobs=true` - Report the URL of each blob instead of the URL of the containers (default: `false`).
- `-verbose=N` - Set verbosity level (default: `1`, min: `0`, max: `3`).
- `-maxpages=N` - Maximum of container pages to traverse looking for blobs (default: `20`, set to `-1` to disable limit or to `0` to avoid listing blobs at all and just check if the container is public)
- `-timeout=N` - Timeout for HTTP requests (seconds, default: `90`)
- `-maxidleconns=N` - `MaxIdleConns` transport parameter for HTTP client (default: `100`)
- `-maxidleconnsperhost=N` - `MaxIdleConnsPerHost` transport parameter for HTTP client (default: `10`)
- `-maxconnsperhost=N` - `MaxConnsPerHost` transport parameter for HTTP client (default: `0`)
- `-skipssl=true` - Skip SSL verification (default: `false`)
- `-invertsearch=true` - Enumerate accounts for each container instead of containers for each account (default: `false`)

For instance, if you just want to find publicly exposed containers using large lists of storage accounts and container names, you should use `-maxpages=0` to prevent the goroutines from paginating the results. Then run it again on the set of results you found with `-blobs=true` and `-maxpages=-1` to actually get the URLs of the blobs.

If, on the other hand, you want to test a small list of very popular container names against a large set of storage accounts, you might want to try `-invertsearch=true` with `-maxpages=0`, in order to see the public accounts for each container name instead of the container names for each storage account.

You may also want to try changing `-goroutines`, `-timeout` and `-maxidleconns`, `-maxidleconnsperhost` and `-maxconnsperhost` and `-skipssl` in order to best use your bandwidth and find results faster.

Experiment with the flags to find what works best for you ;-)

## Example

[![asciicast](https://asciinema.org/a/568038.svg)](https://asciinema.org/a/568038)

# Contributing
Contributions are welcome by [opening an issue](https://github.com/Macmod/goblob/issues/new) or by [submitting a pull request](https://github.com/Macmod/goblob/pulls).

# TODO
* Check blob domain for NXDOMAIN before trying wordlist to save bandwidth (maybe)
* Improve default parameters for better performance
* Improve wordlists
* Guess blob URLs when `-blobs=true` returns the name but doesn't return the URL of the blobs

# License
The MIT License (MIT)

Copyright (c) 2023 Artur Henrique Marzano Gonzaga

Permission is hereby granted, free of charge, to any person
obtaining a copy of this software and associated documentation
files (the "Software"), to deal in the Software without
restriction, including without limitation the rights to use,
copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following
conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.

