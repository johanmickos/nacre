#!/bin/bash
set -x

sudo adduser --disabled-password --gecos "" nacre-admin
sudo passwd -d nacre-admin
sudo usermod -aG docker nacre-admin
sudo usermod -aG sudo nacre-admin
sudo su - nacre-admin

sudo apt-get -y update
sudo apt-get -y install apt-transport-https ca-certificates curl software-properties-common

#### Install service dependencies
# nginx: reverse proxying to nacre web application
sudo apt-get -y install nginx

# certbot: SSL certificate management
sudo snap install core
sudo snap refresh core
sudo apt-get -y remove certbot
sudo snap install --classic certbot
sudo ln -s /snap/bin/certbot /usr/bin/certbot

# docker
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get -y update
apt-cache policy docker-ce
sudo apt-get -y install docker-ce

# docker-compose
mkdir -p ~/.docker/cli-plugins
DOCKER_COMPOSE_VERSION="v2.14.1"
sudo curl -L "https://github.com/docker/compose/releases/download/$DOCKER_COMPOSE_VERSION/docker-compose-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)" -o ~/.docker/cli-plugins/docker-compose
chmod +x ~/.docker/cli-plugins/docker-compose

# TODO DigitalOcean metrics agent
sudo curl -sSL https://repos.insights.digitalocean.com/install.sh | sudo bash

#### Uncomplicated Firewall (UFW) setup
# For details:
#   https://www.digitalocean.com/community/tutorials/ufw-essentials-common-firewall-rules-and-commands
sudo ufw app list        # See registered firewall applications (should see OpenSSH's application profile)
sudo ufw allow OpenSSH   # Allowlist OpenSSH profile before enabling firewall
sudo ufw enable
sudo ufw status          # See current status (should see ALLOW for both OpenSSH and OpenSSH v6)

sudo ufw allow 'Nginx HTTP'
sudo ufw allow 'Nginx HTTPS'
sudo ufw allow 1337/tcp

#### Nacre configuration
sudo mkdir /app && sudo chown nacre-admin /app
sudo mkdir /etc/nginx/stream.d

# FIXME: These should be part of continuous deployment
# Using HTTPS because SSH requires password-protected SSH key
git clone https://github.com/jarlopez/nacre.git /app/nacre

/app/nacre/deploy/digitalocean/deploy-nginx.sh
/app/nacre/deploy/digitalocean/deploy-systemd.sh

sudo nginx -s reload

# TODO: Ensure that a built Docker image is available for running
sudo systemctl enable nacre
sudo service nacre start

#### Final configuration
/app/nacre/deploy/digitalocean/deploy-certbot.sh

echo "Nacre server is ready for application deployment"
set +x