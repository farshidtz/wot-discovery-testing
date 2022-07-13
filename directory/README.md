# WoT Discovery Testing
Test suite for [W3C WoT Discovery](https://www.w3.org/TR/wot-discovery/) APIs.


The list of assertions is read from `report/template.csv` and `report/manual.csv`.
If these files are not available, they will be downloaded from [wot-discovery/testing](https://github.com/w3c/wot-discovery/blob/main/testing) (main branch) and stored locally.
To download the latest assertions, simply remove the local files so that they gets re-downloaded.
To use assertion lists other than the one from the main branch of wot-discovery, replace the default URLs using CLI flags.

The output testing report is written to `report/tdd-auto.csv`.

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
--manualURL string
        URL to download template for assertions that are tested manually (default "https://raw.githubusercontent.com/w3c/wot-discovery/main/testing/manual.csv")
--templateURL string
        URL to download assertions template (default "https://raw.githubusercontent.com/w3c/wot-discovery/main/testing/template.csv")    
-v
        verbose: print additional output
--run regexp
        Run only those tests and examples matching the regular expression.  
```

To get all CLI flags, run: `go test --usage`.

For example, `--run=TestCreateThing` can be set to run only the test function named `TestCreateThing`.

The following commands should be executed from the current (i.e. `directory`) directory.

### Run natively
```bash
go test --server=http://localhost:8081 
```

### Run in a Docker container
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
