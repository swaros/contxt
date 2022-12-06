package tviewapp

import (
	"errors"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/swaros/contxt/module/trigger"
)

const (
	// default events
	EventOnStart = "OnStart"
	EventOnInit  = "OnInit"

	// defaut templates
	TplPageButtonMenu = 1
)

type TViewApplication struct {
	app       *tview.Application
	pages     *tview.Pages
	mainFrame *tview.Frame
	OnStart   *trigger.Event
	OnInit    *trigger.Event
}

// NewApplication creates a new Application with some default settings
// so it adds a main frame that includes some pages
func NewApplication(fullscreen bool) *TViewApplication {
	initDefaults()                            // init template defaults in template.go
	tvApp := TViewApplication{}               // create the main application
	app := tview.NewApplication()             // creates the main tview.Application
	tvApp.app = app                           // just assign them to the TViewApp
	pages := tview.NewPages()                 // creates the pages
	frame := tview.NewFrame(pages)            // creates the mainframe
	frame.SetBorders(1, 1, 1, 1, 0, 0)        // apply the default sizes
	frame.SetBackgroundColor(tcell.ColorGray) // apply the default colors
	tvApp.mainFrame = frame                   // just assign them to the TViewApp
	tvApp.pages = pages                       // just assign them to the TViewApp

	// setup the default trigger handler
	if err := trigger.NewEvents([]string{EventOnStart, EventOnInit}, func(e *trigger.Event) {
		switch e.GetName() { // assign the events to container by the event name
		case EventOnInit:
			tvApp.OnInit = e
		case EventOnStart:
			tvApp.OnStart = e
		}
	}); err != nil {
		panic(err)
	}

	app.SetRoot(frame, fullscreen).EnableMouse(true) // set the frame as the root element
	return &tvApp
}

func (t *TViewApplication) SetHeader(header string) {
	if t.mainFrame != nil {
		t.mainFrame.Clear().
			AddText(header, true, tview.AlignCenter, tcell.ColorLightBlue)
	} else {
		panic("header is not created")
	}
}

func (t *TViewApplication) NewPage(name string, style tvPageStyle, args ...interface{}) error {
	style.values = args
	if page := CreatePageByStyle(style); page != nil {
		t.pages.AddPage(name, page, true, true)
		return nil
	}
	return errors.New("this style with id " + strconv.Itoa(style.template) + " not exists")
}

func (t *TViewApplication) NewPageWithFlex(name string) *tview.Flex {
	createFlex := tview.NewFlex()
	t.pages.AddPage("name", createFlex, true, true)
	return createFlex
}

func (t *TViewApplication) GetPages() *tview.Pages {
	return t.pages
}

// Start executes the Application and send the initial events
func (t *TViewApplication) Start() (*TViewApplication, error) {
	trigger.UpdateEvents()
	t.OnInit.Send()
	err := t.app.Run()
	t.OnStart.Send()
	return t, err
}

func (t *TViewApplication) Stop() {
	t.app.Stop()
}
