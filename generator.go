package gosnowflake

// Unique ID generator.
type Generator interface {
	// Worker ID.
	WorkerId() int

	// Datacenter ID.
	DatacenterId() int

	// Current timestamp.
	//
	// Milliseconds since the UNIX epoch.
	Timestamp() int64

	// Generator epoch.
	//
	// Milliseconds since the UNIX epoch that is considered the generator's
	// epoch.
	Epoch() int64

	// Generate next ID.
	NextId() (int64, error)
}
