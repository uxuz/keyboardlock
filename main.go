package main

import (
	"errors"
	"log"

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

func (a *app) view(c *ui.Context) {
	// The default theme follows the system: its light or dark appearance,
	// the accent color the user chose and Increase Contrast.
	t := c.Theme()
	// Shows the window's material.
	c.Root().Background(ui.Transparent)

	// The window has no title bar, so all of it that is not a control
	// moves it.
	ui.Column(c).Fill().Center().Gap(20).DragWindow().Children(func() {
		// Everything says the state at once: the color and the slash of
		// the icon, as on the system's own symbols, the title and the
		// switch.
		icon, title := keyboardIcon, "Keyboard Unlocked"
		white := ui.Hex("#ffffff")
		from, to, glyph := ui.Hex("#a4a4a9"), ui.Hex("#7c7c81"), white
		if a.disabled {
			icon, title = keyboardOffIcon, "Keyboard Locked"
			from, to, glyph = t.Accent.Mix(white, 0.3), t.Accent, t.AccentText
		}
		ui.Column(c).AlignItems(ui.Center).Gap(12).Children(func() {
			ui.Column(c).Size(64, 64).Center().Radius(15).
				Gradient(from, to, 180).Shadow(0, 1, 2.5, 0, ui.RGBA(0, 0, 0, 0.22)).
				TextColor(glyph).Children(func() {
				ui.Icon(c, icon).Size(42, 42)
			})
			ui.Text(c, title).FontSize(17).FontWeight(600)
		})

		// The switch is a row of its own, as the main switch of a pane of
		// System Settings is, and all of it toggles: a large target for a
		// hand holding a cloth. Its place never changes, so it stays under
		// the pointer for the second click.
		on := a.disabled
		row := ui.Row(c).Size(232, 44).PaddingX(12).Radius(8)
		toggled := row.Clicked()
		fill, line := ui.RGBA(0, 0, 0, 0.035), ui.RGBA(0, 0, 0, 0.07)
		if t.Dark {
			fill, line = ui.RGBA(255, 255, 255, 0.05), ui.RGBA(255, 255, 255, 0.08)
		}
		row.Background(fill).Border(1, line).Children(func() {
			ui.Text(c, "Lock Keyboard").Grow(1)
			toggled = ui.Switch(c, &on).Label("Lock Keyboard").Changed() || toggled
		})
		if toggled {
			a.toggle()
		}

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
