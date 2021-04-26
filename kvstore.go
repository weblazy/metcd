package metcd

type Kvstore interface {
	// Lookup get key value
	Lookup(key string) (string, bool)
	// Propose kv request into raft state machine
	Propose(k, v string)
	// ReadCommits consume entry from raft state machine into KvStore map until error
	ReadCommits(commitC <-chan *string, errorC <-chan error)
	// SnapShot return KvStore snapshot
	SnapShot() ([]byte, error)
	// RecoverFromSnapshot recover from snapshot
	RecoverFromSnapshot(snapshot []byte) error
	// Close close backend databases
	Close() error
}
