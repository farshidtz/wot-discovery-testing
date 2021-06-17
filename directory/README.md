# WoT Discovery Testing
Test suite for [W3C WoT Discovery](https://www.w3.org/TR/wot-discovery/) APIs.

The final report is printed to the standard output as well as to `./report.csv` file.

## Run
CLI Arguments: 
```
--report string
        Path to create report (default "report.csv")
--server string
        URL of the directory service
```

### Go
Linux/macOS: 
```bash
go test --server=http://localhost:8081 
```

### Docker
#### Build
```bash
docker build -t wot-discovery-testing .
```

#### Run
If the server is running locally, give to IP address of the Docker host instead of `localhost`. 

Alternatively, you can run the container in host mode, but that doesn't work with Docker Desktop on macOS:
```bash
docker run --rm --net=host wot-discovery-testing --server=http://localhost:8081 
```