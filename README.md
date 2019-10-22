# Simple Exquisite Webserver
[![Build Status](https://travis-ci.org/opentelekomcloud-infra/simple-exquisite-webserver.svg?branch=master)](https://travis-ci.org/opentelekomcloud-infra/simple-exquisite-webserver)

This is single-purpose web server having following endpoints:

`/` — always returns http code `200`, can be used to validate if server is up and running

`/entities` — for listing all existing entities

`/entity`, `/entity/<uuid>` — for creating and retrieving existing entities

Every server response contains `Server` header with value equal to host name 

Server can use PostgreSQL database

Server configuration is done with configuration yaml file:
```yaml
debug: false
server_port: 9069

postgres: # Required if debug is false
  db_url: 'localhost:46063'
  database: 'users'
  username: 'admin'
  password: 'Qwertyui!2019'
  
  initial_data:  # Records generated at app initialization (skipped if missing)
    count: 10000  # Number of created records
    size: 20000  # Size of each record created
```

Default location of configuration file is `/etc/too-simple/config.yml`,
this can be changed using `--config` argument 

Debug mode is switched using `--debug` argument or setting `debug: true` in configuration.
When debug is on, no database will be used.

You can get application version using `--version` argument
