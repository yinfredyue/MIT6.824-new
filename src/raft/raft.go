package raft

/*
Lab 2A
- [X] This Lab cares about: Sending/Receiving RequestVote RPCs, election-related
rules for servers, and state related to leader election.
- [X] Add Figure 2 states for leader election to the Raft struct. You need to
define a struct to hold information about each log entry.
- [X] Fill in the RequestVoteArgs and RequestVoteReply structs. Modify Make()
to create a background goroutine that will kick off leader election
periodically by sending out RequestVote RPCs when it hasn't heard from another
peer for a while. This way a peer will learn who is the leader, if there is
already a leader, or become the leader itself. Implement the RequestVote() RPC
handler so that servers will vote for one another.
- [X] To implement heartbeats, define an AppendEntries RPC struct (though you
may not need all the arguments yet), and have the leader send them out
periodically. Write an AppendEntries RPC handler method that resets the
election timeout so that other servers don't step forward as leaders when one
has already been elected.
- [X] Make sure the election timeouts in different peers don't always fire at
the same time, or else all peers will vote only for themselves and no one will
become the leader.
- [X] The tester requires that the leader send heartbeat RPCs no more than ten
times per second. The tester requires your Raft to elect a new leader within
five seconds of the failure of the old leader. The paper's Section 5.2 mentions
election timeouts in the range of 150 to 300 milliseconds. Such a range only
makes sense if the leader sends heartbeats considerably more often than once
per 150 milliseconds. Because the tester limits you to 10 heartbeats per
second, you will have to use an election timeout larger than the paper's 150 to
300 milliseconds, but not too large, because then you may fail to elect a
leader within five seconds.
- [X] You'll need to write code that takes actions periodically or after delays
in time. The easiest way to do this is to create a goroutine with a loop that
calls time.Sleep(). Don't use Go's time.Timer or time.Ticker.
- [X] Don't forget to implement GetState().
- [X] The tester calls your Raft's rf.Kill() when it is permanently shutting
down an instance. You can check whether Kill() has been called using
rf.killed(). You may want to do this in all loops, to avoid having dead Raft
instances print confusing messages.

Lab 2B
- [X] Your first goal should be to pass TestBasicAgree2B(). Start by
implementing Start(), then write the code to send and receive new log entries
via AppendEntries RPCs, following Figure 2.
- [X] You will need to implement the election restriction (section 5.4.1).
= [X] One way to fail to reach agreement in the early Lab 2B tests is to hold
repeated elections even though the leader is alive. Look for bugs in election
timer management, or not sending out heartbeats immediately after winning an
election.
- [X] Avoid loops that repeatedly check for evnets. Use conditional variable or
sleeps to avoid such loops.
*/

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"sync"
	"sync/atomic"
	"time"

	"../labrpc"
)

// import "bytes"
// import "../labgob"

//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in Lab 3 you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh; at that point you can add fields to
// ApplyMsg, but set CommandValid to false for these other uses.
//
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}

type LogEntry struct {
	Command interface{}
	Term    int // Term when entry is received by the leader
}

const (
	noPeer = -1

	followerState  = "follower"
	candidateState = "candidate"
	leaderState    = "leader"
)

//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.

	// Fields explicitly mentioned in Figure 2
	// Persistent
	currentTerm int
	votedFor    int
	log         []LogEntry

	// Volatile
	commitIndex int
	lastApplied int

	// Leader only
	nextIndex  []int
	matchIndex []int

	// Fields not explicitly mentioned in Figure 2
	state             string
	electionTimeout   time.Duration
	prevTime          time.Time // Prev time when received AppendEntries from CURRNET leader or granting vote to candidate
	votesReceived     int
	prevAppendEntries time.Time // Prev time when AppendEntries is sent

	// Immutable quantities
	heartbeatInterval          time.Duration
	electionTimerCheckInterval time.Duration
	heartbeatCheckInterval     time.Duration

	applyCh       chan ApplyMsg
	notifyApplyCh chan bool
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	return rf.currentTerm, rf.state == leaderState
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}

