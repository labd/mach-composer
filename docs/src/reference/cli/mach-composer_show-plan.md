## mach-composer show-plan

Show the planned configuration.

```
mach-composer show-plan [flags]
```

### Options

```
  -f, --file string          YAML file to parse. (default "main.yml")
      --force                Force the show plan to run even if the components are considered up to date
  -h, --help                 help for show-plan
      --ignore-version       Skip MACH composer version check
      --no-color             Disable color output
      --output-path string   Outputs path to store the generated files. (default "deployments")
      --reuse                Suppress a terraform init for improved speed (not recommended for production usage)
  -s, --site string          Site to parse. If not set parse all sites.
      --var-file string      Use a variable file to parse the configuration with.
  -w, --workers int          The number of workers to use (default 1)
```

### Options inherited from parent commands

```
      --output string   The output type. One of: console, json (default "console")
      --verbose         Verbose output.
```

### SEE ALSO

* [mach-composer](mach-composer.md)	 - MACH composer is an orchestration tool for modern MACH ecosystems

