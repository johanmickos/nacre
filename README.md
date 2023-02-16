# [nacre](https://nacre.dev)

[Nacre](https://nacre.dev) enables you to stream commandline output to the web and view the realtime output in your browser.

It is an open source replica of the [seashells.io](https://seashells.io/) server.

## Examples

```bash
htop | nacre.dev 1337
```

## What's in a name?

Nacre is another word for mother-of-pearl, the inside of some seashells.

## Running

Nacre can either run natively from your commandline or as a Dockerized application. The application requires Redis to be up and running in order to serve data feeds.

```
# To immediately run the server with default settings:
make run

# Alternatively, build the server and run it explicitly:
make build
./out/bin/nacre-server

# Or leverage docker-compose to run both Nacre and Redis:
make dockerbuild
make dockerrun
```

## Configuration

See the [sample .env file](.env.sample) for configurable parameters.

## Deployment
See the [deployment README](deployment/README.md) for details on how https://nacre.dev is deployed.

## Dependencies

- make
- golang 1.18.4 or higher
- redis
- docker and docker-compose (optional but helpful)

## Acknowledgements

Kudos to [@anishathalye](https://github.com/anishathalye) for their original work on [seashells](https://seashells.io), and to the developers of [xterm.js](https://xtermjs.org/) for enabling web terminal projects like this one.
