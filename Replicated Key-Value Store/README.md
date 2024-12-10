# CSE138_Assignment3

# Acknowledgments
 - N/A

# Citations 
- Used for learning how to build a docker for golang: https://hub.docker.com/_/golang
- We looked at this tutorial to understand how to create a simple http server in Go and what imports to use as well: https://www.digitalocean.com/community/tutorials/how-to-make-an-http-server-in-go
- When creating our data types we looked at Go's types: https://www.geeksforgeeks.org/data-types-in-go/
- We looked at this website to understand how to implement a key-value store in Golang: https://dev.to/ernesto27/key-value-store-in-golang-52h1
- Where we found our HTTP status code for Golang: https://go.dev/src/net/http/status.go
- What we utilized in retrieving the value: https://pkg.go.dev/encoding/json
- When trying to parse the URL we utlized Golang's URL.Path: https://pkg.go.dev/net/url#URL
- Learning more about content-type headers in Golang: https://www.informit.com/articles/article.aspx?p=2861456&seqNum=5
- How we learned to get an enviroment variable in Golang: https://www.kelche.co/blog/go/golang-environment-variables/#:~:text=To%20get%20an%20environment%20variable,will%20return%20an%20empty%20string. 
- When we ran into having issues with request re-trying we found this Stack Overflow that discuessed a solution in Golang: https://stackoverflow.com/questions/23297520/how-can-i-make-the-go-http-client-not-follow-redirects-automatically
- We had issues converting strings to integers so we found this Stack Overflow: https://stackoverflow.com/questions/4278430/convert-string-to-integer-type-in-go
- Golang language learning: https://objectcomputing.com/resources/publications/sett/january-2019-way-to-go-part-2

# Team Contributions
- Sophie Hernandez: setting up the github repo, dockerfile, worked on PUT request implementation, implemented casual-meta data, and replication. 
- Kim Pham: modified view operations, added functions for broadcasting + syncing for replicas, added error handling
- Samiyah Shaikh: Worked on DELETE request implementation, implemented initial version of view operations and error-handling

# Mechanism Description
- Our system tracks casual dependencies by associating a version number with each key in the key-value store (kvs) and we store the version number in stringToIntMap variable that key's integer value of its "version." Each key’s version is incremented with PUT or DELETE requests, and this version number is returned as casual-metadata in responses to indicate the latest state. Also was inspired by Lecture 8. 
- Our system detects when a replica goes down by periodically sending health-check requests to each replica in the view. Specifically, the ReplicaDown function iterates through the current view of replicas (excluding the local replica) and sends an HTTP GET request to the ```/health``` endpoint of each replica. If a response is not received or the status code is not ```200 OK```, the system assumes the replica is down. Upon detecting an unresponsive replica, it removes the replica from the view and broadcasts the updated view to the remaining replicas. This process runs continuously in the background, with a delay of 5 seconds between checks, ensuring the system promptly identifies and adapts to changes in replica availability.
 
