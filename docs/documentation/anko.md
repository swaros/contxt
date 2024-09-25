# anko
<!-- TOC -->
- [anko](#anko)
  - [what is anko](#what-is-anko)
  - [runtime specific](#runtime-specific)
  - [how to run anko commands](#how-to-run-anko-commands)
  - [Anko Usage](#anko-usage)
    - [Basics](#basics)
      - [Variables](#variables)
      - [Arrays](#arrays)
      - [Maps](#maps)
        - [Checking for variable types](#checking-for-variable-types)
        - [variable type conversion](#variable-type-conversion)
        - [type Conversion Functions](#type-conversion-functions)
      - [Return values](#return-values)
      - [Functions](#functions)
        - [Module](#module)
      - [Control Structures](#control-structures)
        - [if](#if)
        - [else if](#else-if)
        - [and](#and)
        - [or](#or)
        - [for](#for)
          - [using arrays](#using-arrays)
          - [using ranges](#using-ranges)
          - [using maps](#using-maps)
          - [until condition is true](#until-condition-is-true)
          - [endless loop](#endless-loop)
          - [endless loop with break](#endless-loop-with-break)


<!-- /TOC -->

## what is anko
checkout [Anko](https://github.com/mattn/anko) to get a understanding of the language.
it is basically a go script that is executed in a go runtime. in context it is extended with some functions to interact with the context.

by using anko  commands you can write way more specific scripts than you could do with bash. and they are independent from the host system.

## runtime specific

anko commands will be executed before the script. so you can use them to prepare variables for the later script.

```yaml
task:
  - id: example
    cmd:
      - setVar("myvar", "hello world")
    script:
      - echo $myvar
```

even this example makes no sense, it shows how you can use anko to prepare the script.


## how to run anko commands

for testing anko you can use the anko command from contxt

for executing simple anko commands you can use the `contxt anko` command.
each command is separated by a space.

```bash
contxt anko 'c = 1+1' 'println(c)'
```
this example is the same as
```
c = 1+1
println(c)
```
to execute a script you can use the `contxt anko -f` command
```bash
contxt anko -f /path/to/script.ank
```

## Anko Usage
### Basics

#### Variables
anko is based on go, so you can use a lot of go functions, but the syntax is a bit different. so it aims to be more easy to use in terms of scripting instead of programming.

as example to point out the differences, you can always use `=` to assign existing variables. and the same for new variables. 

```go
a = 1
b = 2
c = a + b
a = 3
d = a + b
```

the same in golang would be
```go
a := 1
b := 2
c := a + b
a = 3
d := a + b
```

there is also no need to declare the type of the variable. anko will do this for you.

```go
a = 1
b = "hello"
c = true
```
this, of course have some drawbacks. because of this magic "taking care about variable assignment" you can do a reassignement of the variable, even it is a different type.
```go
a = 1
a = "hello"

c = a + 1
println(c) // hello1
```

this will print `hello1` because a is a string at the end.just to kkep this is mind. Anko is not for programming, it is for scripting. so you can do a lot of things, but you have to keep in mind that you have to take care about the types of the variables.

similar to go, you can chain variable assignments.

```go
a, b = 1, 5
println(a, b) // 1 5
```

#### Arrays
you can define arrays in anko. the syntax is the same as in go. but you can also use the `[]` to define an array.

```go
a = [1, 2, 3]
println(a[0]) // 1
```

you can also define arrays with different types. in this case the array will be an array of interfaces. this is matching to any type of variable.
```go
arr = []interface{1, 2}
println(arr) // [1 2]

arr2 = []interface{1, "hello"}
println(arr2) // [1 hello]

arr3 = []int{1, 2, 3}
println(arr3) // [1 2 3]

arr4 = []string{"this","is", "sparta"}
println(arr4) // [this is sparta]
```

but assigning a variable to a specific type will not work. but ehis will also **not** being a error. so this works, even the outcome is unexpected.
```go
arr = []string{"start", 1, "test", 3, "end"}
println(arr) // [start  test  end]


arr = []int{0.75, 1, 3, 8.33}
println(arr) // [0 1 3 8]
```
this happens because the variable is defined as an array of strings. so the variables will be casted to a string. this is a bit tricky, but you have to keep this in mind.

there are also exceptions. so not in any case the mapping will happen.

```go
arr = []float64{0.75, 1, 3, 8.33, "hi"}
// Error: cannot use type string as type float64 as slice value
```

this will end up in an error. because the variable is defined as an array of float64. so the last element is a string and this will not be casted to a float64.


#### Maps
you can define maps in anko. the syntax is the same as in go. but you can also use the `{}` to define a map.

```go
a = {"key": "value"}
println(a["key"]) // value
```

you can also define maps with different types. in this case the map will be a map of interfaces. this is matching to any type of variable.
```go
m = map{1: "hello", "key": 2}
println(m) // map[1:hello key:2]

m2 = map{1: "hello", "key": 2, 3: true}
println(m2) // map[1:hello key:2 3:true]

m3 = map{1: "hello", "key": 2, 3: true, "test": 3.14}
println(m3) // map[1:hello key:2 3:true
```

while runtime, you can access the map by the key. if the key is not available, the outcome will be `nil`.
to use the key you can use the `.` operator.

```go
m = map{1: "hello", "key": 2}
println(m.key) // 2
```

to check if a key exists in a map, try to access them.
you will get a second return value, which is a boolean. if the key exists, the boolean will be `true`, otherwise `false`.

```go
m = map{1: "hello", "key": 2}

check, ok = m[1]
if ok {
    println("found the key 1 ",check)
} else {
    println("not found")
}
```
of course this is working for string keys as well.

```go
m = map{"key": "hello", "key2": 2}
check, ok = m["key"]
if ok {
    println("found the key ",check)
} else {
    println("not found")
}
```

##### Checking for variable types
use the `typeOf` function to check the type of a variable.

```go
a = 1
println(typeOf(a)) // int64
```

##### variable type conversion
you can convert a variable to a different type by using the `to` function.

```go
a = 1
b = toString(a)
println(typeOf(b)) // string
```

##### type Conversion Functions
| Function | Description |
|----------|-------------|
| `toString` | converts a variable to a string |
| `toInt` | converts a variable to an integer |
| `toFloat` | converts a variable to a float |
| `toBool` | converts a variable to a boolean |


#### Return values
different to golang, anko will return the outcome of a function depending how the variable is defined. to explain this, lets have a look at the following example.

```go
func testResult() {
    return 1, "Hello"
}
a = testResult()
println(a[0], a[1]) // 1 Hello
```

the function testResult returns two values. in anko you can assign the return values to a variable. the variable will be an array with the return values. so you can access the values by the index.

but you can also define the variable as a tuple. in this case you can access the values by the name.

```go
func testResult() {
    return 1, "Hello"
}
a, b = testResult()
println(a, b) // 1 Hello
```

like in go you can also use the underscore to ignore a return value.

```go
func testResult() {
    return 1, "Hello"
}
a, _ = testResult()
println(a) // 1
```

#### Functions
you can define functions in anko. the syntax is the same as in go. but you have to keep in mind that you have to define the function before you use it.

```go
func test() {
    println("Hello")
}
test() // Hello
```

##### Module
modules are a way to organize your functions. you can define a module and use the functions in the module.

```go
module test {
    func test() {
        println("Hello")
    }
}
test.test() // Hello
```

#### Control Structures
you can use the following control structures in anko.

- if
- else
- else if
- for
- break
- continue
- return

##### if
```go
a = 1
if a == 1 {
    println("a is 1")
} else {
    println("a is not 1")
}
```

##### else if
```go
a = 1
if a == 1 {
    println("a is 1")
} else if a == 2 {
    println("a is 2")
} else {
    println("a is not 1 or 2")
}
```

##### and
```go
a = 1
b = 2
if a == 1 && b == 2 {
    println("a is 1 and b is 2")
}
```

##### or
```go
a = 1
b = 2
if a == 1 || b == 3 {
    println("a is 1 or b is 3")
}
```

##### for
###### using arrays
```go
for i in [1, 2, 3, 4, 5] {
    println(i)
}
```
###### using ranges
```go
for i in range(5) {
    println(i)
}
```
###### using maps
```go
for k, v in {"key": "value", "key2": "value2"} {
    println(k, v)
}
```

###### until condition is true
```go
i = 0
for i < 2 {
	println(i)
	i++
}
```

###### endless loop
```go
i = 0
for {
    println(i)
    i++
}
```
###### endless loop with break
```go
i = 0
for {
	println(i)
	i++
	if i > 1 {
		break
	}
}
```

