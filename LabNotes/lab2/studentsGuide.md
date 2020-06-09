# Student Guide



## Implementing Raft

- [x] Reset election timmer

> If election timeout elapses without receiving `AppendEntries` RPC *from current leader* or *granting* vote to candidate: convert to candidate.

### The importance of details

- [ ] From the guide: "A leader will occasionally (at least once per heartbeat interval) send out an `AppendEntries` RPC to all peers to prevent them from starting a new election. If the leader has no new entries to send to a particular peer, the `AppendEntries` RPC contains no entries, and is considered a heartbeat. 

  So heartbeat is not special! It's just an `AppendEntries` RPC call. The leader should send out `AppendEntries` to all peers at least once per heartbeat interval. For each peer, if the peer's log is not identical with the leader's, then `args.Entries` would be non-empty. Otherwise, `args.Entries` would be empty and the `AppendEntries` is called a "hearbeat". Thus, no separate routine is needed for heartbeat and it's implicitly handled in the routine that sends `AppendEntries` to sync the log of leaders and its peers.

  - [ ] Mistake 1: returning success on receiving heartbeats, without doing any checking. 

  - [ ] Mistake 2: Upon receiving heartbeat, truncate the follower's log following `prevLogIndex` and append any entries included in the `AppendEntries` argument. This is wrong. From Figure 2:

  > *If* an existing entry conflicts with a new one (same index but different terms), delete the existing entry and all that follow it.

  The *if* here is crucial. If the follower has all entries the leader sent, it **MUST NOT** truncate the log. Any elements *following* `args.Entries` **MUST** be kept. This is because we could receive outdated `AppendEntries` from the leader. 



## Livelocks

- You should restart election timer *only* if a) get an `AppendEntries` from the *current* leader (check `term` argument); b) you're starting an election; c) you *grant* a vote to another peer.

- If you're a candaite and the election timer fires, you should start *another* election. 

- Before handling incoming RPC:

  > If RPC request or response contains term `T > currentTerm`: set `currentTerm = T`, convert to follower (§5.1)

  For example, if you have already voted in the current term, and an incoming `RequestVote` PRC has higher term, you should *first* step down and adopt the term (and resetting `votedFor`) and *then* handles the RPC, which means granting the vote!



## Incorrect RPC handlers

