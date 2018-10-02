= Simple redis relay/proxy with fan-in/fan-out

Another approach to redis replication. Write to multiple systems independently
and return success if more than the configured threshold has passed. For read
pick the highest result (if a version number is present).

