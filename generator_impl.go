package gosnowflake

import (
	"sync"
	"time"
)

const (
	workerIdBits = 5
	maxWorkerId  = (uint64(1) << workerIdBits) - uint64(1)

	datacenterIdBits = 5
	maxDatacenterId  = (uint64(1) << datacenterIdBits) - uint64(1)

	twitterEpoch = uint64(1288834974657)

	sequenceBits = 12
	sequenceMask = uint64(-1 ^ (-1 << sequenceBits))

	workerIdShift     = sequenceBits
	datacenterIdShift = sequenceBits + workerIdBits
	timestampShift    = sequenceBits + workerIdBits + datacenterIdBits
)

// Unique ID generator implementation.
type generator struct {
	workerId      uint64
	datacenterId  uint64
	epoch         uint64
	lastTimestamp uint64
	sequence      uint64
	mutex         sync.Mutex
}

// New unique ID generator.
//
// Worker ID and data center ID must be in the range [0 ; 31].
func NewGenerator(workerId, datacenterId uint64) (Generator, error) {
	// Check worker and data center IDs for sanity.
	if workerId > maxWorkerId {
		return nil, ErrWorkerIdOutOfRange
	}

	if datacenterId > maxDatacenterId {
		return nil, ErrDatacenterIdOutOfRange
	}

	return &generator{
		workerId:      workerId,
		datacenterId:  datacenterId,
		epoch:         twitterEpoch,
		lastTimestamp: 0,
	}, nil
}

// New unique ID generator with custom epoch.
//
// Worker ID and data center ID must be in the range [0 ; 31]. The supplied
// epoch must be the number of milliseconds since the UNIX UTC epoch.
func NewGeneratorEpoch(workerId, datacenterId, epoch uint64) (Generator, error) {
	// Check worker and data center IDs for sanity.
	if workerId > maxWorkerId {
		return nil, ErrWorkerIdOutOfRange
	}

	if datacenterId > maxDatacenterId {
		return nil, ErrDatacenterIdOutOfRange
	}

	return &generator{
		workerId:      workerId,
		datacenterId:  datacenterId,
		epoch:         epoch,
		lastTimestamp: 0,
	}, nil
}

func (g *generator) WorkerId() uint64 {
	return g.workerId
}

func (g *generator) DatacenterId() uint64 {
	return g.datacenterId
}

func (g *generator) Timestamp() uint64 {
	return uint64(time.Now().UTC().UnixNano() / 1000000)
}

func (g *generator) Epoch() uint64 {
	return g.epoch
}

func (g *generator) NextId() (id uint64, err error) {
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
