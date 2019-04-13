package counter

import "sync"

type MemoryBookings struct {
	count    int
	bookings map[string]struct{}
	lock     sync.RWMutex
}

func (m *MemoryBookings) IncrementBookingCounter(bookingUUID string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.bookings == nil {
		m.bookings = map[string]struct{}{}
	}

	if _, ok := m.bookings[bookingUUID]; ok {
		// deduplicated
		return nil
	}
	m.bookings[bookingUUID] = struct{}{}

	m.count += 1
	return nil
}

func (m *MemoryBookings) Count() int {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.count
}