//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	// Your code here (2B).

	rf.mu.Lock()
	defer rf.mu.Unlock()

	if rf.state != leaderState {
		return -1, -1, false
	}

	// Start the agreement process and return immediately
	rf.log = append(rf.log, LogEntry{command, rf.currentTerm})
	DPrintf("[%v] receives command %v from client, currentTerm = %v", rf.me, command, rf.currentTerm)
	DPrintf("    %v", rf.log)
	return len(rf.log) - 1, rf.currentTerm, true
}

//
// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
//
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

/////////////////////// RequestVote (start) //////////////////////

//
// example RequestVote RPC arguments structure.
// field names must start with capital letters!
//
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term         int
	CandidateID  int
	LastLogIndex int
	LastLogTerm  int
}

//
// example RequestVote RPC reply structure.
// field names must start with capital letters!
//
type RequestVoteReply struct {
	// Your data here (2A).
	Term        int
	VoteGranted bool
}

//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).
	rf.mu.Lock()
	defer rf.mu.Unlock()

	// Even if converted to follower, don't return here! Continue handling.
	rf.convertToFollowerIfOutOfTerm(args.Term)

	DPrintf("[%v] receives RequestVote from [%v]", rf.me, args.CandidateID)
	reply.Term = rf.currentTerm

	if args.Term < rf.currentTerm {
		DPrintf("    not granting vote: smaller term", rf.me)
		reply.VoteGranted = false
		return
	}

	myLastLogTerm := rf.log[len(rf.log)-1].Term
	myLastLogIndex := len(rf.log) - 1
	upToDate1 := args.LastLogTerm > myLastLogTerm
	upToDate2 := args.LastLogTerm == myLastLogTerm && args.LastLogIndex >= myLastLogIndex
	upToDate := upToDate1 || upToDate2
	if (rf.votedFor == noPeer || rf.votedFor == args.CandidateID) && upToDate {
		DPrintf("    votes for [%v] in term %v", args.CandidateID, rf.currentTerm)
		rf.votedFor = args.CandidateID
		reply.VoteGranted = true

		// Granting vote to candidate, reset timer
		DPrintf("    resets election timer for granting vote")
		rf.resetElectionTimer()

		return
	}

	DPrintf("    not granting vote")
}

func (rf *Raft) electionTimoutRoutine() {
	for {
		rf.mu.Lock()
		timeout := rf.prevTime.Add(rf.electionTimeout).Before(time.Now())
		if rf.state != leaderState && timeout {
			// Convert to candidate, start election
			rf.currentTerm++
			rf.votedFor = rf.me
			rf.state = candidateState
			rf.resetElectionTimer()
			rf.votesReceived = 1
			DPrintf("[%v] becomes candidate, rf.currentTerm = %v, %v", rf.me, rf.currentTerm, rf.log)

			for i := 0; i < len(rf.peers); i++ {
				if i == rf.me {
					continue
				}

				// Send RequestVote RPCs in parallel
				args := RequestVoteArgs{
					Term:         rf.currentTerm,
					CandidateID:  rf.me,
					LastLogIndex: len(rf.log) - 1,
					LastLogTerm:  rf.log[len(rf.log)-1].Term,
				}
				go rf.sendRequestVoteAndHandleReply(args, i)
			}
		}
		rf.mu.Unlock()

		time.Sleep(rf.electionTimerCheckInterval)
	}
}

func (rf *Raft) sendRequestVoteAndHandleReply(args RequestVoteArgs, server int) {
	reply := RequestVoteReply{}
	ok := rf.sendRequestVote(server, &args, &reply)
	if !ok {
		DPrintf("[%v] sent RequestVote to [%v] but fails to get a reply", rf.me, server)
		return
	}

	// Received reply
	rf.mu.Lock()
	defer rf.mu.Unlock()

	// If out of term
	if rf.convertToFollowerIfOutOfTerm(reply.Term) {
		DPrintf("[%v] is out of term, converts from candidate to follower", rf.me)
		return
	}

	if reply.VoteGranted {
		rf.votesReceived++
		DPrintf("[%v] receives vote from [%v]", rf.me, server)

		if rf.state != leaderState && rf.votesReceived > len(rf.peers)/2 {
			// Receives majority vote, convert to leader
			DPrintf("[%v] elected as LEADER in term %v!", rf.me, rf.currentTerm)
			DPrintf("    %v", rf.log)
			rf.state = leaderState

			// Initialize fields for leader
			rf.nextIndex = make([]int, len(rf.peers))
			rf.matchIndex = make([]int, len(rf.peers))
			for i := 0; i < len(rf.peers); i++ {
				rf.nextIndex[i] = len(rf.log)
				rf.matchIndex[i] = 0
			}

			// Send initial heartbeats to peers in parallel
			DPrintf("    immediately sends heartbeats to peers")
			rf.sendAppendEntriesToPeers()

			// Create routine for sending periodic AppendEntries
			go rf.periodicAppendEntriesRoutine()
		}
	}
}

