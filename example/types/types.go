package types

type Foo struct {
	Name    string `validation:"minLength=4"`
	Address string
	City    string
	State   string `validation:"oneOf=[WI,AZ,CA]"`
	Zip     string
}