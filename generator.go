package gosnowflake

// Unique ID generator.
type Generator interface {
	// Worker ID.
	WorkerId() uint64

	// Datacenter ID.
	DatacenterId() uint64

	// Current timestamp.
	//
	// Milliseconds since the UNIX epoch.
	Timestamp() uint64

	// Generator epoch.
	//
	// Milliseconds since the UNIX epoch that is considered the generator's
	// epoch.
	Epoch() uint64

	// Generate next ID.
	NextId() (uint64, error)
}
