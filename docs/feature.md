# feature

Install devcontainer features

## Synopsis

  Install devcontainer features on the host system.                             

  - Install nodejs:                                                             

      $ feature node 

  - Install Go:                                                                 

      $ feature go 

```
feature [featurename] [flags]
```

## Options

```
  -r, --featureRoot string   Location to checkout feature repository (default "~/.features")
  -h, --help                 help for feature
  -u, --updateRepo           Update the feature repository
```

## See also

* [feature completion](feature_completion.md)	 - Generate the autocompletion script for the specified shell