/////////////////////// RequestVote (end) //////////////////////

/////////////////////// AppendEntries (start) //////////////////////

type AppendEntriesArgs struct {
	Term         int
	LeaderID     int
	PrevLogIndex int // Index of log entry immediately preceeding new ones
	PrevLogTerm  int // Term of prevLogIndex entry
	Entries      []LogEntry
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()

	rf.convertToFollowerIfOutOfTerm(args.Term)

	DPrintf("[%v] receives AppendEntries from [%v]", rf.me, args.LeaderID)
	DPrintf("    %+v", *args)
	DPrintf("    Current log: %v", rf.log)

	if args.Term < rf.currentTerm {
		DPrintf("    replies False: smaller term")
		reply.Term = rf.currentTerm
		reply.Success = false
		defer rf.mu.Unlock()
		return
	}

	// Receives AppendEntries from current leader, reset election timer
	DPrintf("    resets election timer as receiving AppendEntries from leader")
	rf.resetElectionTimer()

	if len(rf.log) < args.PrevLogIndex+1 || rf.log[args.PrevLogIndex].Term != args.PrevLogTerm {
		DPrintf("    replies False: Non existing log entry")
		reply.Term = rf.currentTerm
		reply.Success = false
		defer rf.mu.Unlock()
		return
	}

	DPrintf("    replies True: log entry matches")
	reply.Term = rf.currentTerm
	reply.Success = true

	// Subtle: Read Figure 2 and student's guide carefully.
	// Mentioned in "The importance of details" section of the student guide.
	// Only truncate if conflict exists!
	// rf.log matches the log of the leader at index args.PrevLogIndex
	// Overwrite args.Entries into rf.log

	// Truncate if there's any conflict
	appendStart := -1
	for i := 0; i < len(args.Entries); i++ {
		index := args.PrevLogIndex + 1 + i

		if len(rf.log) > index && rf.log[index] == args.Entries[i] {
			// rf.log[index] exists
			continue
		} else {
			// rf.log[index] doesn't exists or conflict.
			rf.log = rf.log[:index]
			appendStart = i
			break
		}
	}
	DPrintf("    appendStart: %v, rf.log: %v", appendStart, rf.log)

	// Append entries, if needed
	if appendStart >= 0 {
		for i := appendStart; i < len(args.Entries); i++ {
			rf.log = append(rf.log, args.Entries[i])
		}
	}
	DPrintf("    Updated log: %v", rf.log)

	notify := false
	if args.LeaderCommit > rf.commitIndex {
		// The 4-th point in the student guide "Incorrect RPC handlers".
		lastNewEntryIndex := args.PrevLogIndex + len(args.Entries)
		rf.commitIndex = Min(args.LeaderCommit, lastNewEntryIndex)
		notify = true
		DPrintf("    (follower) updates commitIndex to %v", rf.commitIndex)
	}

	rf.mu.Unlock()
	if notify {
		DPrintf("[%v] tries to notify apply", rf.me)
		rf.notifyApplyCh <- true
	}
}

func (rf *Raft) periodicAppendEntriesRoutine() {
	for {
		rf.mu.Lock()
		if rf.state != leaderState {
			defer rf.mu.Unlock()
			return
		}

		if rf.prevAppendEntries.Add(rf.heartbeatInterval).After(time.Now()) {
			rf.mu.Unlock()
			time.Sleep(rf.heartbeatCheckInterval)
			continue
		}

		// HeartbeatInterval reached
		DPrintf("[%v] tries to send periodic AppendEntries to peers", rf.me)
		rf.sendAppendEntriesToPeers()
		rf.prevAppendEntries = time.Now()

		rf.mu.Unlock()
	}
}

