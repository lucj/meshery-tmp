---
layout: guide
title: Multiple Adapters
permalink: guides/multiple-adapters
type: guide
---
# Meshery Adapters: Advanced Configuration

## Modifying the default adapter deployment configuration
The number of adapters, type of adapters, where they are deployed, how they are named and what port they are exposed on are all configurable deployment options. To modify the default configuration, find `~/.meshery/meshery.yaml` on your system. `~/.meshery/meshery.yaml` is a Docker Compose file.

### Configuration: Running fewer Meshery adapters
In the `~/.meshery/meshery.yaml` configuration file, remove the entry(ies) of the adapter(s) you are removing from your deployment.

### Configuration: Running more than one instance of the same Meshery adapter

The default configuration of a Meshery deployment includes one instance of each of the Meshery adapters (that have reached a stable version status). You may choose to run multiple instances of the same type of Meshery adapter; e.g. two instances of the `meshery-istio` adapter. To do so, modify `~/.meshery/meshery.yaml` to include multiple copies of the given adapter.