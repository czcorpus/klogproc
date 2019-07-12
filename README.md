# klogproc

An utility for processing/archiving logs generated by applications run by
the Institute of the Czech National Corpus.

In general klogproc reads a custom format from a file or a Redis queue, parses
individual lines and converts them into a target format which is then stored to
ElasticSearch or InfluxDB (both at the same time can be used).

*klogproc* replaces LogStash as a much less resource-hungry alternative.

## Overview

The program supports three operation modes:

### Batch processing of a directory or a file

*klogproc* creates a worklog for the operation so it is possible to process the
source chunk by chunk.

### Batch processing of a Redis queue

Here a worklog is not needed as *klogproc* just takes N items from the queue
and these items are removed from the queue once they are processed.

### Tail-like listening for changes in multiple files

One or more log file listeners can be configured to read newly added lines.
The log files are checked in regular intervals (i.e. the change is not detected
immediately). This is the mode meant to replace LogStash operations.



## Supported apps

* KonText - out of the box (just enable *ucnk_dispatch_hook* plug-in)
* SkE - the application must import a special Python module which is part of this project
* SyD

## Installation

Compile the program and copy the binary to a desired location (e.g. */usr/local/bin*).
Set a cron job to fetch data from logs to ElasticSearch (here every 5 minutes):

```
*/5 * * * * /usr/local/bin/klogproc proclogs /opt/klogproc/ske.json
```
