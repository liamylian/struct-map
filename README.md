# struct-map

Concurrent safe struct map.

## Usage

```go
import "github.com/liamylian/struct-map"

type Book struct {
	Id        string `struct_map:"pk"`
	Title     string `struct_map:"index"`
	Desc      string
	Author    string `struct_map:"index"`
}

mapper, err := struct_map.NewMapper(&Book{})
if err != nil {
    return
}

book1 := &Book{
    Id:        "1",
    Title:     "Jane Eyre",
    Author:    "Charlotte Brontë",
}
book2 := &Book{
    Id:        "2",
    Title:     "Gone With the Wind",
    Author:    "Margaret Mitchell",
}

mapper.Add(book1)
mapper.Add(book2)

// boos1 = []interface{}{book1}
books1 := mapper.Get("Id", book1.Id)
// boos2 = []interface{}{book2}
books2 := mapper.Get("Title", book2.Title)
```