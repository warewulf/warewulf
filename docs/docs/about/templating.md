---
id: templating
title: Templating
---

# Templating

## loops

### range

iterate over elements of an array

```go
{{ range $devname, $netdev := .NetDevs }}
    # netdev = {{ $netdev.Hwaddr }}
{{ end }}
```

#### increment variable in loop

iterate over elements of an array and increment `i` each loop cycle

```go
{{ $i := 0 }}
{{ range $devname, $netdev := .NetDevs }}
    # netdev{{$i}} = {{ $netdev.Hwaddr }}
    {{ $i = inc $i }}
{{ end }}
```

#### decrement

iterate over elements of an array and decrement `i` each loop cycle

```go
{{ $i := 10 }}
{{ range $devname, $netdev := .NetDevs }}
    # netdev{{$i}} = {{ $netdev.Hwaddr }}
    {{ $i = dec $i }}
{{ end }}
```
