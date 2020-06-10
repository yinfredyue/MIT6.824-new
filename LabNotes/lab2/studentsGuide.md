# Student Guide



## Implementing Raft

- [x] Reset election timmer

> If election timeout elapses without receiving `AppendEntries` RPC *from current leader* or *granting* vote to candidate: convert to candidate.

### The importance of details

- [x] From the guide: "A leader will occasionally (at least once per heartbeat interval) send out an `AppendEntries` RPC to all peers to prevent them from starting a new election. If the leader has no new entries to send to a particular peer, the `AppendEntries` RPC contains no entries, and is considered a heartbeat. 

  So heartbeat is not special! It's just an `AppendEntries` RPC call. The leader should send out `AppendEntries` to all peers at least once per heartbeat interval. For each peer, if the peer's log is not identical with the leader's, then `args.Entries` would be non-empty. Otherwise, `args.Entries` would be empty and the `AppendEntries` is called a "hearbeat". Thus, no separate routine is needed for heartbeat and it's implicitly handled in the routine that sends `AppendEntries` to sync the log of leaders and its peers.

  - [x] Mistake 1: returning success on receiving heartbeats, without doing any checking. 

  - [x] Mistake 2: Upon receiving heartbeat, truncate the follower's log following `prevLogIndex` and append any entries included in the `AppendEntries` argument. This is wrong. From Figure 2:

  > *If* an existing entry conflicts with a new one (same index but different terms), delete the existing entry and all that follow it.

  The *if* here is crucial. If the follower has all entries the leader sent, it **MUST NOT** truncate the log. Any elements *following* `args.Entries` **MUST** be kept. This is because we could receive outdated `AppendEntries` from the leader. 



## Debugging Raft

### Livelocks

- [x] You should restart election timer *only* if a) get an `AppendEntries` from the *current* leader (check `term` argument); b) you're starting an election; c) you *grant* a vote to another peer.

- [x] If you're a candaite and the election timer fires, you should start *another* election. 

- [x] Before handling incoming RPC:

  > If RPC request or response contains term `T > currentTerm`: set `currentTerm = T`, convert to follower (§5.1)

  For example, if you have already voted in the current term, and an incoming `RequestVote` PRC has higher term, you should *first* step down and adopt the term (and resetting `votedFor`) and *then* handles the RPC, which means granting the vote!



### Incorrect RPC handlers

- [ ] If a step says “reply false”, this means you should *reply immediately*, and not perform any of the subsequent steps.

- [x] If you get an `AppendEntries` RPC with a `prevLogIndex` that points beyond the end of your log, you should handle it the same as if you did have that entry but the term did not match (i.e., reply false).

- [x] Check 2 for the `AppendEntries` RPC handler should be executed *even if the leader didn’t send any entries*.

- [x] The `min` in the final step (#5) of `AppendEntries` is *necessary*, and it needs to be computed with the index of the last *new* entry. It is *not* sufficient to simply have the function that applies things (i.e. applying log entries) from your log between `lastApplied` and `commitIndex` stop when it reaches the end of your log. This is because you may have entries in your log that differ from the leader’s log *after* the entries that the leader sent you (which all match the ones in your log). Because #3 dictates that you only truncate your log *if* you have conflicting entries, those won’t be removed, and if `leaderCommit` is beyond the entries the leader sent you, you may apply incorrect entries.

- [x] It is important to implement the “up-to-date log” check *exactly* as described in section 5.4. No cheating and just checking the length!



### Failure to follow the Rules

- [x] If `commitIndex > lastApplied` *at any point* during execution, you should apply a particular log entry. It is not crucial that you do it straight away (for example, in the `AppendEntries` RPC handler), but it *is* important that you ensure that this application is only done by one entity. Specifically, you will need to either have a dedicated “applier”, or to lock around these applies, so that some other routine doesn’t also detect that entries need to be applied and also tries to apply.

- [x] Make sure that you check for `commitIndex > lastApplied` either periodically, or after `commitIndex` is updated (i.e., after `matchIndex` is updated). For example, if you check `commitIndex` at the same time as sending out `AppendEntries` to peers, you may have to wait until the *next* entry is appended to the log before applying the entry you just sent out and got acknowledged.

- [x] If a leader sends out an `AppendEntries` RPC, and it is rejected, but *not because of log inconsistency* (this can only happen if our term has passed), then you should immediately step down, and *not* update `nextIndex`. If you do, you could race with the resetting of `nextIndex` if you are re-elected immediately.

- [ ] A leader is not allowed to update `commitIndex` to somewhere in a *previous* term (or, for that matter, a future term). Thus, as the rule says, you specifically need to check that `log[N].term == currentTerm`. This is because Raft leaders cannot be sure an entry is actually committed (and will not ever be changed in the future) if it’s not from their current term. This is illustrated by Figure 8 in the paper.