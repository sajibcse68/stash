---
title: Stash Recover | Stash
description: Stash Recover
menu:
  product_stash_master:
    identifier: reference-stash-recover
    name: Stash Recover
    parent: reference
    weight: 30
product_name: stash
left_menu: product_stash_master
section_menu_id: reference
---
## stash recover

Recover restic backup

### Synopsis


Recover restic backup

```
stash recover [flags]
```

### Options

```
  -h, --help                   help for recover
      --kubeconfig string      Path to kubeconfig file with authorization information (the master location is set by the master flag).
      --master string          The address of the Kubernetes API server (overrides any value in kubeconfig)
      --recovery-name string   Name of the Recovery CRD.
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
      --analytics                        Send analytical events to Google Analytics (default true)
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO
* [stash](stash.md)	 - Stash by AppsCode - Backup your Kubernetes Volumes

