# encgen-go

**Streamable JSON encoder generator**

## Features

- **Automatic encoder generation** from Go structs
- **Fluent API** ensuring correct encoding order
- **Efficient batch streaming** for slice fields (tagged with `enc:"batch"`)
- **Supports nested structs and custom types**

## Installation

```sh
go install github.com/dsnidr/encgen-go/cmd/encgen@latest
```

## Usage

### 1. Define a struct

Mark any fields you want to stream in batches with `enc:"batch"`.

```go
type Parcel struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Items    []*Item  `json:"items" enc:"batch"`
	Metadata []string `json:"metadata" enc:"batch"`
	Tags     []*Tag   `json:"tags"`
}

type Item struct {
	SKU    string  `json:"sku"`
	Name   string  `json:"name"`
	Weight float64 `json:"weight"` // in kgs
}

type Tag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
```

### 2. Run the generator

Run `go generate` where your struct is defined.

You can do this manually:

```sh
encgen --name <struct name>
```

or using `go generate`:

```go
//go:generate go run github.com/dsnidr/encgen-go/cmd/encgen --name Parcel
type Parcel struct {
  ...
}
```

### 3. Use the generated encoder

```go
buf := bytes.NewBuffer(nil)

itemsEnc := NewParcelEncoder(buf).
  Start().
  ID("test").
  Name("name").
  StartItems()

// your batching logic would go here
itemsEnc.AddItems(
  &Item{ID: "item1", Name: "test item 1", Weight: 592.0},
  &Item{ID: "item2", Name: "test item 2", Weight: 182.15},
  &Item{ID: "item3", Name: "test item 3", Weight: 23717.85})

itemsEnc.AddItems(
  &Item{ID: "item4", Name: "test item 4", Weight: 10.0})
...

err := itemsEnc.FinishItems().
  StartMetadata().
  AddMetadata("one", "two", "three").
  AddMetadata("four", "five").
  FinishMetadata().
  Tags([]*Tag{
    {ID: "tag1", Name: "Home Improvement"},
	  {ID: "tag2", Name: "Digital Goods"},
  }).
  Finish()

if err != nil {
  panic(err)
}

fmt.Println("OUTPUT")
fmt.Println(buf.String())

// OUTPUT
// {"id":"test","name":"name","items":[{"sku":"item1","name":"test item 1","weight":592},{"sku":"item2","name":"test item 2","weight":182.15},{"sku":"item3","name":"test item 3","weight":23717.85},{"sku":"item2","name":"test item 4","weight":10}],"metadata":["one","two","three","four","five"],"tags":[{"id":"tag1","name":"Home Improvement"},{"id":"tag2","name":"Digital Goods"}]}
```

