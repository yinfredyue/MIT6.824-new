# Lec5

## Concurrency Control in Go

- Use outer variable in closures within loop: `loop.go`
- Conditional varialbe: `vote-count-4.go`, `cond.txt`
- WaitGroup: `loop.go`. Note that `wg.Done()` is never called before `wg.Add(1)`: `wg.Add(1)` is called outside the goroutine.
- Try mutex and conditional variables before turning to channels.



## Raft

- Don't hold the lock while making RPC calls. This can cause deadlock.
- Might nee to check `rf.currTerm` sometimes.



Note: The reading about the Go memory model is not necessary. As mentioned in the beginning of the reading as well as by the TA, you're being too clever if you need that knowledge.