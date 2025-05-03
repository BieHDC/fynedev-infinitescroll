package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// Attach a hover to any canvas object
type anyHover struct {
	widget.BaseWidget
	child  fyne.CanvasObject
	cbIn   func(*desktop.MouseEvent)
	cbMove func(*desktop.MouseEvent)
	cbOut  func()
}

var _ desktop.Hoverable = (*anyHover)(nil)

func (ah *anyHover) MouseIn(de *desktop.MouseEvent) {
	if ah.cbIn != nil {
		ah.cbIn(de)
	}
}

func (ah *anyHover) MouseMoved(de *desktop.MouseEvent) {
	if ah.cbMove != nil {
		ah.cbMove(de)
	}
}

func (ah *anyHover) MouseOut() {
	if ah.cbOut != nil {
		ah.cbOut()
	}
}

func NewAnyHover(child fyne.CanvasObject, cbIn func(*desktop.MouseEvent), cbMove func(*desktop.MouseEvent), cbOut func()) *anyHover {
	ah := &anyHover{
		child:  child,
		cbIn:   cbIn,
		cbMove: cbMove,
		cbOut:  cbOut,
	}
	ah.ExtendBaseWidget(ah)
	return ah
}

func (ah *anyHover) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(ah.child)
}

// Attach an arbitrary function to a secondary Tap
type secondaryTapper struct {
	widget.BaseWidget
	child fyne.CanvasObject
	cb    func(*fyne.PointEvent)
}

var _ fyne.SecondaryTappable = (*secondaryTapper)(nil)

func (st *secondaryTapper) TappedSecondary(pe *fyne.PointEvent) {
	if st.cb == nil {
		panic("no callback set, pointless use of widget")
	}
	st.cb(pe)
}

func NewSecondaryTapperLayer(child fyne.CanvasObject, cb func(*fyne.PointEvent)) *secondaryTapper {
	st := &secondaryTapper{child: child, cb: cb}
	st.ExtendBaseWidget(st)
	return st
}

func (st *secondaryTapper) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(st.child)
}

// i just need a button that does not implement hover
var _ fyne.Tappable = (*tappableIcon)(nil)

type tappableIcon struct {
	widget.Icon
	cb func()
}

func newTappableIcon(res fyne.Resource, cb func()) *tappableIcon {
	t := &tappableIcon{}
	t.ExtendBaseWidget(t)
	t.SetResource(res)
	t.cb = cb

	return t
}

func (t *tappableIcon) Tapped(_ *fyne.PointEvent) {
	t.cb()
}
