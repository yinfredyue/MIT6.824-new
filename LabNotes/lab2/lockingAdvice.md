# Locking Advice

- Use locks to protect (1) data that are accessed by multiple threads, and (2) invariants.

- Don't hold lock when doing things that might block: reading/writing a Go channel, sleeping, waiting for RPC reply. This degrades performance and could cause deadlocks. 

- Lots of things can happen between when the goroutine is created and when it runs. 

  ```
  rf.mu.Lock()
  rf.currentTerm += 1
  rf.state = Candidate
  for <each peer> {
      go func() {
          rf.mu.Lock()
          args.Term = rf.currentTerm // Problem!
          rf.mu.Unlock()
          Call("Raft.RequestVote", &args, ...)
          // handle the reply...
      } ()
  }
  rf.mu.Unlock()
  ```

  The problem is that `args.Term` might not be the same as the `rf.currentTerm` when `rf.state = Candidate` is executed. So the term should be passed to the goroutine as an argument. 

  Similarly, reply-handling code should re-check all relevant assumptions when sending the RPCs. For example, it should check that `rf.currentTerm` hasn't changed. 



A pragmatic way to apply the rules:

- If there's no concurrency, then no locks are needed. But we want to send RPCs in parallel and the RPC handler is executed as a separate goroutine (this's how Go's rpc library works).
- So, crudely, we can eliminate the concurrency by acquiring the lock at the start of every goroutine, and releasing the lock only when that goroutine is finished. 
- Fix the code to avoid holding lock when the program blocks. 