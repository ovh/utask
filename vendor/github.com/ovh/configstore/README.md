# configstore

The configstore library aims to facilitate configuration discovery and management.
It mixes configuration items coming from various (abstracted) data sources, called *providers*.

[![GoDoc](https://godoc.org/github.com/ovh/configstore?status.svg)](https://godoc.org/github.com/ovh/configstore) [![Go Report Card](https://goreportcard.com/badge/github.com/ovh/configstore)](https://goreportcard.com/report/github.com/ovh/configstore)

## Items

An *item* is composed of 3 fields:
* **Key**: The name of the item. Does not have to be unique. The provider is responsible for giving a sensible initial value.
* **Value**: The content of the item. This can be either manipulated as a plain scalar string, or as a marshaled (JSON or YAML) blob for complex objects.
* **Priority**: An abstract integer value to use when priorizing between items sharing the same key. The provider is responsible for giving a sensible initial value.

## Configuration format

The item keys are *NOT* case-sensitive. Also, `-` and `_` characters are equivalent.

The exact input format of the configuration depends on the provider. Providers can either be loaded manually in your code, or controlled by the env variable `CONFIGURATION_FROM`.

### Example main.go

```go
func main() {
    configstore.InitFromEnvironment()

    val, err := configstore.GetItemValue("foo")
    if err != nil {
        panic(err)
    }
    fmt.Println(val)
}
```

**Outputs:**
```
bar
```

### Reading from a file

Env:
```sh
CONFIGURATION_FROM=file:foo.cfg
```

Contents of foo.cfg file:
```yaml
- key: foo
  priority: 12
  value: bar
```

Key/value pairs are read from a single file in yaml.

### Reading from env

Env:
```sh
CONFIGURATION_FROM=env:CONFIG
CONFIG_FOO=bar
```

Key/value pairs are read from the environment, with an optional prefix. Remember that key names are case-insensitive, and that `_` and `-` are equivalent in key names.

### Reading from a file hierarchy

Env:
```sh
CONFIGURATION_FROM=filetree:configdir
```

Contents of configdir directory:
```
foo
```

Contents of configdir/foo file:
```
bar
```

Key/value pairs are read by traversing a root directory. Each file in the dir represents an item: the filename is the key, the contents are the value.
To have several items sharing the same key, you can use a single level of sub-directory as such: `configdir/foo/bar1`, `configdir/foo/bar2`, ... The filenames `bar1`/`bar2` are not used in the resulting items.

### Reading from a custom source

These built-in providers implement common sources of configuration, but configstore can be expanded with other data sources.
See [Example: multiple providers](#example-multiple-providers).

## Example 101

file.txt:
```yaml
- key: foo
  value: bar
- key: baz
  value: bazz
```

```go
func main() {
    configstore.File("/path/to/file.txt")
    v, err := configstore.GetItemValue("foo")
    fmt.Println(v, err)
}
```

This very basic example describes how to get a string out of a configuration file (which can be JSON or YAML).
To do more advanced configuration manipulation, see the next examples.

## Example: multiple providers

Configuration *Providers* represent an abstract data source. Their only role is to return a list of *items*.

Some built-in implementations are available (in-memory, file, env, ...), but the library exposes a way to register a provider *factory*, to extend it and bridge with any other existing system.

Example mixing several providers
```go
// custom provider with hardcoded values
func MyProviderFunc() (configstore.ItemList, error) {
    ret := configstore.ItemList{
        Items: []configstore.Item{
            // an item has 3 components: key, value, priority
            // they are defined by the provider, but can be modified later by the library user
            configstore.NewItem("key1", `value1-higher-prio`, 6),
            configstore.NewItem("key1", `value1-lower-prio`, 5),
            configstore.NewItem("key2", `value2`, 5),
        },
    }
    return ret, nil
}

func main() {

    configstore.RegisterProvider("myprovider", MyProviderFunc)
    configstore.File("/path/to/file.txt")
    configstore.Env("CONFIG_")

    // blends items from all sources
    items, err := configstore.GetItemList()
    if err != nil {
        panic(err)
    }

    for _, i := range items.Items {
        val, err := i.Value()
        if err != nil {
            panic(err)
        }
        fmt.Println(i.Key(), val, i.Priority())
    }
}
```

## Example: advanced filtering

When calling *configstore.GetItemList()*, the caller gets an *ItemList*.

This object contains all the configuration items. To manipulate it, you can use a *ItemFilter* object, which provides convenient helper functions to select and reorder the items.

All objects are safe to use even when the item list is empty.

Assuming the following configuration file:
```yaml
- key: database
  value: '{"name": "foo", "ip": "192.168.0.1", "port": 5432, "type": "RO"}'
- key: database
  value: '{"name": "foo", "ip": "192.168.0.1", "port": 5433, "type": "RW"}'
- key: database
  value: '{"name": "bar", "ip": "192.168.0.1", "port": 5434, "type": "RO"}'
- key: other
  value: misc
```

Our program wants to retrieve database credentials, favoring RW over RO when both are present for the same database:
```go
func main() {

    configstore.File("example.txt")

    items, err := configstore.GetItemList()
    if err != nil {
        panic(err)
    }

    // we start by building a filter to manipulate our configuration items
    // we will apply it on our items list later
    filter := configstore.Filter()

    // extract only the "database" items
    filter = filter.Slice("database")

    // now we have a list of database objects, with 3 items:
    // {"name": "foo", "ip": "192.168.0.1", "port": 5432, "type": "RO"}
    // {"name": "foo", "ip": "192.168.0.1", "port": 5433, "type": "RW"}
    // {"name": "bar", "ip": "192.168.0.1", "port": 5434, "type": "RO"}
    //
    // the "database" key provides too little information to further classify items
    // we need to know the database name and type to further regroup and prioritize them
    // for that, we need to drill down into the actual item value

    // we need to unmarshal the JSON representation of every item
    // we pass a factory function that instantiates objects of the correct concrete type
    // it will be called for each item in the sublist, and each item is then unmarshaled (JSON or YAML) into the returned object
    filter = filter.Unmarshal(func() interface{} { return &Database{} })

    // now we want to actually index and lookup by database name, instead of the generic "database" key
    // we apply a rekey function that does payload inspection and renames each item
    //
    // our rekey function was written with knowledge of the objects being manipulated,
    // and uses the unmarshaled objects, not the raw text
    filter = filter.Rekey(rekeyByName)

    // we have redundant items: database "foo" is present twice (RO and RW)
    // we want to favor the RW instance if possible
    // we apply a reordering function that re-assigns item priorities
    // after inspecting the unmarshaled objects
    filter = filter.Reorder(prioritizeRW)

    // we only need 1 of each distinct database
    // items relating to the same database now share the same key,
    // and their priority properly reflects whether they are more or less important (RO or RW)
    // we apply a squash to keep only the items with the single highest priority value, for each key
    // = RW items of each database if available, RO otherwise
    filter = filter.Squash()

    // now we have only 2 items left:
    // {"name": "foo", "ip": "192.168.0.1", "port": 5433, "type": "RW"}
    // {"name": "bar", "ip": "192.168.0.1", "port": 5434, "type": "RO"}

    // we can finally apply it on our list
    items = filter.Apply(items)

    // all these transformations can be chained as a one-liner description of the filter steps:
    filter = configstore.Filter().Slice("database").Unmarshal(func() interface{} { return &Database{} }).Rekey(rekeyByName).Reorder(prioritizeRW).Squash()
    items, err = filter.GetItemList() // shortcut: applies the filter to the full list from configstore.GetItemList()
    if err != nil {
        panic(err)
    }

    // declaring your filter separately like this lets you define it globally and execute it later
    // that way, you can use its description (String()) to generate usage information.
    //
    // in this example, filter.String() would output:
    // database: {"name": "", "ip": "", "port": "", "type": ""}
}

type Database struct {
    Name string `json:"name"`
    IP   string `json:"ip"`
    Port int    `json:"port"`
    Type string `json:"type"`
}

func rekeyByName(s *configstore.Item) string {
    i, err := s.Unmarshaled()
    // we see here the error that was produced when we called *ItemList.Unmarshal(...)*
    // we ignore it for now, it will be handled when the *main()* retrieves the object.
    if err == nil {
        return i.(*Database).Name
    }
    return s.Key()
}

func prioritizeRW(s *configstore.Item) int64 {
    i, err := s.Unmarshaled()
    if err == nil {
        if i.(*Database).Type == "RW" {
            return s.Priority() + 1
        }
    }
    return s.Priority()
}
```
