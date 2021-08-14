# About

The assignment is to make a lww state base graph for solving the conflicts of mutations to the graph while there are multiple versions of different replicas in distributed system.

Task:

- Study: 
  - LWW-Element-Set
  - Graph
- Implementation:
  - state-based LWW-Element-Graph
  - test cases
    - clearly written
    - document what aspect of CRDT they test
  - 4 hours (failed)
- Requirements:
  - add a vertex
  - remove a vertex
  - check if a vertex is in the graph
  - add a edge
  - remove a edge
  - find any path between two vertices
  - query all connected vertices of vertex
  - merge with concurrent changes from other graph/replica


# Background

## Replicas

In distributed system, there are replicas, and concurrent writes happened all the time, so there multiple versions:

```
replica_1(a = 1) -> [update (a = 2)] -> replica_1(a = 2)
                                                        
replica_2(a = 1) -> [update (a = 3)] -> replica_1(a = 3)
```

For solving these kind of problem, versioning by timestamp will be a good option. However, the physical clock of the machine of different replicas are not in sync, so there are multiple solutions like lamport clock, vector clock, version vector, ntp or a centralized machine for solving the clock issue. But they can only help with discovering the conflicts but not the conflict itself. The solutions would by something like CRDTs. LWW etc.

## Graph

graph is a set of vertices connected with each others, the connection between vertices is edge, the edge can be directional or nondirectional.

There are few things that need to be clarify:
- vertex
  - a node, the smallest, the basic component of the graph
- edge
  - pair of the vertex, pair two points and form a line
  - can be directional or not
- graph
  - set of vertex connected
  - the graph is directed-graph or undirected-graph base on the edges

Example:

Undirected-graph (cyclic):
```
A - B - D - E
 \ /  /
  C -
```


Undirected-graph (acyclic):
```
A
|\
B C
|\
D E
```

Directed-graph (cyclic):
```
A -> B -> D
|    |
v    v
C -> E
                
```

Directed-graph (acyclic):
```
A -> B -> D
     |    |
     v    v
     C    E
```

# CRDTs

## About state-based CRDTs

- To keep consistency without consensus
- Guarantee convergence to the same value in spite of:
    - network delays
    - partitions 
    - message reordering
- associative(A U B) U C = A U (B U C)
  - grouping would not affect the consistency
- commutative(A U B = B U A)
  - order would not affect the consistency
- idempotent(A U A = A)
  - duplication would not affect the consistency

## Description

There are several things and concepts in CRDTs:

- Copies(replicas)
- Operations
- States
  
### Copies

There are copies(replicas):

```
Master(version_m0) ---> Replica_1(version_m0_r1)
                 \
                   ---> Replica_2(version_m0_r2)

```
version_m0_r1 and version_m0_r2 are copies of version_m0 object (operations or states)

### Operations

- update
- query
- remove
- merge

# State-based replication

In terms of state-based, it means during merge or replicate, the owners of the replicas exchange the modified versions of the state(payload) of copies.

# LWW (State based)

Last write win is the strategy to pick the last change by timestamp as reference across different versions from replicas in terms of operation-based state-based to overwrite the previous state or operation.

For lww set, there are no real removal in term of erasing the payload. The operations or states of removal or adds are entries that will be kept. There are two types of set to keep the entries, adds and removal.

When adds happened, there are cases of the state:

- not exist in adds and removal sets
- exists in adds sets
- exists in adds and removal sets

As the entry contain a timestamp, the entry will append to the add set if not exist, or update the privious record with greater timestamp.

When removal happened, there are cases of the state:

- not exist in adds and removal sets
- exists in adds sets
- exists in adds and removal sets

For most of the use cases, even the removal object is not in or in either of the sets, the record can still append or update the remove set.

When merge happened, there are cases for the source:

- not exist in set
- exist in the set

No matter it is exist or not, the union will be the outcome, however if there are conflict on the timestamp of same state, which means the state is in add and remove sets with same timestamp, the bias of the replica itself will take place of determining which is the choice.

When conflict(timestamp equal) happened, there are cases for the bias:

- adds, the state exist
- removal, the state not exist 

# Implementation

## LWW State-Based Graph

Let's go through the task:

- make a graph contain and can do:
  - compoments:
    - vertex
    - edge
  - operations:
    - add one vertex
    - add edge between vertices
    - remove one vertex
    - remove one edge
    - check is components(vertex or edge) exist
    - search all the paths from one vertex to another
    - search all neighbours of a vertex

In terms of lww:

## Components

- Vertex
- Edge
- Clock
- Tombstone
- Graph data structure
  
### Vertex and Edge

For a graph, there sould be nodes, connection path(s) between two notes, and direction(s). For the task, as it's an undirected graph, there should be only nodes, one path between two node, and no direction of the connections.

### Clock

