package tviewapp

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/swaros/contxt/trigger"
)

const (
	EventOnStart = "OnStart"
	EventOnInit  = "OnInit"
)

type tvHandle struct {
	tvPrimitive *tview.Primitive
	listener    *trigger.Listener
}

type TViewApplication struct {
	app       *tview.Application
	pages     *tview.Pages
	mainFrame *tview.Frame
	OnStart   *trigger.Event
	OnInit    *trigger.Event
}

func NewApplication(fullscreen bool) *TViewApplication {
	tvApp := TViewApplication{}
	app := tview.NewApplication()
	tvApp.app = app
	pages := tview.NewPages()
	frame := tview.NewFrame(pages)
	frame.SetBorders(1, 1, 1, 1, 0, 0)
	frame.SetBackgroundColor(tcell.ColorGray)

	tvApp.mainFrame = frame
	tvApp.pages = pages

	if t, e := trigger.NewEvent(EventOnStart); e == nil {
		tvApp.OnStart = t
	} else {
		panic(e)
	}

	if t, e := trigger.NewEvent(EventOnInit); e == nil {
		tvApp.OnInit = t
	} else {
		panic(e)
	}

	app.SetRoot(frame, fullscreen)
	return &tvApp
}

func (t *TViewApplication) GetPages() *tview.Pages {
	return t.pages
}

func (t *TViewApplication) Start() (*TViewApplication, error) {
	trigger.UpdateEvents()
	t.OnInit.Send()
	err := t.app.Run()
	t.OnStart.Send()
	return t, err
}
