# db-xl-dump

Command to connect to an Oracle Database and export a table, view or sql query to a Microsoft Excel spreadsheet

## Prerequisites

Oracle Instant Client must already be installed

[Oracle Instant Client](https://www.oracle.com/database/technologies/instant-client.html)

Note - Oracle Instant Client must be configured per your environment (please follow the instructions provided by Oracle).

## Table of Contents

- [db-xl-dump](#db-xl-dump)
  - [Prerequisites](#Prerequisites)
  - [Table of Contents](#Table-of-Contents)
  - [Installation](#Installation)
  - [Building](#Building)
  - [Usage](#Usage)
  - [Support](#Support)
  - [Contributing](#Contributing)

## Installation

1) Clone this repository into a local directory, copy the db-xl-dump executable into your $PATH

```bash
$ git clone https://github.com/apexevangelists/db-xl-dump
```

## Building

Pre-requisite - install Go

Compile the program -

```bash
$ go build
```

## Usage

```bash-3.2$ ./db-xl-dump -h
Usage of ./db-xl-dump:
  -configFile string
    	Configuration file for general parameters (default "config")
  -connection string
    	Configuration file for connection
  -db string
    	Database Connection, e.g. user/password@host:port/sid
  -debug
    	Debug mode (default=false)
  -e string
    	Table(s), View(s) or queries to export
  -export string
    	Table(s), View(s) or queries to export
  -headers
    	Output Headers (default true)
  -o string
    	Output Filename (default "output.xlsx")
  -output string
    	Output Filename (default "output.xlsx")

bash-3.2$
```

## Support

Please [open an issue](https://github.com/apexevangelists/db-xl-dump/issues/new) for support.

## Contributing

Please contribute using [Github Flow](https://guides.github.com/introduction/flow/). Create a branch, add commits, and [open a pull request](https://github.com/apexevangelists/db-xl-dump/compare).