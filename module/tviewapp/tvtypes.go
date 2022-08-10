package tviewapp

type TvButton struct {
	OnClick func()
	OnFocus func()
	Label   string
}
