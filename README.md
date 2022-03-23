# Guestbook Example Application

This repository contains the source code to a demo application based on the [Kubernetes guestbook application](https://github.com/kubernetes/examples/tree/master/guestbook-go) for the purposes of demonstrating a GitOps based CI/CD pipeline. The git repository containing the Kubernetes deployment manifests is located at a separate repository: https://github.com/akuity/guestbook-deploy. 

A code change to this repository will cause:
1. A new [ghcr.io/akuity/guestbook](https://github.com/akuity/guestbook/pkgs/container/guestbook) image to be published with a unique image tag that incorporates the commit SHA into the image tag (e.g. `ghcr.io/akuity/guestbook:00003-f32b7f8`).
1. An automated git commit to be pushed to the staging manifests with the new image, resulting in an automated deploy to the staging environment
1. A PR to be created against the prod manifests, for manual approval and deploy

## Screenshot

![Guestbook](guestbook-page.png)

