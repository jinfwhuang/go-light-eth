## Consensus Light Client

```bash
# Start light client
go run -v ./cmd/light-client/ \
--networking-mode grpc \
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

## Networking

#### grpc
The server needs to expose these [grpc services](https://github.com/jinfwhuang/prysm/blob/light-update/proto/prysm/v1alpha1/light_client.proto). 
```bash
# Check endpoint
grpcurl -plaintext host:4000 list ethereum.eth.v1alpha1.LightClient
```

#### json-rpc
Not defined, not implemented

#### libp2p
Not implemented. 

In the process of defined. See related PRs:
- https://github.com/ethereum/consensus-specs/pull/2267
- https://github.com/ethereum/consensus-specs/pull/2786
- https://github.com/ethereum/consensus-specs/pull/2802

#### portal network
Not implemented.

See:
- https://github.com/ethereum/portal-network-specs/tree/master/beacon-chain

