# Deployment

## `nacre.dev`

The "official" [nacre.dev](https://nacre.dev) application is deployed to DigitalOcean (DO) with nginx, the Dockerized web application, and redis (also Dockerized) all running on a single droplet. This isn't an ideal setup for high availability, but it's enough for demonstrating the project.

The [nacre.dev](https://nacre.dev) domain name is managed by [NameCheap](https://www.namecheap.com), and for simplicity we leverage NameCheap's DNS nameservers. The DO droplet is assigned a [reserved IP](https://docs.digitalocean.com/products/networking/reserved-ips/) address which is used to configure the DNS records within NameCheap.

[Let's Encrypt](https://letsencrypt.org/) and [certbot](https://certbot.eff.org/) are used to secure the web traffic and manage SSL certificates.

Finally, we leverage GitHub Actions to drive continuous deployment to the DO droplet from the base branch in this repository.


