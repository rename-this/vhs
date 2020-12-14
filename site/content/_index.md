---
title: VHS Documentation
weight: 10

cascade:
- type: "docs"
  _target:
    path: "/**"
---

## Introduction to VHS
 
VHS is a versatile tool for network traffic capture. VHS can be run as a command line tool or be deployed into your
Kubernetes cluster as a sidecar, where it can capture traffic to and from your services. Captured traffic
can be analyzed to produce live Prometheus metrics or saved for use in offline analysis, load testing, or whatever
you can imagine.
