# Yet Another Map Converter

## Description

this library is build to avoid the need of knowing the data structure for reading data files in different formats.
so the problem is depending on two types of data structures you need to know to read data from a source and use them in your code:
- **Array**: which is a list of values of the same type.
- **Object**: which is a list of key-value pairs.

### The Issue

for parsing unknown source data, you run in trouble, if you don't know the "root" data structure.


objects read as `map[string]interface{}` 


```json
{
    "name": "John Doe",
    "age": 30,
    "cars": [
        "Ford",
        "BMW",
        "Fiat"
    ]
}
```
and arrays read as `[]interface{}`

```json
[
    "Ford",
    "BMW",
    "Fiat"
]
```

so you need to know the data structure to read the data, and use it in your code. or you will get an error. like this one:

```go
jsonSource := []byte(`[
		{
			"id": 1,
			"name": "John Doe",
			"email": "sfjhg@sdfkfg.net",
			"phone": "555-555-5555"
		},
		{
			"id": 2,
			"name": "Jane Doe",
			"email": "lllll@jjjjj.net",
			"phone": "555-555-5521"
		},
		{
			"id": 3,
			"name": "John Smith",
			"email": "kjsahd@fjjgg.net",
			"phone": "888-555-5555"
		}
	]`)

	var data map[string]interface{}
	if err := json.Unmarshal(jsonSource, &data); err != nil {
		panic(err)
	}
```
this would result in the following error:

```bash
panic: json: cannot unmarshal array into Go value of type map[string]interface {}
```

the code is obvious, we are trying to unmarshal an array into a map, which is not possible. all of them is fine if we can trust the source data, but what if we can't? for example to have some more generic code, or to read data from a user input, or a file, or a database, or a network connection, or ... .

### The Solution

YAMC is an abbreviation for Yet Another Map Converter, which is a library to convert data from one data structure to another. it is a simple library, which is build to solve the problem of reading data from unknown sources, and convert them to a known data structure, so you can use them in your code.

the goal is to have a library that reads data from different sources, depending on Readers specialized to a Datatype (json yaml and so on) , and convert them to a known data structure, and provide a simple API to get the data from the source.

one usecase is to have no longer taking care about value files they are used together with go/template, and you even not know what value files provide what data, and how they paths are composed. because booth is then depending the specific usage later on.


#### depending this data:
for the example we use just an byte array.

```go
jsonSource := []byte(`[
   {
        "id": 1,
        "name": "John Doe",
        "email": ""
    },
]`)
```

#### Create an instance of Yamc

```go
mapData := yamc.New()
```

#### use an Reader to parse the data

```go

if err := mapData.Parse(yamc.NewJsonReader(), jsonSource); err != nil {
    panic(err)
}
```
#### get what structure is used
this is useful if you want to know what data structure was used to parse the data. at least if you have to compose paths to get the data from the source.
    
```go
switch mapData.GetSourceDataType() {
case yamc.TYPE_ARRAY:
	fmt.Println("source data was in form of []interface{}")
case yamc.TYPE_STRING_MAP:
	fmt.Println("source data was in form of map[string]interface{}")
default:
	fmt.Println("this should not happen")
}
```


#### get the data from the source
a simple dotted path where each string is used as keyname, and each number is used as index.
`0.name` would be the name of the first element in the array. index 0 the element named "name".
```go
data, err := mapData.Get("0.name")
if err != nil {
    panic(err)
}
fmt.Println(fmt.Sprintf("%v", data))
```
#### more complex example paths
for more complex paths, YAMC implements [tidwall/gson](https://github.com/tidwall/gjson) so you can use the same syntax as in gson.

```go
source := []byte(`{
	"name": {"first": "Tom", "last": "Anderson"},
	"age":37,
	"children": ["Sara","Alex","Jack"],
	"fav.movie": "Deer Hunter",
	"friends": [
	  {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
	  {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
	  {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
	]
  }`)

data := yamc.New()
data.Parse(yamc.NewJsonReader(), source)

paths := []string{
	"name.last",
	"age",
	"children",
	"children.#",
	"children.1",
	"child*.2",
	"c?ildren.0",
	"fav\\.movie",
	"friends.#.first",
	"friends.1.last",
	"friends.#(last==\"Murphy\").first",
	"friends.#(last==\"Murphy\")#.first",
	"friends.#(age>45)#.last",
	"friends.#(first%\"D*\").last",
	"friends.#(first!%\"D*\").last",
	"friends.#(nets.#(==\"fb\"))#.first",
}
// just use any of the examples from gson README.md (https://github.com/tidwall/gjson)
for _, gsonPath := range paths {
	name, _ := data.GetGjsonString(gsonPath)	
	fmt.Println(gsonPath, " => ", name)
}
```

see in `example/gson/main.go`.