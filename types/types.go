package types

import "net/url"

const (
	Asc  SortOrder = "asc"
	Desc SortOrder = "desc"

	Equal            FilterModifier = "eq"
	NotEqual         FilterModifier = "ne"
	LessThan         FilterModifier = "le"
	LessThanEqual    FilterModifier = "lte"
	GreaterThan      FilterModifier = "gt"
	GreaterThanEqual FilterModifier = "gte"
	Prefix           FilterModifier = "prefix"
	Like             FilterModifier = "like"
	NotLike          FilterModifier = "notlike"
	Null             FilterModifier = "null"
	NotNull          FilterModifier = "notnull"

	String    FieldType = "string"
	Multiline FieldType = "multiline"
	Masked    FieldType = "masked"
	Password  FieldType = "password"
	Float     FieldType = "float"
	Int       FieldType = "int"
	Date      FieldType = "date"
	Blob      FieldType = "blob"
	Boolean   FieldType = "boolean"
	Json      FieldType = "json"
	Version   FieldType = "version"

	Enum      FieldType = "enum"
	Reference FieldType = "reference"
	Schema    FieldType = "schema"
	Array     FieldType = "array"
	Map       FieldType = "map"
)

type SchemaResource struct {
	Id                string                  `json:"id"`
	Type              string                  `json:"type"`
	Links             map[string]url.URL      `json:"links"`
	ResourceFields    map[string]Field        `json:"fields"`
	ResourceMethods   []string                `json:"resourceMethods"`
	ResourceActions   map[string]SchemaAction `json:"resourceActions"`
	CollectionMethods []string                `json:"collectionMethods"`
	CollectionActions map[string]url.URL      `json:"collectionActions"`
	CollectionFields  map[string]Field        `json:"collectionFields"`
	CollectionFilters map[string]Filter       `json:"collectionFilters"`
}

type SchemaAction struct {
	Input  string `json:"input,omitempty"`
	Output string `json:"output,omitempty"`
}

type FieldType string

type Field struct {
	Type                FieldType     `json:"type"`
	Default             interface{}   `json:"default"`
	Unique              bool          `json:"unique"`
	Nullable            bool          `json:"nullable"`
	Create              bool          `json:"create"`
	Required            bool          `json:"required"`
	Update              bool          `json:"update"`
	MinLength           uint          `json:"minLength"`
	MaxLength           uint          `json:"maxLength"`
	Min                 int           `json:"min"`
	Max                 int           `json:"max"`
	Options             []interface{} `json:"options"`
	ValidChars          string        `json:"validChars"`
	InvalidChars        string        `json:"invalidChars"`
	ReferenceCollection url.URL       `json:"referenceCollection"`
	Satisfies           string        `json:"satisfies"`
}

type BaseResource struct {
	Type    string             `json:"type"`
	Links   map[string]url.URL `json:"links"`
	Actions map[string]url.URL `json:"actions"`
}

type Resource struct {
	BaseResource
	Id string `json:"id"`
}

type CollectionResource struct {
	BaseResource
	ResourceType   string                 `json:"resourceType"`
	Data           []interface{}          `json:"data"`
	Pagination     Pagination             `json:"pagination"`
	Sort           Sort                   `json:"sort"`
	SortLinks      map[string]url.URL     `json:"sortLinks"`
	Filters        map[string]Filter      `json:"filters"`
	CreateDefaults map[string]interface{} `json:"createDefaults"`
}

type SortOrder string

type Sort struct {
	Name    string    `json:"name"`
	Order   SortOrder `json:"order"`
	Reverse url.URL   `json:"reverse,string"`
}

type Pagination struct {
	First    url.URL `json:"first,string"`
	Previous url.URL `json:"previous,string"`
	Next     url.URL `json:"next,string"`
	Last     url.URL `json:"last,string"`
	Limit    uint    `json:"limit"`
	Total    uint    `json:"total"`
	Partial  bool    `json:"partial"`
}

type FilterModifier string

type Filter struct {
	Modifier FilterModifier `json:"modifier"`
	Value    string         `json:"value"`
}

type ErrorResponse struct {
	Type    string `json:"type"`
	Status  uint   `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

func NewError(status uint, code string, message string, detail string) ErrorResponse {
	return ErrorResponse{
		Type:    "error",
		Status:  status,
		Code:    code,
		Message: message,
		Detail:  detail,
	}
}

func NewResource(self url.URL, resourceType string) Resource {
	return Resource{
		BaseResource: BaseResource{
			Type: resourceType,
			Links: map[string]url.URL{
				"self": self,
			},
		},
	}
}

func NewCollectionResource(self url.URL) CollectionResource {
	return CollectionResource{
		BaseResource: BaseResource{
			Type: "collection",
			Links: map[string]url.URL{
				"self": self,
			},
		},
	}
}
