package main

import (
	"errors"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

// Quartz event services, <CoreGraphics/CGEventTypes.h>.
const (
	hidEventTap        = 0 // kCGHIDEventTap: events as they enter the window server
	headInsertEventTap = 0 // kCGHeadInsertEventTap
	tapOptionDefault   = 0 // kCGEventTapOptionDefault: a tap that may drop events

	eventKeyDown       = 10
	eventKeyUp         = 11
	eventFlagsChanged  = 12 // modifier keys
	eventSystemDefined = 14 // NX_SYSDEFINED: among others, the media keys

	eventTapDisabledByTimeout   = 0xFFFFFFFE
	eventTapDisabledByUserInput = 0xFFFFFFFF

	// NX_SUBTYPE_AUX_CONTROL_BUTTONS: the system defined events of the
	// volume, brightness, playback and other special keys.
	subtypeAuxControlButtons = 8

	hidParamConnectType = 1 // kIOHIDParamConnectType
	hidCapsLockState    = 1 // kIOHIDCapsLockState
)

var (
	cgEventTapCreate func(tap, place, options uint32, events uint64, callback, userInfo uintptr) uintptr
	cgEventTapEnable func(tap uintptr, enable bool)

	cfMachPortCreateRunLoopSource func(alloc, port uintptr, order int) uintptr
	cfMachPortInvalidate          func(port uintptr)
	cfRunLoopGetMain              func() uintptr
	cfRunLoopAddSource            func(rl, src, mode uintptr)
	cfRunLoopRemoveSource         func(rl, src, mode uintptr)
	cfDictionaryCreate            func(alloc uintptr, keys, values *uintptr, n int, keyCallBacks, valueCallBacks uintptr) uintptr
	cfRelease                     func(obj uintptr)

	axIsProcessTrusted            func() bool
	axIsProcessTrustedWithOptions func(options uintptr) bool

	ioServiceMatching           func(name string) uintptr
	ioServiceGetMatchingService func(mainPort uint32, matching uintptr) uint32
	ioServiceOpen               func(service, owningTask, typ uint32, connect *uint32) int32
	ioServiceClose              func(connect uint32) int32
	ioObjectRelease             func(object uint32) int32
	ioHIDGetModifierLockState   func(connect uint32, selector int32, state *bool) int32
	ioHIDSetModifierLockState   func(connect uint32, selector int32, state bool) int32

	kCFRunLoopCommonModes           uintptr
	kCFBooleanTrue                  uintptr
	kCFTypeDictionaryKeyCallBacks   uintptr
	kCFTypeDictionaryValueCallBacks uintptr
	kAXTrustedCheckOptionPrompt     uintptr
	machTaskSelf                    uint32

	selEventWithCGEvent, selSubtype objc.SEL

	loadOnce sync.Once
)

func load() {
	open := func(path string) uintptr {
		lib, err := purego.Dlopen(path, purego.RTLD_GLOBAL|purego.RTLD_NOW)
		if err != nil {
			panic(err)
		}
		return lib
	}
	sym := func(lib uintptr, name string) uintptr {
		p, err := purego.Dlsym(lib, name)
		if err != nil {
			panic(err)
		}
		return p
	}
	// value reads the pointer a constant such as kCFBooleanTrue holds.
	value := func(lib uintptr, name string) uintptr {
		p := sym(lib, name)
		return **(**uintptr)(unsafe.Pointer(&p))
	}

	cg := open("/System/Library/Frameworks/CoreGraphics.framework/CoreGraphics")
	purego.RegisterLibFunc(&cgEventTapCreate, cg, "CGEventTapCreate")
	purego.RegisterLibFunc(&cgEventTapEnable, cg, "CGEventTapEnable")

	cf := open("/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation")
	purego.RegisterLibFunc(&cfMachPortCreateRunLoopSource, cf, "CFMachPortCreateRunLoopSource")
	purego.RegisterLibFunc(&cfMachPortInvalidate, cf, "CFMachPortInvalidate")
	purego.RegisterLibFunc(&cfRunLoopGetMain, cf, "CFRunLoopGetMain")
	purego.RegisterLibFunc(&cfRunLoopAddSource, cf, "CFRunLoopAddSource")
	purego.RegisterLibFunc(&cfRunLoopRemoveSource, cf, "CFRunLoopRemoveSource")
	purego.RegisterLibFunc(&cfDictionaryCreate, cf, "CFDictionaryCreate")
	purego.RegisterLibFunc(&cfRelease, cf, "CFRelease")
	kCFRunLoopCommonModes = value(cf, "kCFRunLoopCommonModes")
	kCFBooleanTrue = value(cf, "kCFBooleanTrue")
	kCFTypeDictionaryKeyCallBacks = sym(cf, "kCFTypeDictionaryKeyCallBacks")
	kCFTypeDictionaryValueCallBacks = sym(cf, "kCFTypeDictionaryValueCallBacks")

	as := open("/System/Library/Frameworks/ApplicationServices.framework/ApplicationServices")
	purego.RegisterLibFunc(&axIsProcessTrusted, as, "AXIsProcessTrusted")
	purego.RegisterLibFunc(&axIsProcessTrustedWithOptions, as, "AXIsProcessTrustedWithOptions")
	kAXTrustedCheckOptionPrompt = value(as, "kAXTrustedCheckOptionPrompt")

	iokit := open("/System/Library/Frameworks/IOKit.framework/IOKit")
	purego.RegisterLibFunc(&ioServiceMatching, iokit, "IOServiceMatching")
	purego.RegisterLibFunc(&ioServiceGetMatchingService, iokit, "IOServiceGetMatchingService")
	purego.RegisterLibFunc(&ioServiceOpen, iokit, "IOServiceOpen")
	purego.RegisterLibFunc(&ioServiceClose, iokit, "IOServiceClose")
	purego.RegisterLibFunc(&ioObjectRelease, iokit, "IOObjectRelease")
	purego.RegisterLibFunc(&ioHIDGetModifierLockState, iokit, "IOHIDGetModifierLockState")
	purego.RegisterLibFunc(&ioHIDSetModifierLockState, iokit, "IOHIDSetModifierLockState")

	// mach_task_self() is a macro for this variable.
	task := sym(open("/usr/lib/libSystem.B.dylib"), "mach_task_self_")
	machTaskSelf = **(**uint32)(unsafe.Pointer(&task))

	// AppKit, which MyGo has loaded, decodes system defined events.
	selEventWithCGEvent = objc.RegisterName("eventWithCGEvent:")
	selSubtype = objc.RegisterName("subtype")
}

// eventTap disables the keyboard with a Quartz event tap: a filter the
// window server passes every key event through before any app, itself
// included, gets it. The tap drops them all, and never sees the pointer's
// events. It lives only while the keyboard is disabled, and the system
// removes it should the app exit or crash, so the keys cannot stay off.
type eventTap struct {
	port, source uintptr

	// Caps Lock toggles below the tap, so cleaning may leave it on: Enable
	// puts it back as Disable found it.
	hid      uint32
	capsLock bool
}

func newKeyboard() keyboard { return &eventTap{} }

// activeTap is the tap tapCallback serves. Both are used on the main
// thread alone, whose run loop delivers the events.
var activeTap *eventTap

var tapCallback = purego.NewCallback(func(proxy, typ, event, userInfo uintptr) uintptr {
	t := activeTap
	if t == nil {
		return event
	}
	switch uint32(typ) {
	case eventTapDisabledByTimeout, eventTapDisabledByUserInput:
		// The system turns off a tap that was slow to answer.
		cgEventTapEnable(t.port, true)
		return event
	case eventSystemDefined:
		// Only those of the special keys: the others are not the keyboard's.
		ev := objc.ID(objc.GetClass("NSEvent")).Send(selEventWithCGEvent, event)
		if ev == 0 || objc.Send[int16](ev, selSubtype) != subtypeAuxControlButtons {
			return event
		}
	}
	return 0 // drops the event
})

func (t *eventTap) Disable() error {
	loadOnce.Do(load)
	if t.port != 0 {
		return nil
	}
	if !axIsProcessTrusted() {
		// Adds the app to the list in System Settings and has the system
		// offer to open it.
		keys, values := kAXTrustedCheckOptionPrompt, kCFBooleanTrue
		options := cfDictionaryCreate(0, &keys, &values, 1, kCFTypeDictionaryKeyCallBacks, kCFTypeDictionaryValueCallBacks)
		axIsProcessTrustedWithOptions(options)
		cfRelease(options)
		return errNoAccess
	}
	events := uint64(1<<eventKeyDown | 1<<eventKeyUp | 1<<eventFlagsChanged | 1<<eventSystemDefined)
	port := cgEventTapCreate(hidEventTap, headInsertEventTap, tapOptionDefault, events, tapCallback, 0)
	if port == 0 {
		return errors.New("macOS refused to filter the keyboard's events")
	}
	t.port = port
	t.source = cfMachPortCreateRunLoopSource(0, port, 0)
	activeTap = t
	cfRunLoopAddSource(cfRunLoopGetMain(), t.source, kCFRunLoopCommonModes)
	cgEventTapEnable(port, true)

	if service := ioServiceGetMatchingService(0, ioServiceMatching("IOHIDSystem")); service != 0 {
		if ioServiceOpen(service, machTaskSelf, hidParamConnectType, &t.hid) != 0 ||
			ioHIDGetModifierLockState(t.hid, hidCapsLockState, &t.capsLock) != 0 {
			t.closeHID()
		}
		ioObjectRelease(service)
	}
	return nil
}

func (t *eventTap) Enable() {
	if t.port == 0 {
		return
	}
	cgEventTapEnable(t.port, false)
	cfRunLoopRemoveSource(cfRunLoopGetMain(), t.source, kCFRunLoopCommonModes)
	cfMachPortInvalidate(t.port)
	cfRelease(t.source)
	cfRelease(t.port)
	activeTap = nil
	t.port, t.source = 0, 0

	if t.hid != 0 {
		var on bool
		if ioHIDGetModifierLockState(t.hid, hidCapsLockState, &on) == 0 && on != t.capsLock {
			ioHIDSetModifierLockState(t.hid, hidCapsLockState, t.capsLock)
		}
		t.closeHID()
	}
}

func (t *eventTap) closeHID() {
	if t.hid != 0 {
		ioServiceClose(t.hid)
		t.hid = 0
	}
}
