# Goblob 🫐

Goblob is a lightweight and fast enumeration tool designed to aid in the discovery of sensitive information exposed publicy in Azure blobs, which can be useful for various research purposes such as vulnerability assessments, penetration testing, and reconnaissance.

*Warning*. Goblob will issue individual goroutines for each container name to check in each storage account, only limited by the maximum number of concurrent goroutines specified in the `-goroutines` flag. This implementation can exhaust bandwidth & memory pretty quickly in most cases with the default wordlist, or potentially cost you a lot of money if you're using the tool in a cloud environment. Make sure you understand what you are doing before running the tool.

# Installation
`go install github.com/Macmod/goblob@latest`

# Usage

To use goblob simply run the following command:

```bash
$ ./goblob <storageaccountname>
```

Where `<target>` is the target storage account name to enumerate public Azure blob storage URLs on.

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

## Wordlists

Goblob comes bundled with two basic wordlists that can be used with the `-containers` option:

- [wordlists/goblob-folder-names.txt](wordlists/goblob-folder-names.txt) (default) - Adaptation from xajkep's [directory_only_one.small.txt](https://github.com/xajkep/wordlists/blob/master/discovery/directory_only_one.small.txt) - a wordlist containing generic folder names that are likely to be used as container names.
- [wordlists/goblob-folder-names.small.txt](wordlists/goblob-folder-names.small.txt) - Subset of the default wordlist containing only words that have been found as container names in a real experiment with over 35k distinct storage accounts + words from the default wordlist that are part of the NLTK corpus.

## Optional Flags
- `-goroutines=N` - Maximum number of concurrent goroutines to allow (default: `5000`).
- `-blobs=true` - Report the URL of each blob instead of the URL of the containers (default: `false`).
- `-verbose=N` - Set verbosity level (default: `1`, min: `0`, max: `3`).
- `-maxpages=N` - Maximum of container pages to traverse looking for blobs (default: `20`, set to `-1` to disable limit)
- `-timeout=N` - Timeout for HTTP requests (seconds, default: `20`)
- `-maxidleconns=N` - `MaxIdleConns` transport parameter for HTTP client
- `-maxidleconnsperhost=N` - `MaxIdleConnsPerHost` transport parameter for HTTP client
- `-maxconnsperhost=N` - `MaxConnsPerHost` transport parameter for HTTP client

## Example

TODO: Put example here

# Contributing
Contributions are welcome by [opening an issue](https://github.com/Macmod/goblob/issues/new) or by [submitting a pull request](https://github.com/Macmod/goblob/pulls).

# TODO
* Improve project structure
* Check blob domain for NXDOMAIN before trying wordlist to save bandwidth
* Option to read accounts from stdin

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

