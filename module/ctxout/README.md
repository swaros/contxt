# Contxt Output Handle

> NOTE: execution is delayed by an timer. see the code in example folder

![example](https://github.com/swaros/docu-asset-store/blob/main/ctxout_example.gif)


this library maps different ways to render output.
- post filter
- injected PrinterInterface
- stream handler


## basic usage

#### simple output
````go
ctxout.Print("hello world")
````

#### output with post filter (table for example)

a Post filter is an output handler, which is called after the output is rendered. 
once added, it will be called on every output, until it is removed.

````go
ctxout.AddPostFilter(ctxout.NewTabOut())
ctxout.PrintLn(
			ctxout.Table(
				ctxout.Row(
					ctxout.TD("hello", "size=50", "origin=2"),
					ctxout.TD("world", "size=50", "origin=1"),
				),
			),
		)

````

#### output with injected printer (colored for example)

a printer is an output handler, which is called before the output is rendered. but instead of a post filter, it is only called once.

so it must be added before the output is rendered and will not be called again on the next print command.


````go
ctxout.PrintLn(
    ctxout.NewMOWrap(), 
    ctxout.BoldTag, 
    ctxout.ForeYellow, 
    "hello world", 
    ctxout.ResetCode,
)
````

### examples

if you checkout the sources, just run
````bash
go run ./module/ctxout/examples/base.go
````

this example is made like an little on screen demo app, to show how the output handler behaves.
so the delay is done by an timer and it is not a performance issue.



