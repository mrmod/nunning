# BuildTool

A tool which uses filesystem layers as the Cached Action Store (CAS) / Content-Addressable-Storage system for built items using layerfs.

# FuzzySearch

Term-frequency inverted tree [implementation](github.com/mrmod/fuzzysearch).

A sample using QuickWit instead [implementation](fuzzysearch).

# Homewatch

## Agent

The [HomewatchAgent](github.com/mrmod/homewatch/) proxies DAV-encoded H.265 videos it gets for Lorex cameras to S3 so authorized users can view them. It's (poorly) evented on Syslog messages from the SFTP service which receives the DAV-encoded video uploads.

This is a subtree. Adding a remote and pulling is a way to keep up to date on it.

```
git remote add homewatch-agent https://github.com/mrmod/homewatch-agent.git

git subtree pull --prefix homewatch-agent homewatch-agent main
```

# Terrastate
Shows terraform changes in web ui

