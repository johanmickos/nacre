#!/usr/bin/env bash
sudo cp /app/nacre/deploy/config/nginx/nginx.conf /etc/nginx/nginx.conf
sudo cp /app/nacre/deploy/config/nginx/www/* /var/www/html/
sudo cp /app/nacre/deploy/config/nginx/nacre.dev /etc/nginx/sites-available
sudo cp /app/nacre/deploy/config/nginx/nacre-data-stream.conf /etc/nginx/stream.d
sudo ln --symbolic --force /etc/nginx/sites-available/nacre.dev /etc/nginx/sites-enabled/nacre.dev
sudo cp /app/nacre/deploy/config/systemd/nacre.service /etc/systemd/system
