# Keyboard Lock

A tiny macOS app that turns the keyboard's keys off with one click and back
on with another, so you can wipe the keyboard without typing anything. The
trackpad and mouse keep working.

This is a toy project, made to try out [MyGo](https://mygo.egoist.dev/docs),
a toolkit for building desktop apps in Go. The window is MyGo's native UI,
written in Go alone, with no web page behind it. Expect rough edges: the app
is signed ad hoc and has only been used on one Mac.

The code, the app icon and this README were mostly generated with
[Claude](https://claude.com/claude-code), and reviewed and steered by a
human.

## Run

You need Go 1.27.1 or later and macOS.

```sh
go tool mygo build
open "build/darwin-arm64/Keyboard Lock.app"
```

`go tool mygo dev` runs a development version that restarts when the code
changes, and `go test` runs the tests of the view.

The first time you flip the switch, macOS asks for the Accessibility
permission, which it requires of apps that filter input: allow Keyboard Lock
under System Settings › Privacy & Security › Accessibility, then flip the
switch again. Every rebuild is a new app to macOS, which then wants the
permission given again.

## How it works

While the switch is on, the app holds a Quartz event tap that drops every
key event before any app gets it. Pointer events never pass through it. The
tap belongs to the app's process, so quitting the app, or a crash, gives the
keys back.

Not blocked: the Touch ID and power button, and typing into password
fields, the lock screen's included, whose keys macOS hides from every app.

| File                 |                                                  |
| -------------------- | ------------------------------------------------ |
| `main.go`            | the window and its switch                        |
| `keyboard_darwin.go` | the event tap, called without cgo through purego |
| `icons.go`           | the glyphs the window shows                      |
| `scripts/icon.swift` | draws the app icon, `resources/icon.png`         |

## Credits

The keyboard glyphs are from [Tabler Icons](https://github.com/tabler/tabler-icons)
by Paweł Kuna, under the MIT License.
