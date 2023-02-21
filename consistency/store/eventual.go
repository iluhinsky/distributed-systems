package store

import (
	"context"

	"github.com/iluhinsky/distributed-systems/network"
	"github.com/iluhinsky/distributed-systems/util"
)

type eventualStoreUpdate struct {
	key   int64
	value util.TimestampedValue
}

type EventualStore struct {
	rid        int64
	localClock int64
	store      map[int64]util.TimestampedValue
	replicas   map[int64]network.Link
}

func NewEventualStore(rid int64) *EventualStore {
	return &EventualStore{
		rid:      rid,
		store:    map[int64]util.TimestampedValue{},
		replicas: map[int64]network.Link{},
	}
}

func (s *EventualStore) Write(key int64, value int64) bool {
	s.localClock++

	var tValue = util.TimestampedValue{
		Val: value,
		Ts: util.Timestamp{
			Number: s.localClock,
			Rid:    s.rid,
		},
	}

	s.store[key] = tValue

	for _, r := range s.replicas {
		r.Send(context.Background(), eventualStoreUpdate{
			key:   key,
			value: tValue,
		})
	}

	return true
}

func (s *EventualStore) Read(key int64) int64 {
	if row, ok := s.store[key]; ok {
		return row.Val
	}

	return 0
}

func (s *EventualStore) Introduce(rid int64, link network.Link) {
	if rid != 0 && link != nil {
		s.replicas[rid] = link
	}
}

func (s *EventualStore) Receive(rid int64, msg interface{}) interface{} {
	if update, ok := msg.(eventualStoreUpdate); ok {
		if row, ok := s.store[update.key]; !ok || row.Ts.Less(update.value.Ts) {
			s.store[update.key] = update.value
		}

		if s.localClock < update.value.Ts.Number {
			s.localClock = update.value.Ts.Number
		}
	}
	return nil
}
