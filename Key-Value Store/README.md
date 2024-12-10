## ✩ OVERVIEW
This project involves implementing a distributed key-value store with forwarding capabilities. 

The project is divided into two main parts:
### Part 1: Single-site Key-Value Store
- Implementation of a basic key-value store with HTTP API support
- In-memory storage (no persistence required)
- Support for basic operations: PUT, GET, DELETE

### Part 2: Key-Value Store with Proxies
- Extension of Part 1 to support distributed operation
- Implementation of main and forwarding instances
- Network communication between instances
- Handling of instance failures

## ✩ KEY FEATURES
- PUT /kvs/<key>
  - Creates or updates key-value mappings
  - Handles different response codes (200, 201, 400)
  - Validates key length and request body format
  - Supports any JSON value type
- GET /kvs/<key>
  - Retrieves values for existing keys
  - Returns appropriate error for non-existent keys
  - Status codes: 200 for success, 404 for not found
- DELETE /kvs/<key>
  - Removes key-value mappings
  - Returns confirmation of deletion
  - Handles non-existent key errors

## ✩ DISTRIBUTED SYSTEM FEATURES
- Main instance for direct request handling
- Forwarding instances that proxy requests to main instance
- Error handling for instance failures (503 Service Unavailable)
- Docker network configuration for instance communication

## ✩ SETUP AND RUNNING
```
# Build the container
docker build -t proj2img .

# Create network
docker network create --subnet=10.10.0.0/16 proj2net

# Run main instance
docker run --rm -p 8082:8090 --net=proj2net --ip=10.10.0.2 --name main-instance proj2img

# Run forwarding instance
docker run --rm -p 8083:8090 --net=proj2net --ip=10.10.0.3 -e FORWARDING_ADDRESS=10.10.0.2:8090 --name forwarding-instance1 proj2img
