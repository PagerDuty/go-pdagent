package persistentqueue

// Retries events that are in an error state, either for an routing key or
// for all events in error if none is provided.
func (q *PersistentQueue) Retry(routingKey string) (int, error) {
	var err error
	var events []Event

	err = q.Events.Find("Status", StatusError, &events)
	if err != nil {
		return 0, err
	}

	for _, event := range events {
		if routingKey == "" || event.RoutingKey == routingKey {
			q.processEvent(&event)
		}
	}

	return len(events), nil
}
