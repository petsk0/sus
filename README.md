sus - Simple URL Shortener
==========================

### Prerequisites

Working linux distribution, go 1.18 compiler and internet connection

### Installation and usage

##### Installation

From the command line
```
mkdir shortener && cd shortener
git clone https://github.com/petsk0/sus .
go build ./cmd/sus/main.go
./main
```

##### Configuration

Configuration of the server is done through command line flags.
For more information see:
```
./main -h
```

##### Usage

Assuming all default flags, after running the binary open your web browser and go to localhost:8080.

Insert the URL you wish to shorten into the top text field and press enter (or click the button). Then retrieve your shortened URL from the bottom text field.

##### Caveats

sus will shorten relative URL paths without prefixing them with schemes or domain names!
