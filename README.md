# rancher-autocrypt

Automatically generate Let's Encrypt certificates in Rancher and add them to your certificate storage in rancher and update the load balancer configuration.

## Requirements

Needs at least Rancher 1.6.6 because of
https://github.com/rancher/rancher/issues/9118
https://github.com/rancher/rancher/issues/7892

## Installation

### Install the rancher stack

1. Go to Admin ==> Settings
2. Add a new catalog with URL = https://github.com/KeeTraxx/rancher-autocrypt.git
3. Add the rancher autocrypt stack and specify your contact email.

### Setup your load balancer

1. Add a selector rule with ch.compile.letsencrypt=true
2. Add a service rule with port=80 path=/.well-known/acme-challenge ==> rancher-autocrypt

### Setup a service

Simple webserver example

docker-compose.yml:

    version: '2'
    services:
    web:
        image: nginx
        stdin_open: true
        tty: true
        labels:
        io.rancher.container.pull_image: always
        ch.compile.letsencrypt: true


rancher-compose.yml
    version: '2'
    services:
    web:
        scale: 1
        start_on_create: true
        lb_config:
            port_rules:
            - target_port: 80
            hostname: myserver.com
            - target_port: 80
            hostname: www.myserver.com


`myserver.com` and `www.myserver.com` must point to your load balancer!
