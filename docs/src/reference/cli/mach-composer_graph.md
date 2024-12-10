## mach-composer graph

Print the execution graph for this project

### Synopsis


Print the execution graph for this project. Note that the output will be in the DOT Language (https://graphviz.org/about/).

This output can be used to generate actual files with the dependency graph. 
A tool like graphviz can be used to make this transformation:
  
  'mach-composer graph -f main.yml | dot -Tpng -o image.png'
	
	

```
mach-composer graph [flags]
```

### Options

```
  -d, --deployment           print the deployment graph instead of the dependency graph
  -f, --file string          YAML file to parse. (default "main.yml")
  -h, --help                 help for graph
      --ignore-version       Skip MACH composer version check
  -o, --output-path string   Outputs path to store the generated files. (default "deployments")
      --var-file string      Use a variable file to parse the configuration with.
  -w, --workers int          The number of workers to use (default 1)
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

