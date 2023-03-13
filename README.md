# Goblob 🫐

Goblob is a lightweight and fast enumeration tool designed to aid in the discovery of sensitive information exposed publicy in Azure blobs, which can be useful for various research purposes such as vulnerability assessments, penetration testing, and reconnaissance.

# Installation
`go install github.com/Macmod/goblob@latest`

# Usage

To use goblob simply run the following command:

```bash
$ ./goblob <target>
```

Where `<target>` is the target storage account name to enumerate public Azure blob storage URLs on.

You can also specify a list of storage account names:
```bash
$ ./goblob -accounts accounts.txt
```

By default, the tool will use a list of common Azure Blob Storage container names to construct potential URLs. However, you can also specify a custom list of container names using the `-containers` option. For example:

```bash
$ ./goblob example.com -accounts accounts.txt -containers wordlists/goblob-folder-names.txt
```

The tool also supports outputting the results to a file using the `--output` option:
```bash
$ ./goblob example.com -output results.txt
```

## Optional Flags
- `--goroutines` - Maximum number of concurrent goroutines to allow (default: `5000`).
- `--blobs=true` - Report the URL of each blob instead of the URL of the containers (default: `false`).
- `--verbose=N` - Set verbosity level (default: `1`, min: `0`, max: `3`).

## Example

TODO: Put example here

# Contributing
Contributions are welcome by [opening an issue](https://github.com/Macmod/goblob/issues/new) or by [submitting a pull request](https://github.com/Macmod/goblob/pulls).

# TODO
* Improve project structure
* Check blob domain for NXDOMAIN before trying wordlist

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

