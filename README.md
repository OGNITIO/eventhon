# Eventhon

This program is a simple and *unfinished* pluggable subscriber for
Marathon event bus. It captures task related events and send them to
sentry. More flexibility in the event filtering configuration are yet
to come in the near future.

### Commandline Options

Eventhon comes with limited commandline options:

```
Usage of ./eventhon:
-addr string
    IP address and port of eventhon (e.g. localhost:1337)
-sentry_dsn string
    Sentry client key
-sentry_project string
    Sentry project ID
```

### Marathon Configuration

As the program listens and serves POSTed events at `/callbacks`, you
might add the follow options when running Marathon:

```
./bin/start --master ... --event_subscriber http_callback --http_endpoints http://localhost:1337/callbacks 
```

### Resources

- Sentry: http://getsentry.com
- Apache Mesos: http://mesos.apache.org/
- Marathon: https://mesosphere.github.io/marathon/
