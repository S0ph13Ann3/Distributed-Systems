## ✩ OVERVIEW
This project involves creating an HTTP web service that handles different HTTP verbs (GET and POST) and specific URI paths (/hello, /hello/<name>, and /test). The service is containerized using Docker/Podman and listens on port 8090.

## ✩ KEY FEATURES
- HTTP endpoints with specific behaviors:
  - /hello: Handles GET requests with JSON response
  - /hello/<name>: Handles POST requests with path parameters
  - /test: Handles both GET and POST requests with query parameters
- Package the service in a container using Docker/Podman
- Listens on port 8090
- Follow REST API principles for request/response handling
- Uses proper HTTP status codes (200, 400, 405)

## ✩ SETUP AND RUNNING
```
# Build the container
docker build -t proj1 .

# Run the container
docker run --rm -p 8090:8090 proj1
