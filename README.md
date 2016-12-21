GG
==

Watch file changes and exec commands (YAML + GoLang)


### Installation
```
go get github.com/mattes/gg
```


### Usage

Create ``gg.yaml`` file in your working directory.

```yaml
watch:

- pattern: "*.txt"
  commands:
    hello: "echo hello world, txt"

- pattern: "*.go"
  command:
    hello: "echo hello world, go"
```

Run ``gg`` afterwards.
