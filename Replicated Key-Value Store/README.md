## ✩ OVERVIEW
This project implements a replicated, fault-tolerant, and causally consistent key-value store. The system runs as a collection of communicating instances (replicas) where key-value pairs are replicated across all instances to ensure data availability and consistency.

## ✩ KEY FEATURES
- Replication and Fault Tolerance
  - Data is replicated across multiple instances
  - System remains available if one replica crashes
  - No data persistence required (in-memory storage)
  - Automatic replica failure detection
- Causal Consistency
  - Maintains causal ordering of events
  - Tracks causal dependencies using metadata
  - Ensures "read your writes" consistency
  - Handles concurrent operations
- API Endpoints
  - View Operations (/view)
    - PUT: Add new replica to the view
    - GET: Retrieve current view of replicas
    - DELETE: Remove replica from view
- Key-Value Operations (/kvs)
  - PUT: Create/update key-value pairs
  - GET: Retrieve values
  - DELETE: Remove key-value pairs
  - All operations include causal metadata

## ✩ ARCHITECTURE
- Distributed system with multiple replicas
- Each replica maintains its own copy of the data
- Replicas communicate state changes among themselves
- Clients can interact with any replica
- Causal consistency maintained across all operations

## ✩ MECHANISM DESCRIPTION 
- The system tracks casual dependencies by associating a version number with each key in the key-value store (kvs) and we store the version number in stringToIntMap variable that key's integer value of its "version." Each key’s version is incremented with PUT or DELETE requests, and this version number is returned as casual-metadata in responses to indicate the latest state.
- The system detects when a replica goes down by periodically sending health-check requests to each replica in the view. Specifically, the ReplicaDown function iterates through the current view of replicas (excluding the local replica) and sends an HTTP GET request to the ```/health``` endpoint of each replica. If a response is not received or the status code is not ```200 OK```, the system assumes the replica is down. Upon detecting an unresponsive replica, it removes the replica from the view and broadcasts the updated view to the remaining replicas. This process runs continuously in the background, with a delay of 5 seconds between checks, ensuring the system promptly identifies and adapts to changes in replica availability.
 
## ✩ SETUP AND RUNNING
```
# Build the container
docker build -t proj3img .

# Create network
docker network create --subnet=10.10.0.0/16 proj3net

# Run replicas
docker run --rm -p 8082:8090 --net=proj3net --ip=10.10.0.2 --name=alice \
-e=SOCKET_ADDRESS=10.10.0.2:8090 \
-e=VIEW=10.10.0.2:8090,10.10.0.3:8090,10.10.0.4:8090 proj3img

docker run --rm -p 8083:8090 --net=proj3net --ip=10.10.0.3 --name=bob \
-e=SOCKET_ADDRESS=10.10.0.3:8090 \
-e=VIEW=10.10.0.2:8090,10.10.0.3:8090,10.10.0.4:8090 proj3img

docker run --rm -p 8084:8090 --net=proj3net --ip=10.10.0.4 --name=carol \
-e=SOCKET_ADDRESS=10.10.0.4:8090 \
-e=VIEW=10.10.0.2:8090,10.10.0.3:8090,10.10.0.4:8090 proj3img
