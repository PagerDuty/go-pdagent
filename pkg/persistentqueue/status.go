package persistentqueue

type StatusItem struct {
	RoutingKey string `json:"routing_key"`
	Pending    int    `json:"pending"`
	Success    int    `json:"success"`
	Error      int    `json:"error"`
}

// Returns aggregate stats per routing key for pending and enqueued events.
func (q *PersistentQueue) Status(routingKey string) ([]StatusItem, error) {
	var err error
	var events []Event
	agg := map[string]*StatusItem{}

	if routingKey == "" {
		err = q.Events.All(&events)
	} else {
		err = q.Events.Find("RoutingKey", routingKey, &events)
	}
	if err != nil {
		return nil, err
	}

	for _, event := range events {
		item, ok := agg[event.RoutingKey]
		if !ok {
			item = &StatusItem{RoutingKey: event.RoutingKey}
		}

		switch event.Status {
		case StatusPending:
			item.Pending++
		case StatusSuccess:
			item.Success++
		case StatusError:
			item.Error++
		}

		agg[event.RoutingKey] = item
	}

	var items []StatusItem
	for _, v := range agg {
		items = append(items, *v)
	}

	return items, nil
}
