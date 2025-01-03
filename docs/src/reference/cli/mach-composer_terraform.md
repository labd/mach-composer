## mach-composer terraform

Execute terraform commands directly

```
mach-composer terraform [flags]
```

### Options

```
  -f, --file string               YAML file to parse. (default "main.yml")
  -h, --help                      help for terraform
      --ignore-change-detection   Ignore change detection to run even if the components are considered up to date. Per default the proxy will ignore change detection (default true)
      --ignore-version            Skip MACH composer version check
  -o, --output-path string        Outputs path to store the generated files. (default "deployments")
      --var-file string           Use a variable file to parse the configuration with.
  -w, --workers int               The number of workers to use (default 1)
```

### Options inherited from parent commands

```
  -g, --github          Whether logs should be decorated with github-specific formatting
      --output string   The output type. One of: console, json (default "console")
  -q, --quiet           Quiet output. This is equal to setting log levels to error and higher
  -v, --verbose         Verbose output. This is equal to setting log levels to debug and higher
```

### SEE ALSO

* [mach-composer](mach-composer.md)	 - MACH composer is an orchestration tool for modern MACH ecosystems

