## Lab3a

Lab3a is about implementing a KV store based on raft. Each client talks to a number of servers, while each server runs on a Raft instance.

```
  +-------------+
  |				|
  +-------------|-----------+				
  |				V			V
Client 		+-------+	+-------+	+-------+
            |  k/v  |	|  k/v  |	|  k/v  |
            +-------+	+-------+ 	+-------+
            |  Raft	|	|  Raft	|	|  Raft	|
            +-------+	+-------+	+-------+

```

### A common bug
Note that in `Get` and `PutAppend`, you should use `sleep` instead of channel or
condition variable. The reason is that if there's no event to trigger the sending
on the channel or waking up `Get`/`PutAppend` waiting on the conditional variable,
the server gets stuck. Refer to `server.go-bug`.

### A buggy implementation: `server.go-bug`
In `readApplyChRoutine`, whenever a valid `ApplyMsg` is received, store relevant information in `committedOps`. The `Get` and `PutAppend` handler would check that data structure several times for the `OpID`. If found, the handler would apply the command and remove the `OpID` from the data strcure. This easily passes the first 7 testcases, but fail the others. 

### What's the problem?
This impelenmentaion has a important flaw. The command is executed only on the leader, which receives the `Get`/`PutAppend` RPC call and waits for the `OpID` to appears in the `committedOps`. For other followers, they received the `Put`/`AppendEntries` call but doesn't check the `committedOps` and thus would never execute any command! So if we have no failure or network partition, the only server is alwasy alive this implementation works. But surely it would not work if failure/partition happens! This's fixed in `server.go`.


### Reconsider the definition of "Commit"
Now reconsider what's the meaning of a command being **committed** and gets received by the k/v server on the `applyCh`. It means that the server should definitely **execute** the command! So the correct way to execute the command immediately in `readApplyChRoutine`. **This also guarantees the linearizablity** since we're receving commands in log order!


### Evicted leader
As mentioned in the spec, you need to check for evicted leaders.



### Sidenote
Compared to Lab2, this time we don't have detailed instruction to refer to. This gives us more freedom but also requires more careful thoughts before the implementation.  
My implementation uses `sleep`, which is easy but might not be the most neat way.