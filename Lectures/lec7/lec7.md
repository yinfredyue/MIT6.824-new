# Lec 8

## Linearizability

An execution history is linearizable if (1) there exists an order of operations that matches real-time order for non-concurrent operations, and (2) each read sees the most recent write in order.

Linearizable:

```
  |-Wx1-| |-Wx2-|
    |---Rx2---|
      |-Rx1-|
```

Not linearizable:

```
  |-Wx1-| |-Wx2-|
    |--Rx2--|
              |-Rx1-|
```

Two strategy to show if a system is linearizable or non-linearizable:

1. Find a sequence of operations that satisfy the outcome;
2. Find a cycle in the dependency graph.

Linearizable:

```
|--Wx0--|  |--Wx1--|
            |--Wx2--|
        |-Rx2-| |-Rx1-|
```

Non-linearizable:

```
|--Wx0--|  |--Wx1--|
            |--Wx2--|
C1:     |-Rx2-| |-Rx1-|
C2:     |-Rx1-| |-Rx2-|
```

Linearizablity is about the history, the behavior of the system. It's not a definition about system design, but about the behavior of the system. 

Non-linearizable:

```
|-Wx1-|
        |-Wx2-|
                |-Rx1-|
```

suppose clients re-send requests if they don't get a reply
in case it was the response that was lost:
  leader remembers client requests it has already seen
  if sees duplicate, replies with saved response from first execution
but this may yield a saved value from long ago -- a stale value!
what does linearizabilty say?

```
C1: |-Wx3-|          |-Wx4-|
C2:          |-Rx3-------------|
```

order: Wx3 Rx3 Wx4
so: returning the old saved value 3 is correct, linearizable!