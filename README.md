# IP over UDP
## Design Documentation
```
+--------------------------+
|                          |
|    Application Layer     |   +-------------------+
|                          |   |                   |
+--------------------------+   |  Rip Layer        |
                               |                   |
                               +-------------------+
+--------------------------+
|                          |
|    Network Layer         |
|                          |
+--------------------------+

+--------------------------+
|                          |
|     Link Layer           |
|                          |
+--------------------------+
```


### Link Layer Abstraction
All Link Layer functionalities are encapsulated within a `LinkAccessor` structure and this layer does not expose that a UDP socket is used the transport data. The Primary functionality of this layer is to send and receive data to and from other nodes in the network.

For sending, it takes a structure, containing the data to be sent along with a next hop VirtualIP address, as input argument. It uses the next hop VirtualIP address to determin the physical UDP address and pass the data out.

For receiving, it reads UDP datagrams into 64Kib buffers, and simply pass the data back to the upper layer (via a channel) in a packet structure format.

### Network Layer (IP)
All Network Layer functionalities are encapsulated within a `NetworkAccessor` structure and this layer primarily does routing and forwarding. It holds a `RoutingTable` structure to use as the guide to decide where route IP packets should go next. It is not aware of the different interfaces in the link layer and it specifies where to send data by naming the next hop VirtualIP address.

The Network layer takes payload and destination as inputs when sending out or forwarding data. It looks up the routing table with the desired destination and decide which next hop address to forward to, construct (in the case of initiating a new message) or update (in the case of forwarding a packet) the packet before placing the request in a channel. The request will eventually be picked up by the Link Layer (detailed later) and gets sent out.

Upon receiving a packet (via a channel, which is constantly populated by the link layer), the Network layer does a series of checks before passing the received packet to a handler. It first has to check whether there are violations regarding the packet such as the TTL and checksum on the header. It then has to decide whether to forward to packet or not by comparing the packet's destination to all the local VirtualIPs. If it does decide to forward, it update the TTL on the header, recalculate the checksum and pass the packet along to the next hop. If the packet has one of the local VirtualIP as its destination, the Network layer will invoke the corresponding handler to handle the packet. The invocation is done on a different thread so the Network can continue to handle new packets while the handler does its work.

The network layer expose a method to register handlers to be associated with a protocol number. When a packet arrives at its final destionation, a handler is choosen based on the protcol number in the IP header and the packet is then handled by that specific handler.

### RIP
When a IP packet with protocol 200 arrives at its destination, the Network layer will invoke the RIP handler (previous registered at program start) to handle the RIP message. The RIP handler, who also has a reference to the routing table that is used in the network layer, will update the routing table according to the RFC specification and the assignment requirements. The routing table is thread-safe with a internal Read-Write mutex (detailed later), so that concurrency issues can be avoided.

Split-horizon with poisoned-reverse is implemented by remembering from which node a route is learned (by looking at the IP header in RIP messages). When broadcasting, split-horizon with poison reverse is applied where appropriate.

All routing entries are timestamped. When new routes comes in, existing entries will have their timestamps extended. If an entry is not renewed after 12 seconds, it will be marked as expired by a sweeper (detailed below in section Runners). During this expiration period, the route is still eligible to broadcasted but with a cost of 16 (according to the RFC spec). After another 6 seconds, it will be garbage collected and deleted from the table.

### Runners
There are 4 different runners in the system. Each one simply represents a continously-run thread that will be started by the main driver after program starts. Each of them have very simple logic and mostly does coordinations between different layers of the system and move data around different channels. This design decouples the Network layer and Link layer, eliminates strictly sequencial calls so a higher throughput of both send and receive can be achieved via the non-blocking nature of channels.

##### Link Receive Runner
This runner will continuously call the `#Receive()` method at the Link layer level to receive data off the UDP socket, it will then write the received data into a channel. Another runner for the Network layer will read data off this channel and asks the Network layer to process them.

##### Link Send Runner
This runner listens to a channal in a loop to send data out to other nodes. The channel this runner listens to is populated by the network layer.

##### Network Runner
This runner uses the `Select` call to manage 2 channels. One channel contains the data received from the Link layer and the other channel contains data to be sent out (either initiated by the `send` command or by the RIP handler). It continuously looks at the 2 channels, takes available data off a channel and calls the appropritate method on the Network layer to have the data processed.

##### RIP Runner
The RIP runner initiated the RIP logic periodically. Every 5 seconds, it invoke the RIP handler to broadcast all routes out to neighbors. Every 0.5 seconds, it invoke the RIP handler to broadcast updated routes to neighbors. The 0.5 second setting is configurable and arbitrarily chose since the assignment did not specify what the number needs to be and the RFC specification used a random number between 1~5. This setting helps reduce the number of Triggered Update broadcast within the network as it buffers 0.5 second worth of updates in a single broadcast.

The runner also continously scan the routing table for entries that should expire. It marks all expired routes as garbage-collectable. Another 6 seconds after the expiration mark, the entry will be sweeped and deleted from the routing table.

### Channels
There are 3 channels used in the system.
* Message channel
    - This channel is used to pass data to the Network layer. Whether the data a message to be sent out by the driver, or a RIP broadcast message, the data should first be handled by the Network layer to decided where the data should go. Most of the time the data will be passed down to the Link layer via another channl, but sometimes the data is handled locally (sending a message to a local VirutalIP). This channel mostly represent data initiated by the local node and going outward, but before getting processed by the Network layer.
* Link layer receive channel
    - This channel is written by the link layer and read by the network layer. The link layer packs the data received and place them into this channel. This channel represents the data receivd from other nodes in the network and coming inward.
* Link layer send channel
    -   This channel is written by the network layer and read by the link layer. The Network layer packs the payload into correct IP format and down pass them down. This channel represents the data going out to other nodes in the network after they are processed by the Network layer.

### Threading & Synchornization
At program start, a main thread starts to read input file and initiate components in the system. It spawns other threads to starts the Network and Link layer before going into a loop to receive command line inputs.

A thread is spawned for each of the runner mentioned above, therefore there are a total of 5 long-running thread in the system. A short-lived thread is spawned when a handler process a received IP packet (whether it simply prints out the packet or executes the RIP logic).

Synchornization is built into the `RoutingTable` since it is accessed by multiple thread. Read-write mutex is used and all methods into the routing table is guarded internally by the mutex. Therefore the synchornization is transparent outside of the data structure and the entire routing table structure is thread-safe.

The Interface table has a similar nature and we will be porting a similar locking mechanism fromt the routing table into the interface table.

### Next Steps
- Interface Table is currently only written when an interface is toggled between enabled status and disable status. It has not caused us any synchornization problems yet but we will be adding mutex mechanism into it.
- The routing table currently returns a pointer to the routing entry on read. We will motify this to return a deep copy of the entry so that no called can mutate the entry to cause synchornization issues. Any update required to made to the entry will be encapsulated by new methods ont the routing table itself, which is thread-safe.