// sendAppendEntriesToPeers is called in two situations:
// 1. Immediately after a leader is elected, to suppress new elections;
// 2. Periodically by periodicAppendEntriesRoutine().
// The code in this function doesn't distinguish heartbeat and normal
// AppendEntries RPCs. If a peer's log happens to be in sync, then that
// AppendEntries RPC would be a heartbeat. Other than this, there's no
// difference. This's from "The importance of details" in the student guide.
func (rf *Raft) sendAppendEntriesToPeers() {
	for i := 0; i < len(rf.peers); i++ {
		if i == rf.me {
			continue
		}

		rf.prevAppendEntries = time.Now()
		args := AppendEntriesArgs{
			Term:         rf.currentTerm,
			LeaderID:     rf.me,
			PrevLogIndex: rf.nextIndex[i] - 1, // TOOD
			PrevLogTerm:  rf.log[rf.nextIndex[i]-1].Term,
			Entries:      rf.log[rf.nextIndex[i]:],
			LeaderCommit: rf.commitIndex,
		}
		DPrintf("[%v] sends AppendEntries to [%v]", rf.me, i)
		DPrintf("    %+v", args)
		go func(server int, args AppendEntriesArgs) {
			// Figures 2: retry if AppendEntries fails due to log inconsistency:
			// Thus, if the RPC receives no reply, we don't retry.
			var ok bool

			for {
				reply := AppendEntriesReply{}
				ok = rf.sendAppendEntries(server, &args, &reply)
				if !ok {
					DPrintf("[%v]'s AppendEntries to [%v] fails to receive a reply", rf.me, server)
					return
				}

				// Receives a reply
				rf.mu.Lock()

				// Quit if out of term
				if rf.convertToFollowerIfOutOfTerm(reply.Term) {
					DPrintf("[%v] is out of term, convert from leader to follower", rf.me)
					defer rf.mu.Unlock()
					return
				}

				// Quit if no longer leader
				if rf.state != leaderState {
					DPrintf("[%v] no longer a leader", rf.me)
					defer rf.mu.Unlock()
					return
				}

				if reply.Success {
					DPrintf("[%v] receives SUCCESS AppendEntries from [%v]", rf.me, server)
					DPrintf("    Original: nextIndex: %v, matchIndex: %v", rf.nextIndex[server], rf.matchIndex[server])

					// AppendEntries succeeded
					// update nextIndex and matchIndex for the follower
					rf.nextIndex[server] = args.PrevLogIndex + len(args.Entries) + 1
					rf.matchIndex[server] = Max(0, rf.nextIndex[server]-1)

					DPrintf("    Updated: nextIndex: %v, matchIndex: %v", rf.nextIndex[server], rf.matchIndex[server])

					// Try to increment rf.commitIndex
					oldCommitIndex := rf.commitIndex
					rf.tryCommit()
					newCommitIndex := rf.commitIndex

					rf.mu.Unlock()
					if oldCommitIndex < newCommitIndex {
						DPrintf("[%v] tries to notify apply", rf.me)
						rf.notifyApplyCh <- true
					}
					return
				} else {
					DPrintf("[%v] receives FAIL AppendEntries from [%v]", rf.me, server)
					DPrintf("    Original: nextIndex: %v, matchIndex: %v", rf.nextIndex[server], rf.matchIndex[server])

					// AppendEntries fail due to log inconsistency: decrement
					// nextIndex and retry
					if rf.nextIndex[server] > 1 {
						rf.nextIndex[server]--

						// Subtle: Don't forget to Update args
						args.PrevLogIndex = rf.nextIndex[server] - 1
						args.PrevLogTerm = rf.log[args.PrevLogIndex].Term
						args.Entries = rf.log[rf.nextIndex[server]:]
					}

					DPrintf("    Updated: nextIndex: %v, matchIndex: %v", rf.nextIndex[server], rf.matchIndex[server])
				}

				rf.mu.Unlock()

				DPrintf("[%v]'s AppendEntries to [%v] fail due to log inconsistency, retry", rf.me, server)
				time.Sleep(10 * time.Millisecond)
			}
		}(i, args)
	}
}

