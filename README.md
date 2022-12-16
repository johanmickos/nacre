# nacre

Nacre is an open source replica of the seashells.io server.

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
make run
```

## Deployment

### `nacre.dev`

The "official" [nacre.dev](https://nacre.dev) application is deployed to DigitalOcean (DO) with nginx, the Dockerized web application, and redis (also Dockerized) all running on a single droplet. This isn't an ideal setup for high availability, but it's enough for demonstrating the project.

The [nacre.dev](https://nacre.dev) domain name is managed by [NameCheap](https://www.namecheap.com), and for simplicity we leverage NameCheap's DNS nameservers. The DO droplet is assigned a [reserved IP](https://docs.digitalocean.com/products/networking/reserved-ips/) address which is used to configure the DNS records within NameCheap.

[Let's Encrypt](https://letsencrypt.org/) and [certbot](https://certbot.eff.org/) are used to secure the web traffic and manage SSL certificates.

Finally, we leverage GitHub Actions to drive continuous deployment to the DO droplet from the base branch in this repository.


## Dependencies

- make
- golang 1.18.4 or higher
- redis
- docker and docker-compose (optional but helpful)

## Acknowledgements

Kudos to [@anishathalye](https://github.com/anishathalye) for their original work on [seashells](https://seashells.io), and to the developers of [xterm.js](https://xtermjs.org/) for enabling web terminal projects like this one.
