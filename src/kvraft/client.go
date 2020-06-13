package kvraft

import (
	"crypto/rand"
	"math/big"

	"../labrpc"
)

// The client talks to the service through a Clerk with Put/Append/Get methods.
// A Clert manages RPC connections with the servers.
// Clerk sends Put(), Append(), and Get() RPCs to the kvservers whose assoicated
// Raft is the leader. The kvserver code submits the operation to Raft. All the
// kvservers exeucte operations from Raft log in order, applying them to the
// key/value database. This guarantees identical replicas.
// If the operation is committed, the leader reports the result to the Clerk
// by responding to its RPC. Otherwise, the server reports an error, and the
// Clerk retries.
// So client (client.go) talks to server (server.go). The server talks with
// the Raft instance (raft.go).
type Clerk struct {
	servers []*labrpc.ClientEnd

	// You will have to modify this struct.
	prevLeader int
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

func MakeClerk(servers []*labrpc.ClientEnd) *Clerk {
	ck := new(Clerk)
	ck.servers = servers
	// You'll have to add code here.
	return ck
}

//
// fetch the current value for a key.
// returns "" if the key does not exist.
// keeps trying forever in the face of all other errors.
//
// you can send an RPC with code like this:
// ok := ck.servers[i].Call("KVServer.Get", &args, &reply)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
//
func (ck *Clerk) Get(key string) string {
	for i := ck.prevLeader; ; i = (i + 1) % len(ck.servers) {
		args := GetArgs{key}
		reply := GetReply{}
		ok := ck.servers[i].Call("KVServer.Get", &args, &reply)

		if !ok {
			DPrintf("Client cannot contact [Server %v]", i)
			continue
		}

		// Received reply
		switch reply.Err {
		case OK:
			ck.prevLeader = i
			return reply.Value
		case ErrNoKey:
			ck.prevLeader = i
			return ""
		case ErrWrongLeader:
			continue
		}
	}
}

//
// shared by Put and Append.
//
// you can send an RPC with code like this:
// ok := ck.servers[i].Call("KVServer.PutAppend", &args, &reply)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
//
func (ck *Clerk) PutAppend(key string, value string, op string) {
	// You will have to modify this function.
	for i := ck.prevLeader; ; i = (i + 1) % len(ck.servers) {
		args := PutAppendArgs{key, value, op}
		reply := PutAppendReply{}
		ok := ck.servers[i].Call("KVServer.PutAppend", &args, &reply)

		if !ok {
			DPrintf("Client cannot contact [Server %v]", i)
			continue
		}

		// Received reply
		switch reply.Err {
		case OK:
			ck.prevLeader = i
			return
		case ErrNoKey:
			DPrintf("Unexpected: PutAppend gets ErrNoKey")
			return
		case ErrWrongLeader:
			continue
		}
	}
}

func (ck *Clerk) Put(key string, value string) {
	ck.PutAppend(key, value, "Put")
}
func (ck *Clerk) Append(key string, value string) {
	ck.PutAppend(key, value, "Append")
}
