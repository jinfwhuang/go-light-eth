## Ethereuem Lightweight Clients

### Status
experimental

### Goals
This project seeks to create extremely light-weight, trustless entrypoints to query and interact with Ethereum. 
This project will allow users to run different flavors of Ethereum light clients.

We are starting with the consensus client. We will support all the well-defined networking options. These networking 
specifications are still in the process of being discussed and finalized. We will implement them as they become finalized.

We also aim to add support for execution client. It will share similar functionalities to [trin](https://github.com/ethereum/trin).

- Lightweight consensus client
  - grpc networking: <span style="color:green">working poc</span>, see [instruction](cmd/consensus/README.md)
  - json-rpc networking: <span style="color:grey">not started</span>
  - libp2p networking: <span style="color:grey">not started</span>
  - portal network: <span style="color:grey">not started</span>

- Lightweight execution client
  - [LES](https://github.com/ethereum/devp2p/blob/master/caps/les.md): We do not intend to implement this option 
  - portal network: <span style="color:grey">not started</span>
