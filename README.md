# celestia-peer-checker
Checks the ASNs of your nodes peers to see if there is a centralization risk. Having your peers centralized on certain cloud providers can cause a risk since if the provider goes down, you lose most of your peers and might lose connection to the network. This also shows the centralization around networks and providers like Contabo.
![alt text](https://github.com/fkbenjamin/celestia-peer-checker/blob/main/preview.png?raw=true)

## Setup
### Required
You need to have a recent version of Go installed
### Run
```
go run main.go
```
### Build
```
go build main.go
```
