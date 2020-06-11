## Lab2

This lab is about implementing Raft. Effectively, it covers up to section 5 in the paper. Section 6 is not discussed and section 7 ~ 8 is in Lab3.

Thoughts:

1. Reading the provided material is important. Other than the paper itself, the student guide, the two guides (locking, structure) on Raft are all important and very insightful. Read them carefully!

2. In my initial implementation, I created many background routine in `Make`, to handle operations like commit, apply, election timeout, send heartbeat, send log replication request. This is ok, but less efficient and could cause weird bugs (one is mentioned in the comment of `raft.go`). So following the suggestion in the Raft structure guide that:

    > ```
    > It's easiest to do the RPC reply processing in the same goroutine, rather than sending reply information over a channel.
    > ```

    I create only two background routines in `Make`: `electionTimeoutRoutine` and `applyCommandRoutine`. For other operations, I either handle the reply directly when I receive them, or start them only when needed (like `periodicHeartbeatRoutine` and `logReplicationRoutine`).

3. Design issue. How to set up the `applyCommandRoutine`? The guide mentions that:

    > ```
    > You'll want to have a separate long-running goroutine that sends
    > committed log entries in order on the applyCh. It must be separate,
    > since sending on the applyCh can block; and it must be a single
    > goroutine, since otherwise it may be hard to ensure that you send log
    > entries in log order. The code that advances commitIndex will need to
    > kick the apply goroutine; it's probably easiest to use a condition
    > variable (Go's sync.Cond) for this.
    > ```

    So there're 3 issues: seprate go routine (to avoid blocking), in log order, how to kick it? 

4. There're actually many similaries between heartbeat and normal AppendEntries. Actually, heartbeat should be different from AppendEntries only in the `entries` send, the `AppendEntries` RPC handler should handle them in the same way and the reply from heartbeat and normal AppendEntries should be handled in the same way! This is done in `sendAppendEntriesAndHandleReply`.

5. The relationship between heartbeats and normal AppendEntries. From `studentGuide.md`:

    > From the guide: "A leader will occasionally (at least once per heartbeat interval) send out an `AppendEntries` RPC to all peers to prevent them from starting a new election. If the leader has no new entries to send to a particular peer, the `AppendEntries` RPC contains no entries, and is considered a heartbeat. 
    >
    > So heartbeat is not special! It's just an `AppendEntries` RPC call. The leader should send out `AppendEntries` to all peers at least once per heartbeat interval. For each peer, if the peer's log is not identical with the leader's, then `args.Entries` would be non-empty. Otherwise, `args.Entries` would be empty and the `AppendEntries` is called a "hearbeat". Thus, no separate routine is needed for heartbeat and it's implicitly handled in the routine that sends `AppendEntries` to sync the log of leaders and its peers.

    So we are sending periodic `AppendEntries` essentially! Check `raft.go:periodicAppendEntriesRoutine()` and `raft.go:sendAppendEntriesToPeers()`.

6. Fast backup is important to pass the test in 2B and 2C. If you implement it in 2B, 2C would not take too many efforts. The lecture notes explains the mechanism clearly, but there're can be subtle bugs during implementation.