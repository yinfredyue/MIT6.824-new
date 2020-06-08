# Lec3

## Distributed storage system

- Why hard?

  Performance -> sharding

  Failures -> Fault tolerance

  Fault tolerance -> replication

  Replication -> Inconsistency

  Consistency -> Low Performance

- Strong consistency: behave like a single server.

- Bad replication design

   

## GFS

Big, fast, general-purpose, sharding, automatic recover, single data center, internal use, optimized for big sequential read/write.

- Master data strctures

  Table 1: file name -> list of chunk handles (disk)

  Table 2: chunk handle -> list of chunk servers holding the chunk, version number (disk), primary chunkserver, lease expiration.

  The two tables are in RAM, and also consistently flushed to the on-disk log for each mutation. Checkpoints are also used. 

- Read

  - name, offset -> Master
  - Master sends back chunk handle, and list of chunkservers. Client caches this info.
  - Client reads the data from one of the chunkservers.

- Record append

  - name, offset -> Master

  - If no primary, the master needs to select a primary. It finds the up-to-date replica, whose version number equals to the version number stored at the master. Make it as the primary. (The chunkserver also remembers the version number of each chunk it stores). The master increments the version number, write log to disk, and tells the primary and secondaries the new version number. Then reply the client with the primary and secondaries. 

  - The client sends data to primary and secondaries.

  - The primary picks an offset, and tells all replicas (including itself) to write the data to that offset. If all reply YES, the primary replies success to the client. Otherwise, the primary replies failure to the client. The client should re-issue the request. 

    If the failure case above happens, the data has been written to a subset of replicas, so the replicas are no longer identical. This is allowed in GFS.

- The lease avoids having two primaries for the same chunk at the same time, "split brain", caused by *network partition*. 

  Why is this a problem? Doesn't the client always talks to the master for server information?

  - The client caches the information.
  - Even if the client doesn't cache the information, conflict could happen. Suppose the master tells client that S1 is the primary, but immediately after sending out this message, it chooses S2 as the primary. Then S1 and S2 are serving clients at the same time and S1 could reply success to client. 

- Example

  ```
  Primary
  
  D				D
  B		B		B
  C		C		C
B		B
  A		A		A
  ```
  
  The application should be able to tolerate the situation. 



