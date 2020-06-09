# Structure Advice

A Raft instance needs to handle external events (RPC handlers) and execute periodic tasks (elections and heartbeats).

Raft instance states would be updated in response to the external events. Use shared data + locks. 

Create a separate goroutine for the election timeout, which periodically checks if the timeout is reached. Use `time.Sleep()`. 

You need a separate long-running goroutine to send committed log entries in order on `applyCh`. It must be a separate goroutine, to avoid blocking. Also, it must be a single goroutine, to maintain the log. The code that advances `commitIndex` should kick this routine. You can use a condition variable. 

Each RPC should probably be sent in its own goroutine, and it's reply should probably also be processed in a separate goroutine. It's easiest to do the RPC reply processing in the same goroutine, rather than sending reply information over a channel.

The leader must be careful when processing replies. It must check that the term hasn't changed since sending the RPC, and must handle the possibility that replies from concurrent RPCs to the same follower have changed its own state (e.g. nextIndex).