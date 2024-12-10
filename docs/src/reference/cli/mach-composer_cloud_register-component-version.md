## mach-composer cloud register-component-version

Register a new version for an existing component

```
mach-composer cloud register-component-version [name] [version] [flags]
```

### Options

```
      --auto                          Add the version commits automatically based on the current branch
      --branch string                 The branch to use for the version. Defaults to the backend default if not set
      --create-component              Will create the component if it does not already exist
      --dry-run                       Dry run
      --git-filter-path stringArray   Filter commits based on given paths
  -h, --help                          help for register-component-version
      --organization string           Organization key
      --project string                Project key
```

### Options inherited from parent commands

```
  -g, --github          Whether logs should be decorated with github-specific formatting
      --output string   The output type. One of: console, json (default "console")
  -q, --quiet           Quiet output. This is equal to setting log levels to error and higher
  -v, --verbose         Verbose output. This is equal to setting log levels to debug and higher
```

### SEE ALSO

* [mach-composer cloud](mach-composer_cloud.md)	 - Manage your Mach Composer Cloud

