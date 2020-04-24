# PagerDuty Agent: Persistentqueue Package

Provides on-disk persistence for an underlying event queue from the [eventqueue package](../pkg/eventqueue).

This persistence is primarily leveraged during startup to ensure that any pending events from a previous shutdown are still processed and to provide queue analysis.

For example usage see:

  - The [server package](../pkg/server)'s Queue interface.
