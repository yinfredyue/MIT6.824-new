# Lec1

## Overview

Goal: Parallelism, fault-tolerance via replication, physically distributed.

Challenges: concurrency, partial failure, scalability. 

Infrastructure for applications: (1) storage, (2) communication, (3) computation. The goal is to provide the abstractions that hide the complexity of distribution.

-  Implementation

  RPC, threads, concurrency control, etc.

- Performance

  Goal: scalability. Handling more loads only requires buying more machines, instead of re-designing the program. 

- Fault tolerance

  Failure is the normal case in a big cluster of computers. We want

  - Availability: apps can make progress despite failures
  - Recoverability: apps will come back to functioning when failures are fixed

  Two main tools: (1) Non-volatile storage, (2) replicated servers. 

- Consistency

  Consistency and performance are hard to achieve at the same time. 

  Strong consistency requires many communications. Many design only provide weak consistency for better performance. 



## MapReduce

Goal: provide framework for non-specialist programmers to write distributed applications. MapReduce framework hides all details of the distribution.

```
Input1 -> Map -> a,1 b,1
Input2 -> Map ->     b,1
Input3 -> Map -> a,1     c,1
                  |   |   |
                  |   |   -> Reduce -> c,1
                  |   -----> Reduce -> b,2
                  ---------> Reduce -> a,2

Map: (k1, v1) -> list of (k2, v2)
Reduce: (k2, list of v2) -> list(v2)
```

For each input file, Map produces set of (k2, v2) intemediate results. MR gathers all intermediate v2's for a given k2, and passes (k2, list of v2) to a Reduce call.

Example: word count.

```
Map(k, v):
  split v into words
  	for each word w
  emit(w, "1")

Reduce(k, v):
	emit(len(v))
```

MapReduce scales well, as Map's can run in parallel, so do Reduce's. 

MapReduce hides details about distributed system, but limits what apps can do: Map and Reduce must be functional.

GFS helps MR to achieve better parallelism. In 2004, the bottleneck of MR is the network bandwidth.

The master gives Map tasks to workers until all Maps complete. Map splits output by hash into one file per Reduce task, and writes output to the local disk. Master starts handing out Reduce tasks only after all Map tasks are done. 

- Tricks used by MR to minimize network use: 
  - Master tries to run Map tasks on GFS server that stores its input
  - Intermediate data goes over the network only once, as Map workers write to local disk. Reduce workers read from Map workers, not via GFS.

- MR gets good load balancing by:
  - Assign backup tasks to avoid stragglers
  - Much more tasks than workers

- MR gets fault tolerance by re-execution. MR re-runs the task on failure, which is ok as Map and Reduce are functional and deterministic
  - On Map worker crash
    - Completed task: Master marks the task handed to this worker as idle, and hands out to other workers. This's because the intermediate result is stored on the local disk and inaccessible.
    - In-progress task: reset to idle and assign to other workers.
  - On Reduce worker crash
    - Completed task: Finished tasks are ok, already stored on GFS.
    - In-progress task: reset to idle and assign to other workers.