# Lec2

- Why Go?

  Threads, garbage collection, convenient RPC.

- Goroutine

  Threads, sharing memory space. 

  Using thread is simpler than event-driven code.



## Concurrency Control

- Web crawler

  Note that we need to make a copy when creating goroutines within the for loop.

  If a closure uses a variable in the outer function, Go compiler would allocate that variable on the heap so that the closure can use it even after the outer function returns. 

  [Shared memory + lock] VS [channel].

  - ConcurrentMutex

    Lock(), Unlock()

    sync.WaitGroup

  - ConcurrentChannel

    The master keeps a count of workers in `n` to determine the time to quit.

- It's not a race that multiple threads use the same channel since the internal implementation of channel avoids this using concurrency control primitives, like mutex. 

- When to use sharing and locks, versus channels?

  Most problems can be solved in both ways.

  Depends on how the programmer reasons about the problem.

  - State: sharing and locks
  - Communication: channels

  For the labs, use sharing+locks for state, `sync.Cond`/channels/`time.Sleep()` for waiting/notification. 



## Remote Procedure Call (RPC)

- Software structure

  ```
  client app        handler fns
   stub fns         dispatcher
   RPC lib           RPC lib
     net  ------------ net
  ```

- Read comments in `kv.go`

- Go RPC behavior

  Go RPC writes request only once to the TCP connection. If no reply, returns an error. 