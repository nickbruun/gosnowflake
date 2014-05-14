package gosnowflake

import (
	"testing"
	"time"
)

// Generator creator function.
type generatorCreator func(workerId, datacenterId uint64) (Generator, error)

// Test constants.
func TestConstants(t *testing.T) {
	if sequenceMask != uint64(4095) {
		t.Errorf("Expected sequence mask to be 4095, but it is %d", sequenceMask)
	}
}

// Test generator implementation.
func testGeneratorImplementation(t *testing.T, gName string, creator generatorCreator, epoch uint64) {
	// Creating new generator with out of range IDs returns expected error.
	g, err := creator(32, 0)
	if err != ErrWorkerIdOutOfRange {
		t.Errorf("Expected error from out of range worker ID to be ErrWorkerIdOutOfRange, but it is: %v", err)
	}
	if g != nil {
		t.Errorf("Expected generator returned from call to %s(..) with out of range worker ID to be nil", gName)
	}

	g, err = creator(0, 32)
	if err != ErrDatacenterIdOutOfRange {
		t.Errorf("Expected error from out of range data center ID to be ErrDatacenterIdOutOfRange, but it is: %v", err)
	}
	if g != nil {
		t.Errorf("Expected generator returned from call to %s(..) with out of range data center ID to be nil", gName)
	}

	// Creating new generator with valid IDs returns a generator and no
	// error.
	for _, args := range []struct {
		workerId     uint64
		datacenterId uint64
	}{
		{0, 0},
		{15, 16},
		{31, 31},
	} {
		g, err = creator(args.workerId, args.datacenterId)
		if err != nil {
			t.Errorf("Unexpected error returned from %s(%d, %d, ..): %v", gName, args.workerId, args.datacenterId, err)
		}
		if g == nil {
			t.Errorf("Generator returned from %s(%d, %d, ..) is unexpectedly nil", gName, args.workerId, args.datacenterId)
		}

		workerId := g.WorkerId()
		if workerId != args.workerId {
			t.Errorf("Unexpected worker ID returned from generator, expected %d but got %d", args.workerId, workerId)
		}

		datacenterId := g.DatacenterId()
		if datacenterId != args.datacenterId {
			t.Errorf("Unexpected data center ID returned from generator, expected %d but got %d", args.datacenterId, datacenterId)
		}

		if g.Epoch() != epoch {
			t.Errorf("Unexpected epoch returned from generator, expected %d but got %d", epoch, g.Epoch())
		}
	}

	// A generator returns correct timestamps.
	g, err = creator(5, 24)
	if err != nil {
		t.Fatalf("Unexpected error creating generator: %v", err)
	}

	now := uint64(time.Now().UTC().UnixNano() / 1000000)
	timestamp := g.Timestamp()

	if timestamp < now || timestamp > now+2 {
		t.Errorf("Timestamp returned from generator seems to be adrift from system clock by %d ms", timestamp-now)
	}

	// A generator returns unique IDs - we'll prove this by generating
	// 10^7 of them.
	count := 10000000
	generatedIds := make(map[uint64]struct{}, count)

	for c := 0; c < count; c++ {
		now = uint64(time.Now().UTC().UnixNano()/1000000) - g.Epoch()

		id, err := g.NextId()
		if err != nil {
			t.Fatalf("Getting next ID from generator failed: %v", err)
		}

		if _, exists := generatedIds[id]; exists {
			t.Fatalf("Non-unique ID returned from generator: %d", id)
		}

		if (id>>12)&31 != 5 {
			t.Fatalf("Invalid worker ID in generated ID")
		}

		if (id>>17)&31 != 24 {
			t.Fatalf("Invalid data center ID in generated ID")
		}

		timestamp = id >> 22

		if timestamp < now || timestamp > now+2 {
			t.Fatalf("Timestamp of generated ID seems to be adrift from system clock by %d ms", timestamp-now)
		}

		generatedIds[id] = struct{}{}
	}

	// A generator fails ID generator when clock has drifted backwards.
	gi := g.(*generator)

	gi.lastTimestamp += 1000
	_, err = gi.NextId()

	if err != ErrClockMovingBackwards {
		t.Errorf("Expected ID generation to return ErrClockMovingBackwards, but it got: %v", err)
	}
}

// Test generator with default epoch.
func TestGenerator(t *testing.T) {
	testGeneratorImplementation(t, "NewGenerator", NewGenerator, twitterEpoch)
}

// Test generator with custom epoch.
func TestGeneratorEpoch(t *testing.T) {
	testGeneratorImplementation(t, "NewGeneratorEpoch", func(w, d uint64) (Generator, error) {
		return NewGeneratorEpoch(w, d, 1400027069000)
	}, 1400027069000)
}