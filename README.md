# Klogproc

*Klogproc* is a utility for processing/archiving logs generated by applications
run by the Institute of the Czech National Corpus.

In general *Klogproc* reads an application-specific log record format from a file
or a Redis queue, parses individual lines and converts them into a target format
which is then stored to ElasticSearch or InfluxDB (both at the same time can be used).

*Klogproc* replaces LogStash as a less resource-hungry and runtime environment demanding
alternative. All the processing (reading, writing, handling multiple files) is performed
concurrently which makes it quite fast.

## Overview

### Supported applications

| Name    | config code |
|---------|-------------|
| Calc    | calc        |
| KonText | kontext     |
| Kwords  | kwords      |
| Morfio  | morfio      |
| SkE     | ske         |
| SyD     | syd         |
| Treq    | treq        |
| WaG     | wag         |

The program supports three operation modes - *batch*, *tail*, *redis*

### Batch processing of a directory or a file

For non-regular imports e.g. when migrating older data, *batch* mode allows
importing of multiple files from a single directory. The contents of the directory
can be even changed over time by adding **newer** log records and *klogproc* will
be able to import only new items as it keeps a worklog with the newest record
currently processed.

### Batch processing of a Redis queue

Note: On the application side, this is currently supported only in KonText
and SkE (with added special Python module *scripts/redislog.py* which is part of
the *klogproc* project).

In this case, an application writes its log to a Redis queue (*list* type) and
*klogproc* regularly takes N items from the queue (items are removed from there),
transforms them and stores them to specified destinations.

### Tail-like listening for changes in multiple files

This is the mode which replaces CNC's LogStash solution and it is a typical
mode to use. One or more log file listeners can be configured to read newly
added lines. The log files are checked in regular intervals (i.e. the change is
not detected immediately). *Klogproc* remembers current inode and current seek position
for watched files so it should be able to continue after outages etc. (as long as
the log files are not overwritten  in the meantime due to log rotation).


## Installation

[Install](https://golang.org/doc/install) the *Go* language if it is not already
available on your system.

Clone the *klogproc* project:

`git clone https://github.com/czcorpus/klogproc.git`

Build the project:

`go build`

Copy the binary somewhere:

`sudo cp klogproc /usr/local/bin`

Create a config file (e.g. in /usr/local/etc/klogproc.json):

```
{
    "logPath": "/var/log/klogproc/klogproc.log",
    "logTail": {
	"intervalSecs": 15,
      "worklogPath": "/var/opt/klogproc/worklog-tail.log",
      "files": [
        {"path": "/var/log/ucnk/syd.log", "appType": "syd"},
        {"path": "/var/log/treq/treq.log", "appType": "treq"},
	    {"path": "/var/log/ucnk/morfio.log", "appType": "morfio"},
	    {"path": "/var/log/ucnk/kwords.log", "appType": "kwords"}
      ]
    },
    "elasticSearch": {
	  "majorVersion": 5,
      "server": "http://elastic:9200",
      "index": "log_archive",
      "pushChunkSize": 500,
      "scrollTtl": "3m",
      "reqTimeoutSecs": 10
    },
    "geoIPDbPath": "/var/opt/klogproc/GeoLite2-City.mmdb",
    "localTimezone": "+01:00",
    "anonymousUsers": [0, 1, 2]
}
```
(do not forget to create directory for logging, worklog and also
download and save GeoLite2-City database).

Configure systemd (/etc/systemd/system/klogproc.service):

```
[Unit]
Description=A custom agent for collecting UCNK apps logs
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/klogproc tail /usr/local/etc/klogproc.json
User=klogproc
Group=klogproc

[Install]
WantedBy=multi-user.target
```

Reload systemd config:

`systemctl daemon-reload`

Start the service:

`systemctl start klogproc`


## ElasticSearch compatibility notes

Because ElasticSearch underwent some backward incompatible changes between versions 5.x.x and 6.x.x ,
the configuration contains the *majorVersion* key which specifies how *klogproc* stores the data.

### ElasticSearch 5

This version supports multiple data types ("mappings") per index which was also
the default approach how CNC applications were stored - single index, multiple document
types (each per application). In this case, the configuration directive *elasticSearch.index*
specifies directly the index name *klogproc* works with. Individual document types
can be distinguished either via ES internal *_type* property or via normal property *type*
which is created by *klogproc*.

### ElasticSearch 6

Here, multiple data mappings per index are being removed. *Klogproc* in this case
uses its *elasticSearch.index* key as a prefix for index name created for an individual
application. E.g. index = "log_archive" with configured "treq" and "morfio" apps expects
you to have two indices: *log_archive_treq* and *log_archive_morfio". Please note
that *klogproc* does not create the indices for you. The property *type* is still present
in documents.


## InfluxDB notes

InfluxDB is a pure time-based database with focus on processing (mostly numerical) measurements.
Compared with ElasticSearch, its search capabilities are limited so it cannot be understood
as a possible replacement of ElasticSearch. With configured InfluxDB output, *klogproc* can be used
to add some more useful data to existing measurements generated by other applications (typically
[Telegraf](https://github.com/influxdata/telegraf), [Netdata](https://github.com/netdata/netdata)).

Please note that the InfluxDB output is not currently used in production.
