package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/dsnidr/encgen-go/example"
)

func main() {
	buf := bytes.NewBuffer(nil)

	itemsEnc := example.NewParcelEncoder(buf).Start().
		ID("test").
		Name("name").
		StartItems()

	itemsEnc.AddItems(
		&example.Item{
			SKU:    "item1",
			Name:   "test item 1",
			Weight: 592.0,
		},
		&example.Item{
			SKU:    "item2",
			Name:   "test item 2",
			Weight: 182.15,
		},
	)

	itemsEnc.AddItems(
		&example.Item{
			SKU:    "item3",
			Name:   "test item 3",
			Weight: 23717.85,
		},
		&example.Item{
			SKU:    "item2",
			Name:   "test item 4",
			Weight: 10.0,
		},
	)

	metaEnc := itemsEnc.FinishItems().StartMetadata()
	metaEnc.AddMetadata("one", "two", "three", "four")
	metaEnc.AddMetadata("five")

	if err := metaEnc.FinishMetadata().
		Tags([]*example.Tag{
			{ID: "tag1", Name: "Home Improvement"},
			{ID: "tag2", Name: "Digital Goods"},
		}).Finish(); err != nil {
		panic(err)
	}

	fmt.Println("RAW OUTPUT")
	fmt.Println(buf.String())

	fmt.Print("\n\n")
	fmt.Println("FORMATTED OUTPUT")

	// Proof of structure compatability + formatted output
	unmarshalled := &example.Parcel{}
	if err := json.Unmarshal(buf.Bytes(), unmarshalled); err != nil {
		panic(err)
	}

	bs, err := json.MarshalIndent(unmarshalled, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(bs))
}
