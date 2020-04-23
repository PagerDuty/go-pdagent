# PagerDuty Agent: Eventqueue Package

An in-memory event queue for for use on top of the [eventsapi package](../pkg/eventsapi).

Features include:

- Ensuring ordering on a per-routing key basis.
- Handling back-pressure.

For example usage see:

  - The [persistentqueue package](../pkg/persistentqueue).
