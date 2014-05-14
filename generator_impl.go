package gosnowflake

import (
	"sync"
	"time"
)

const (
	workerIdBits = 5
	maxWorkerId  = (1 << workerIdBits) - 1

	datacenterIdBits = 5
	maxDatacenterId  = (1 << datacenterIdBits) - 1

	sequenceBits = 12
	sequenceMask = int64(-1 ^ (-1 << sequenceBits))

	workerIdShift     = sequenceBits
	datacenterIdShift = sequenceBits + workerIdBits
	timestampShift    = sequenceBits + workerIdBits + datacenterIdBits
)

// Unique ID generator implementation.
type generator struct {
	workerId      int64
	datacenterId  int64
	epoch         int64
	lastTimestamp int64
	sequence      int64
	mutex         sync.Mutex
}

// New unique ID generator.
//
// Worker ID and data center ID must be in the range [0 ; 31].
func NewGenerator(workerId, datacenterId int) (Generator, error) {
	return NewGeneratorEpoch(workerId, datacenterId, TwitterEpoch)
}

// New unique ID generator with custom epoch.
//
// Worker ID and data center ID must be in the range [0 ; 31]. The supplied
// epoch must be the number of milliseconds since the UNIX UTC epoch.
func NewGeneratorEpoch(workerId, datacenterId int, epoch int64) (Generator, error) {
	// Check worker and data center IDs for sanity.
	if workerId < 0 || workerId > maxWorkerId {
		return nil, ErrWorkerIdOutOfRange
	}

	if datacenterId < 0 || datacenterId > maxDatacenterId {
		return nil, ErrDatacenterIdOutOfRange
	}

	return &generator{
		workerId:      int64(workerId),
		datacenterId:  int64(datacenterId),
		epoch:         epoch,
		lastTimestamp: 0,
	}, nil
}

func (g *generator) WorkerId() int {
	return int(g.workerId)
}

func (g *generator) DatacenterId() int {
	return int(g.datacenterId)
}

func (g *generator) Timestamp() int64 {
	return int64(time.Now().UTC().UnixNano() / 1000000)
}

func (g *generator) Epoch() int64 {
	return g.epoch
}

func (g *generator) NextId() (id int64, err error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Get the current timestamp.
	timestamp := g.Timestamp()

	// Make sure the clock is not moving backwards.
	if g.lastTimestamp > timestamp {
		err = ErrClockMovingBackwards
		return
	}

	// Update the sequence counter.
	if timestamp == g.lastTimestamp {
		g.sequence = (g.sequence + 1) & sequenceMask

		if g.sequence == 0 {
			// Spin until we get the next millisecond. Not super efficient,
			// but given that explicitly sleeping is not always particularly
			// precise, this keeps latencies down.
			for timestamp == g.lastTimestamp {
				timestamp = g.Timestamp()
			}

			g.lastTimestamp = timestamp
		}
	} else {
		g.sequence = 0
		g.lastTimestamp = timestamp
	}

	// Generate the ID.
	id = ((timestamp - g.epoch) << timestampShift) | (g.datacenterId << datacenterIdShift) | (g.workerId << workerIdShift) | g.sequence

	return
}
