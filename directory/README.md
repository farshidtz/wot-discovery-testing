# WoT Discovery Testing
Test suite for [W3C WoT Discovery](https://www.w3.org/TR/wot-discovery/) APIs.


The list of assertions is read from `report/template.csv`.
If this file is not available, it will be downloaded from [wot-discovery/testing/template.csv](https://github.com/w3c/wot-discovery/blob/main/testing/template.csv) (main branch) and stored locally.
To download the latest assertions, simply remove the local file so that it gets re-downloaded.
To use an assertion list other than the one from the main branch of wot-discovery, simply place the file at the same path.

The output testing reports are written to:
- `report/tdd-auto.csv` (test report)
- `report/tdd-manual.csv` (untested assertions)

The test results are printed to standard output.

## Run
Useful CLI Arguments: 
```
--server string
        URL of the directory service
-testJSONPath
        perform informative JSONPath testing
-testXPath
        perform informative XPath testing
-v
        verbose: print additional output
--run regexp
        Run only those tests and examples matching the regular expression.      
```

For example, `--run=TestCreateThing` can be set to run only the test function named `TestCreateThing`.

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
docker run --rm -v $(pwd)/report:/report wot-discovery-testing --server=http://directory:8081
```
where `$(pwd)/report` is the path to the directory on the host.
