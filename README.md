# File Sending Service
A file sending service written in [Golang](https://golang.org/)

## Prerequisites
Please make sure that you have [Golang](https://golang.org/doc/) installed on your machine. The latest Golang can be downloaded at: https://golang.org/doc/install

## How to use the program
1. Open project to root directory in three separate terminals.
2. Starting relay program:
```
    cd relay
    go run relay.go :<port>
```
    The relay server is now ready for transferring files.
3. Starting sender program:
```
    cd sender
    go run sender.go <relay-url> <file-name>
```
    This will output a secret code that will need to be used for receiver program
4. Starting receiver program:
```
    cd receiver
    go run receiver.go <relay-url> <secret-code> <output-directory>
```
    This will start the transaction process

## Architecture
### TCP

### Context
    The service needs to stream data through a server without saving data onto the server
### Decision
    I choose to use TCP layer instead of the UDP protocol because TCP is more reliable - making sure that data sent with TCP is not lost or corrupted in transit.
### Consequences
    However, because of the reliability of TCP, it tend to be slower than UDP. If the server is going to handle an significantly large amount of users transferring data at the same time, UDP would be faster, but at the same time, not as secure.

### Golang

#### Context
    The service needs to be able to handle multiple user transferring data at the same time

#### Decision
    I choose to use Golang for accomplishing this task because the easy implementation on goroutine for concurrency.
#### Consequences
    I have never write Golang except doing some introduction tutorials. It's a static typing language which is very different from JavaScript, the language I use for work. I spent quite some time before starting the project on learning the basics of Golang. I also spent more time on learning conventions for writing a Go project. If given more time, I would do further research and implement Golang best practices to improve performance and code quality

### Using a map to store token as the key and connections as the value

#### Context
    The service needs to transfer data through two different accepted connections

#### Decision
    I created a global map to store the connections I want to write to. If a coming request token exists in the map, the relay will read from the connection and take those chucks and write them to the request connection. Otherwise, it will store the connection as the value and the token as the key into the map. I chose this approach because it uses data as the boundary to differentiate between sender request and receiver request without having to implement a UUID or signal for relay to know which request is sender and which request is receiver

#### Consequences
    The flaw in this approach so far I have found is that the map will takes up a lot of memory when there are a lot of accepted connections at the same time going though the server. I have addressed it with setting a deadline on the connection and also remove closed connection from the map each time when a new connection established. However, that will slow down the transaction speed with all these process running.


## Future Improvement
- Ensure 100% code coverage of all methods
- Further modularize the code(Create a utility package to decrease code replication)
- Add more formal documentation with a Golang docs tool
- Add a locale file to allow localization
- Better error handling system (Use a logging tool for better error logging)
- Learn more about channel to have a better control on the goroutine