// tryCommit tries to increment rf.commitIndex. This function requires the
// caller holds rf.mu throughout the call.
func (rf *Raft) tryCommit() {
	// Subtle: N should start with a large value and decrements. If N starts
	// out small and increases, the leader wouldn't be able to commit once
	// it encounters a entry received from a previous term but replicated in
	// current term (rf.log[N].term == rf.currentTerm always fails!).
	for N := len(rf.log) - 1; N > rf.commitIndex; N-- {
		replicateCount := 0
		for i := 0; i < len(rf.peers); i++ {
			if i == rf.me {
				replicateCount++
				continue
			}

			if rf.matchIndex[i] >= N && rf.log[N].Term == rf.currentTerm {
				DPrintf("%v-th entry is replicated on [%v]", N, i)
				replicateCount++
			} else {
				DPrintf("%v-th entry is NOT replicated on [%v]", N, i)
			}
		}

		if replicateCount > len(rf.peers)/2 {
			rf.commitIndex = N
			DPrintf("[%v] commits %v", rf.me, rf.log[N].Command)
			return
		}
	}
}

/////////////////////// AppendEntries (end) //////////////////////

// resetElectionTimer resets the election timer and generates a new random
// election timeout. It requires the caller holding rf.mu throughout the call.
func (rf *Raft) resetElectionTimer() {
	rf.prevTime = time.Now()
	rf.electionTimeout = time.Duration(RangeInt(500, 800)) * time.Millisecond
}

// convertToFollowerIfOutOfTerm converst to follower if RPC request/response
// term > rf.currentTerm. rf.currentTerm is updated.
// This function requires caller holds rf.mu throughout the call.
// Return value: True means convertted to follower.
func (rf *Raft) convertToFollowerIfOutOfTerm(rpcTerm int) bool {
	if rpcTerm > rf.currentTerm {
		rf.currentTerm = rpcTerm
		rf.votedFor = noPeer
		rf.state = followerState
		return true
	}

	return false
}

func (rf *Raft) applyCommittedRoutine() {
	// Later change this to use condition variable
	for {
		DPrintf("[%v] waiting to be notified", rf.me)
		<-rf.notifyApplyCh
		DPrintf("[%v] Notified to apply", rf.me)

		toApply := false
		var applyMsg ApplyMsg

		rf.mu.Lock()
		for rf.commitIndex > rf.lastApplied {
			rf.lastApplied++
			applyMsg = ApplyMsg{
				CommandValid: true,
				Command:      rf.log[rf.lastApplied].Command,
				CommandIndex: rf.lastApplied,
			}
			toApply = true

			rf.mu.Unlock()
			if toApply {
				DPrintf("[%v] sends %v on applyCh", rf.me, applyMsg.Command)
				rf.applyCh <- applyMsg
			}
			rf.mu.Lock()
		}
		rf.mu.Unlock()
	}
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (2A, 2B, 2C).
	rf.currentTerm = 0
	rf.votedFor = noPeer

	// rf.log stats with a dummy LogEntry.The reason is to follow Figure 2
	// exactly. Figure 2 says rf.log is one-indexed, i.e. the first index is 1.
	// rf.commitIndex = 0 and rf.lastApplied = 0 at boot. However, in Go, slice
	// is zero-indexed. Creating a dummy logEntry for each Raft instance's log
	// at boot works right with the meaning of commitIndex and lastApplied.
	rf.log = []LogEntry{LogEntry{Term: 0}}
	rf.commitIndex = 0
	rf.lastApplied = 0

	rf.state = followerState
	rf.resetElectionTimer()
	rf.votesReceived = 0
	rf.heartbeatInterval = 200 * time.Millisecond
	rf.electionTimerCheckInterval = 150 * time.Millisecond
	rf.heartbeatCheckInterval = 50 * time.Millisecond
	rf.applyCh = applyCh
	rf.notifyApplyCh = make(chan bool)

	// Leader only fields would be initialized later when elected

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// Start goroutines for periodic tasks
	go rf.electionTimoutRoutine()
	go rf.applyCommittedRoutine()

	DPrintf("[%v] Starts off", rf.me)
	return rf
}
