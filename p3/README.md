##Problem 3 Description##
Your task is to implement node logic that allows an arbitrary set of active nodes to agree on a "leader" node. If the leader fails, the remaining nodes should elect a new leader. Once elected, the leader must determine the active nodes in the system and advertise this set to all the nodes in the system (through the key-value service). The set of active nodes may change (as nodes may fail or join the system) and the leader must re-advertise the node set to reflect these events. Active nodes should periodically retrieve this list of active nodes and print it out.

Individual keys in the key-value service may experience permanent unavailability. Your node implementation must be robust to such unavailability and continue to elect leaders that will properly advertise the set of active nodes.

##Problem 3 Write Up##

###Assigning Key-Value Pairs
When a node initializes, we iterate through numbers starting at zero (first key = 0, second
key = 1, etc.) until we find a free key; a free key is one that doesn’t have a corresponding
value and is available. The node’s ID then becomes the free key’s value, and the key is no
longer free. Keys that have been marked as unavailable do not get re-assigned a value. The
same process is used for restarted nodes and to reassign a node to a new key should key
unavailability occur.

###Leader Election Algorithm
The leader node is whichever node is in the first chronological key-value pair. For example, if
k0 had value node1, node1 would be the leader. Nodes assume this to be true. If a leader
node’s corresponding key becomes unavailable and the node is assigned to a new key, it will
no longer be the leader as the new assigned key will not be the first key-value pair (by 1
above). If a leader node fails, the same election conditions apply. To put it succinctly, the
Leader Election algorithm is a process of self elimination. Each node checks whether it’s the
first key-value pair and assigns, or eliminates, itself as leader.

###Node Tracking/Advertisement
To get a list of nodes in the system, because we assigned node IDs to the value of a key, we
iterated through all available keys to get the corresponding node IDs. If a key is unavailable
or dead, we ignore it.
To deal with the case of node failure, we implemented a dead/alive protocol. To check if a
node is alive, since if a node fails it cannot set itself as dead, we added a PingBit to the end of
each node’s ID. This bit flips between 1 and 0 every time the node pings the key-value
service. We ping the service whenever we get an updated list of available node IDs. When
getting IDs, the node checks the current bit of another node against its last bit recorded with
the current node. If the bit is repeated, that indicates a dead node, as alive nodes will have
alternated their bit. The dead node’s key is then flagged; the value is replaced with “dead”.

###Varying RPC Times
Each node outputs IDs at a constant rate relative to their own system environment time.
