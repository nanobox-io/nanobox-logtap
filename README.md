# Logtap

Logtap is an embeddable log aggregation, storage, and publishing service.

## Memory Usage

## Drain

A `logtap.Drain` is a simple endpoint that accepts logs that are sent through logtap. Multiple drains can be created and added to logvac. A drain can represent logs that are streamed to stdout, a file, a tcp socket, or anything that can be wrapped to accept `logtap.Message` structs.

## Collector

A Collector can be anything that calls `logtap.Publish()`. There are no limitations placed on Collectors.

## Archive

A `logtap.Archive` is a simple endpoint for retreiving `logtap.Message` structs.

## Api :

The logtap api

### Notes: