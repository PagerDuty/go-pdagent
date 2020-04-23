# PagerDuty Agent: Eventsapi Package

A minimal client library for PagerDuty's Events API V1 and V2.

The basic API consists of sending an `EventV1` or `EventV2` to `Enqueue` which then automatically determines how and where to send the corresponding event based on version. 

For example usage see:

  - The [eventsapi package](../pkg/eventsapi).
