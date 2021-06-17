# WoT Discovery Testing
Test suite for [W3C WoT Discovery](https://www.w3.org/TR/wot-discovery/) APIs.

The test results are printed to standard output. The test report is written to a file in csv format.

## Run
CLI Arguments: 
```
--report string
        Path to create report (default "report.csv")
--server string
        URL of the directory service
-v
        verbose: print additional output
```

The following commands should be executed from the current (i.e. `directory`) directory.

### Go
```bash
go test --server=http://localhost:8081 
```

### Docker
#### Build
```bash
docker build -t wot-discovery-testing .
```

#### Run
If the server is running locally, pass IP address of the Docker host instead of `localhost`. E.g.:
```
docker run --rm wot-discovery-testing --server=http://172.17.0.1:8081
```

Docker Desktop for Mac and Windows add a special DNS name (`host.docker.internal`) which resolves to host:
```
docker run --rm wot-discovery-testing --server=http://host.docker.internal:8081
```

Alternatively, you can run the container in host mode. This may not work with Docker Desktop:
```bash
docker run --rm --net=host wot-discovery-testing --server=http://localhost:8081 
```

##### Report
To get the report, mount a volume on where the report is being generated. E.g.:
```
docker run --rm -v $(pwd)/report:/report wot-discovery-testing --server=http://directory:8081 --report=/report/report.csv
```
where `$(pwd)/report` is the path to the directory on the host.