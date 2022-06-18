# Web Crawler

This is a simple webcrawler that crawls a given URL as well as visiting the links on the page. This crawler avoids
visiting the same URL twice, and any external links are ignored.

## Requirements

These dependencies can also be installed with [Homebrew](https://brew.sh/).

* Requires Go 1.18 or greater. This can be installed with brew `brew install go` or
  downloaded [here](https://golang.org/doc/install).
* Requires Golangci Lint. This can be installed with brew `brew install golangci-lint` or
  downloaded [here](https://golangci-lint.run/usage/install/#local-installation).

### Install Dependencies

Install dependencies, issue the following command(s):

```bash
make install
```

### Testing and Formatting

To run the tests, issue the following command(s):

```bash
make test
```

#### Lint only

Run linting only:

```bash
make lint
```

## How to Run

To run the application with the default settings, simply issue the following example command(s):

```bash
make run https://example.com/
```

If you wish to change the default settings, navigate to `cmd/main.go` and change the default settings. Here is the list
of settings that can be changed:

- `retryMax`: The maximum number of times to retry a failed request.
- `retryMaxWait`: The maximum amount of time to wait before retrying a failed request.
- `workers`: The number of concurrent workers to use.