Assume different replicas store in different machine across network, the physical clock will be not in sync, so there should be implementation like like lamport clock, vector clock, version vector, ntp or a centralized machine for solving the clock issue for syncronization of time across machines. Although right now using physical clock, there should be a interface that use to implement the component.

### Tombstone

The delete set for the vertex or the edge. 

### Graph data structure

Adjacency Matrix and Adjacency List are in used for maintain the record

## Design

### Existence

For any of the components, they need to have the add and remove sets for the entries. However, as it is a graph, edges can only be added or removed when vertices exist, so the remove set of edge can only be updated when two different vertices exist. Also, the vertex can only be removed when it exist in local replica, to make the replica as a set of payloads which is reasonable to form a graph. Moreover, when merge operation comes into play, there might be unpredictable behaviours after solving conflicts:

Assume all of the other components are created before or equal to t0:

```
timeline
|
v
t1 Source Add V(D) at t1
|
v
t2 <- Replica Remove non-exist V(D) in graph at t2
|
v
t3 Source Add E(D, B), E(D,E) at t3
|
v
t4 Merge

```

Source at t1
```
A - B  D  E
\ /  
 C
```

Source at t3
```
A - B - D - E
\ /   /
 C --
```

Replica at t2
```
A - B - E
\ /  /
 C -
```

Source at t4
```
A - B - E
\ /  /
 C -
```

As Replica's E(B,E) is not exist priviously in source, so the edge will be merged into source. However, as the E(D,E) and E(B,D) are added at t3, they should be exist so the easiest way is to check the existence.

### Components dependencies

For Vertex, vertex form edge so edges depending on the vertices, remove vertex means the connections should be removed.

For Edge, it forms by two vertices, so when removing edge, need to check existance of the vertices.

## Search

This part discuss:
- operations
  - check is components(vertex or edge) exist
  - search all the paths from one vertex to another
  - search all neighbours of a vertex

### Check existence

For the vertex and edge, they exist when they are: 

 - in the add set and not in the remove set
 - in the add set and in the remove set, but the add set has a greater timestamp
 - in the add and remove set with same timestamp with adds bias

The graph implementation has a dictionary for vertex add set and one for remove set to loop through the vertex list to get the record from add set and remove set, and the dictionaries will expend when add or remove action happened. After that, it will check the bias and compare with the timestamps.

For searching the paths and neighours of vertex(s), there are two major parts of the implementation:

1. Adjacency Matrix and Adjacency List for record
2. Depth First Search for path finding


#### Adjacency Matrix and Adjacency List

The Adjacency Matrix records the existances of the vertices and the graph recorded every entry into the matrix

Example:
```
A - B - D - E
\ /    /
 C ---
```
can be reprecented by
```
m\n	A	B	C	D	E
A   -	1	1	1	0
B   1	-	1	1	0
C   1	1	-	1	0
D   1	1	1	-	1
E   0	0	0	1	-
```

The Adjacency Matrix mark the vertices of a graph on the matrix and it recorded the connection information. 

When there is adds operations, the record will append or update the record in the matrix, when there is removal, the tombstone matrix will take the play.

When checking the neighbours of a vertex, it can be retreived by the matrix.


#### Depth First Search

For path finding, the DFS first travels to the farthest distance from the starting point which means DFS prioritized far-reaching at the first place for searching through graph. 

## Tests

The test functions achieved:

1. Correctness of code functionality
2. Correctness of behavior

And it is good to be:

1. Protected from regression
2. Refactoring resistance
3. Fast feedback
4. Maintainable

However, basically, if I want bug free codes, test coverage 100%; if there is refactoring in terms of making new feature on existing codes without changing interfaces, unit test failed 100%. Therefore, I tend to write unit test that can run fast and check for correctness with large enough coverage.

Most of the cases, there are three parts of my test functions, 

1. arrangement
2. act
3. assert

Make long story short, usally I mock the graph(s) and run the function under differnet scenarios, and see if the result match the expectations. To make test cases easier to extend, in terms of no interface changes for the functions on the long run, I stick to the table driven style of test, run the test function with all of the cases to ensure the function of its correctness from different aspects.

### Aspect of CRDTs

- associative(A U B) U C = A U (B U C)
- commutative(A U B = B U A)
- idempotent(A U A = A)

There are unit tests for these three aspect when different replicas merge together. Those functions merge in different order, different grouping, and also duplicated, and the results are identical and checked by retrieve the adjacency list, which is the relationship of different vertices, if the results are identical, they are consider as positive consistence result.

# Postscript

Actually I immidiately felt I write shitty code again, for example, I wanted to make design clean by abstract the lww set but actually I didn't, I put into the graph directly for easier to retrieve value by the map for different actions. Also the test cases seems like not really look easy to for people who doesn't familar with what's going on for the code.

Hope you enjoy the work I did. I felt relieve as I tried to make it as a completed piece. Thanks for the chance and it is always interesting for challenge like this task. Looking forward for the feedback :). Thanks for your time.