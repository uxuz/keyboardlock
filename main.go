package main

import (
	"errors"
	"log"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// accessibilitySettings opens the pane of System Settings that lists the
// apps allowed to control the computer.
const accessibilitySettings = "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility"

// errNoAccess is what keyboard.Disable returns until the user allows the
// app under Accessibility, which macOS requires of apps that filter input.
var errNoAccess = errors.New("the Accessibility permission is needed")

// keyboard turns the keys of every keyboard off and on. The pointer is
// never affected.
type keyboard interface {
	Disable() error
	Enable()
}

// app is the state the window shows. Its view builds the interface from
// it, on the main thread, whenever the window needs a frame: after input
// alone, as nothing else changes the state.
type app struct {
	kb       keyboard
	disabled bool
	err      error
}

func (a *app) toggle() {
	if a.disabled {
		a.kb.Enable()
		a.disabled = false
		return
	}
	a.err = a.kb.Disable()
	a.disabled = a.err == nil
}

// The themes give the widgets the colors of macOS: its label colors and
// system blue, over the material of the window.
var lightTheme, darkTheme = macTheme(ui.LightTheme()), macTheme(ui.DarkTheme())

func macTheme(t *ui.Theme) *ui.Theme {
	if t.Dark {
		t.Text, t.TextMuted = ui.RGBA(255, 255, 255, 0.85), ui.RGBA(255, 255, 255, 0.55)
		t.Accent, t.AccentPressed = ui.Hex("#0a84ff"), ui.Hex("#409cff")
	} else {
		t.Text, t.TextMuted = ui.RGBA(0, 0, 0, 0.85), ui.RGBA(0, 0, 0, 0.5)
		t.Accent, t.AccentPressed = ui.Hex("#007aff"), ui.Hex("#0068d9")
	}
	// Buttons of macOS do not change under the pointer.
	t.AccentHover = t.Accent
	return t
}

func (a *app) view(c *ui.Context) {
	t := lightTheme
	if c.Theme().Dark {
		t = darkTheme
	}
	c.SetTheme(t)
	// Shows the window's material.
	c.Root().Background(ui.Transparent)

	// The window has no title bar, so all of it that is not a control
	// moves it.
	ui.Column(c).Fill().Center().Gap(20).DragWindow().Children(func() {
		// Everything says the state at once: the color and the slash of
		// the icon, as on the system's own symbols, the title and the
		// switch.
		icon, title := keyboardIcon, "Keyboard Unlocked"
		from, to := ui.Hex("#a4a4a9"), ui.Hex("#7c7c81")
		if a.disabled {
			icon, title = keyboardOffIcon, "Keyboard Locked"
			from, to = ui.Hex("#4aa3ff"), ui.Hex("#007aff")
		}
		ui.Column(c).AlignItems(ui.Center).Gap(12).Children(func() {
			ui.Column(c).Size(64, 64).Center().Radius(15).
				Gradient(from, to, 180).Shadow(0, 1, 2.5, 0, ui.RGBA(0, 0, 0, 0.22)).
				TextColor(ui.Hex("#ffffff")).Children(func() {
				ui.Icon(c, icon).Size(42, 42)
			})
			ui.Text(c, title).FontSize(17).FontWeight(600)
		})

		// The switch is a row of its own, as the main switch of a pane of
		// System Settings is, and all of it toggles: a large target for a
		// hand holding a cloth. Its place never changes, so it stays under
		// the pointer for the second click.
		on := a.disabled
		row := ui.SwitchBase(c, &on).Size(232, 44).PaddingX(12).Radius(8)
		if row.Changed() {
			a.toggle()
		}
		fill, line := ui.RGBA(0, 0, 0, 0.035), ui.RGBA(0, 0, 0, 0.07)
		if t.Dark {
			fill, line = ui.RGBA(255, 255, 255, 0.05), ui.RGBA(255, 255, 255, 0.08)
		}
		row.Background(fill).Border(1, line).Children(func() {
			ui.Text(c, "Lock Keyboard").Grow(1)
			a.track(c, row.Animate("knob", b2f(a.disabled), 150*time.Millisecond))
		})

		if a.err != nil {
			ui.Column(c).Absolute().Left(24).Right(24).Bottom(16).AlignItems(ui.Center).Gap(3).Children(func() {
				note := "The keyboard could not be locked: " + a.err.Error() + "."
				if errors.Is(a.err, errNoAccess) {
					note = "Allow Keyboard Lock in Privacy & Security › Accessibility, then try again."
				}
				ui.Text(c, note).FontSize(11).TextColor(t.TextMuted).TextAlign(ui.Center)
				if errors.Is(a.err, errNoAccess) {
					ui.Link(c, "Open Accessibility Settings", accessibilitySettings).FontSize(11)
				}
			})
		}
	})
}

// track draws the switch of the row: the track in the accent color as far
// as pos, from 0 for off to 1 for on, has the knob moved across it.
func (a *app) track(c *ui.Context, pos float32) {
	t := c.Theme()
	off, knob := ui.RGBA(0, 0, 0, 0.12), ui.Hex("#ffffff")
	if t.Dark {
		off, knob = ui.RGBA(255, 255, 255, 0.16), ui.Hex("#e4e4e6")
	}
	ui.Box(c).Size(38, 22).Radius(11).Background(off.Mix(t.Accent, pos)).Draw(func(p *ui.Painter, r ui.Rect) {
		d := r.H - 4
		k := ui.Rect{X: r.X + 2 + pos*(r.W-r.H), Y: r.Y + 2, W: d, H: d}
		p.Shadow(k, d/2, 0, 1, 2.5, 0, ui.RGBA(0, 0, 0, 0.25))
		p.Fill(k, knob, d/2)
	})
}

func b2f(b bool) float32 {
	if b {
		return 1
	}
	return 0
}

func main() {
	a := &app{kb: newKeyboard()}
	mygo.App.WhenReady(func() {
		// A small panel in the manner of the system's own: its controls over
		// the content, on a translucent material, at one size.
		mygo.NewWindow(mygo.WindowOptions{
			Title:             "Keyboard Lock",
			Width:             280,
			Height:            300,
			TitleBarStyle:     mygo.TitleBarHidden,
			Vibrancy:          mygo.VibrancyUnderWindow,
			DisableResize:     true,
			DisableMaximize:   true,
			DisableFullScreen: true,
			// The window shows the interface MyGo draws, not a web page.
			Content: ui.View(a.view),
		})
	})
	// Closing the window quits the app, which must not leave Caps Lock as
	// the cleaning set it.
	mygo.App.OnQuit(func() {
		if a.disabled {
			a.kb.Enable()
		}
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
