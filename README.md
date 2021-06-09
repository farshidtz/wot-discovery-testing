# WoT Discovery Testing
Test suite for [W3C WoT Discovery](https://www.w3.org/TR/wot-discovery/) APIs

## Run

### Go
Linux/macOS: 
```bash
URL=http://localhost:8081 go test -v
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
docker run --rm --net=host -e URL=http://localhost:8081 wot-discovery-testing
```