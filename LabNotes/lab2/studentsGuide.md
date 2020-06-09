# Student Guide

- Reset election timer

  > If election timeout elapses without receiving `AppendEntries` RPC *from current leader* or *granting* vote to candidate: convert to candidate.

- A heartbeat should be treated just like an `AppendEntries` RPC. Wrong: return success directly without checking.
- When receiving `AppendEntries` request, truncate log only if conflict exists.



## Livelocks

- You should restart election timer *only* if a) get an `AppendEntries` from the *current* leader (check `term` argument); b) you're starting an election; c) you *grant* a vote to another peer.

- If you're a candaite and the election timer fires, you should start *another* election. 

- Before handling incoming RPC:

  > If RPC request or response contains term `T > currentTerm`: set `currentTerm = T`, convert to follower (§5.1)

  For example, if you have already voted in the current term, and an incoming `RequestVote` PRC has higher term, you should *first* step down and adopt the term (and resetting `votedFor`) and *then* handles the RPC, which means granting the vote!



## Incorrect RPC handlers

