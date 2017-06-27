Enterprise Partner Feed Parser
==============================

Golang Apple EPF (Enterprise Partner Feed) Parser

## Example

```go
import "github.com/euskadi31/go-epf"

parser, err := epf.NewParser("itunes/artist")
if err != nil {
    panic(err)
}
defer parser.Close()

// Metadata of epf file, Fileds, Types, PrimaryKey, ...
md := parser.Metadata()

// Read first row of epf file
item, err := parser.Read()
if err != nil {
    panic(err)
}

// item is a map[string]interface{}

```


## License

go-epf is licensed under [the MIT license](LICENSE.md).
