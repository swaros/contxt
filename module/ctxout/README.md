# Contxt Output Handle

> NOTE: execution is delayed by an timer. see the code in example folder

![example](https://github.com/swaros/docu-asset-store/blob/main/ctxout_example.gif)


this library maps different ways to render output.
- post filter
- injected PrinterInterface
- stream handler

usecase is to have an output handler, which can be used in different ways depending on the capabilities of the output device.
and this whithout to have to taking care of the output device.

the base idea is to have one way to render output, and extend them with different handlers, they are responsible for the output by itself.
so if i like to have an colored output, i just add an printer, which is responsible for the coloring.
so i expect i can add any color code to the output, and the printer will handle it even if the output device is not able to handle it.

for that being possible, any code, like defining a color, is done by an markup, that comes from the output handler.

so instead of writing
````go
fmt.Printf("\033[1;33m%s\033[0m", "hello world")
````
i can write
````go
ctxout.PrintLn(ctxout.NewMOWrap(),ctxout.BoldTag, ctxout.ForeYellow, "hello world", ctxout.ResetCode)
````
the printer will handle the markup and the output device will get the colored output.
and if the output device is not able to handle the markup, the printer will handle it and ignore the colorcodes.

the responsible outputhandler needs to be added to the output chain, before the markup is used. *(ctxout.PrintLn(ctxout.NewMOWrap(),...) )*

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



