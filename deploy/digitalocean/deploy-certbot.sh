
#!/usr/bin/env bash
# FIXME: This can only be done once nginx is up and running with a "Hello, world" page
# certbot: acquire certificates and set up SSL for nginx
# NOTE: This only needs to be done on the Internet-facing load balancer/reverse proxy,
#       since we're OK with HTTP on the path between nginx and the application.
DEV_EMAIL="johan.mickos@gmail.com"
# NOTE: Requires nginx to be up and serving with correct configuration
# of "server_name" to match the domains provided (nacre.dev and www.nacre.dev)
sudo certbot -n --agree-tos -m $DEV_EMAIL --nginx -d nacre.dev -d www.nacre.dev
