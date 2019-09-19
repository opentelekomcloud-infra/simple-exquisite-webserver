# Simple Exquisite Webserver
[![Build Status](https://travis-ci.org/outcatcher/simple-exquisite-webserver.svg?branch=master)](https://travis-ci.org/outcatcher/simple-exquisite-webserver)

This is single-purpose web server having following endpoints:

`/` — always returns http code `200`, can be used to validate if server is up and running

`/entities` — for listing all existing entities

`/entity`, `/entity/<uuid>` — for creating and retrieving existing entities

Every server response contains `Server` header with value equal to host name 

Server can use either debug sqlite DB or PostgreSQL database

Server configuration is done with configuration yaml file:
```yaml
debug: true
server_port: 5449 

pg_bd_url: 'localhost:9999'
pg_database: 'users'
pg_username: 'admin'
pg_password: 'qwertyui!'
```

Default location of configuration file is `/etc/too-simple/config.yml`,
this can be changed using `--config` argument 

Debug mode is switched using `--debug` argument and enabled by default
