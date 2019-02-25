# mock-taps-server
A simple server meant to mimic the behavior bts.ucsc.edu:8081/location/get

## Installation

```shell
$ go get -u github.com/slugbus/mock-taps-server
```

## Usage

```
$ mock-taps-server [flags]
```
```
flags:

-data string
    	file to use as mock data (default "data/mock-data-from-feb-24.json")

-interval duration
    	the interval that the mock data is spaced apart (default 3s)

 -port uint
    	port to listen on (default 8080)
```

## Examples

```
$ mock-taps-server
```

Output:
```
2019/02/24 23:59:08 Using file data/mock-data-from-feb-24.json as mock data
2019/02/24 23:59:08 Data points are updated ~every 3s
2019/02/24 23:59:08 Starting server on 0.0.0.0:8080
2019/02/24 23:59:08 Send queries to http://0.0.0.0:8080/location/get
```
