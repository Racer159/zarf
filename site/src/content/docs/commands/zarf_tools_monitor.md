---
title: zarf tools monitor
description: Zarf CLI command reference for <code>zarf tools monitor</code>.
tableOfContents: false
---

<!-- Page generated by Zarf; DO NOT EDIT -->

## zarf tools monitor

Launches a terminal UI to monitor the connected cluster using K9s.

```
zarf tools monitor [flags]
```

### Options

```
  -A, --all-namespaces                 Launch K9s in all namespaces
      --as string                      Username to impersonate for the operation
      --as-group stringArray           Group to impersonate for the operation
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
  -c, --command string                 Overrides the default resource to load when the application launches
      --context string                 The name of the kubeconfig context to use
      --crumbsless                     Turn K9s crumbs off
      --headless                       Turn K9s header off
  -h, --help                           help for monitor
      --insecure-skip-tls-verify       If true, the server's caCertFile will not be checked for validity
      --kubeconfig string              Path to the kubeconfig file to use for CLI requests
      --logFile string                 Specify the log file
  -l, --logLevel string                Specify a log level (error, warn, info, debug)
      --logoless                       Turn K9s logo off
  -n, --namespace string               If present, the namespace scope for this CLI request
      --readonly                       Sets readOnly mode by overriding readOnly configuration setting
  -r, --refresh int                    Specify the default refresh rate as an integer (sec) (default 2)
      --request-timeout string         The length of time to wait before giving up on a single server request
      --screen-dump-dir string         Sets a path to a dir for a screen dumps
      --splashless                     Turn K9s splash screen off
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
      --write                          Sets write mode by overriding the readOnly configuration setting
```

### Options inherited from parent commands

```
      --plain-http   Force the connections over HTTP instead of HTTPS. This flag should only be used if you have a specific reason and accept the reduced security posture.
```

### SEE ALSO

* [zarf tools](/commands/zarf_tools/)	 - Collection of additional tools to make airgap easier

