package metcd

import (
	"context"

	"github.com/etcd-io/etcd/raft"
	"github.com/etcd-io/etcd/raft/raftpb"
	"github.com/etcd-io/etcd/wal"
	"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp"
)

// A key-value stream backed by raft
type raftNode struct {
	proposeC    <-chan string            // proposed messages(k,v)
	confChangeC <-chan raftpb.ConfChange // proposed cluster config change
	commitC     <-chan *string           // entries to committed to log (k,v)
	errorC      <-chan error             // errors from raft session
	id          int                      // client ID for raft session
	node        *raftNode
	raftStorage *raft.MemoryStorage
	wal         *wal.WAL
	transport   *rafthttp.Transport
}

// var transport = &rafthttp.Transport{
// 	Logger:       zap.NewExample(),
// 	ID:           types.ID(rc.id),
// 	ClusterId:    0x1000,
// 	Raft:         rc,
// 	ServerStates: stats.NewLeaderStas(strconv.Itoa(rc.id)),
// 	ErrorC:       make(chan error),
// }

func (rc *raftNode) serveChannels() {
	// send proposals over raft
	go func() {
		confChangeCount := uint64(0)
		for rc.proposeC != nil && rc.confChangeC != nil {
			select {
			case prop, ok := <-rc.proposeC:
				if !ok {
					rc.proposeC = nil
				} else {
					// blocks until accepted by raft state machine
					rc.node.Propose(context.TODO(), []byte(prop))
				}
			case cc, ok := <-rc.confChangeC:
				if !ok {
					rc.confChangeC = nil
				} else {
					confChangeCount++
					cc.ID = confChangeCount
					rc.node.ProposeConfChange(context.TODO(), cc)
				}
			}

		}
	}()

	// event loop on raft state machine updates
	for {
		select {
		case <-ticker.C:
			rc.node.Tick()
			// store raft entries to wal, then publish over commit channel
		case rd := <-rc.node.Ready:
			rc.wal.Save(rd.HardState, rd.Entries)
			if !raft.IsEmptySnap(rd.Snapshot) {
				rc.saveSnap(rd.Snapshot)
				rc.raftStorage.ApplySnapshot(rd.Snapshot)
				rc.publishSnapshot(rd.Snapshot)
			}
			rc.raftStorage.Append(rd.Entries)
			rc.transport.Send(rd.Message)
			if ok := rc.publishEntries(rc.entriesToApply(rd.CommitedEntries)); !ok {
				rc.stop()
				return
			}
			rc.maybeTriggerSnapshot()
			rc.node.Advance()
		}
	}
}
