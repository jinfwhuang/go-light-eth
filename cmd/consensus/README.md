## Light-client Server

This is implemented as the "light-update" service and server in the beacon node.
- service: beacon-chain/light-update/service.go
- api server: beacon-chain/rpc/prysm/v1alpha1/light-update

```bash
# Start a server that supports the APIs needed by a light client
go run ./cmd/beacon-chain --datadir=../prysm-data/mainnet --http-web3provider=https://mainnet.infura.io/v3/ecfa17010caa47a0afd0a543c43bbcc3

# Check the necessary APIs are available 
grpcurl -plaintext localhost:4000 list ethereum.eth.v1alpha1.LightClient
```

## Light-client Client

This is a new, independent program. All the codes are: cmd/light-client
- exposed api: cmd/light-client/rpc/server.go 

```bash
# Start light client
go run -v ./cmd/light-client/ \
--full-node-server-endpoint=127.0.0.1:4000 \
--grpc-port=4001 \
--data-dir=../prysm-data/lightnode \
--sync-mode=latest \
--trusted-current-committee-root='xxxx'

# --trusted-current-committee-root='UeSv92gwGs+DSk34NqOaCM1DaU9zyclQE6Tc9morK0M='  // roughly 2021-12-02
# --trusted-current-committee-root='rcWo3eE6KOLBLDQeahrXkdzxjWnE8qYHmL8HyNWv7b8='  // roughly 2021-12-03

# Check the server availability
grpcurl -plaintext localhost:4000 list

# Ge the current head
grpcurl -plaintext localhost:4001 ethereum.eth.v1alpha1.LightNode.Head
```



## Big TODOs
- Move the core light-client code outside of `cmd/.` 
- etc.