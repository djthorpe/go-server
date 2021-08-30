package chromecast

import (
	"fmt"
	"strconv"

	// Modules
	. "github.com/djthorpe/go-server"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Cast struct {
	id, fn string
	md, rs string
	st     uint

	host string
	port uint16
	//vol *Volume
	//app *App
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewCast(instance Service) *Cast {
	this := new(Cast)

	// Set host and port for connection
	this.host = instance.Host()
	this.port = instance.Port()

	// Set Chromecast ID
	if id := instance.ValueForKey("id"); id == "" {
		return nil
	} else {
		this.id = id
	}

	// Set chromecast name
	if fn := instance.ValueForKey("fn"); fn == "" {
		this.fn = "Chromecast"
	} else {
		this.fn = fn
	}

	// Model, application, state
	this.md = instance.ValueForKey("md")
	this.rs = instance.ValueForKey("rs")
	if st, err := strconv.ParseUint(instance.ValueForKey("st"), 0, 64); err == nil {
		this.st = uint(st)
	}

	return this
}

func (this *Cast) connect() error {
	return nil
}

func (this *Cast) disconnect() error {
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Cast) String() string {
	str := "<cast"
	if this.host != "" && this.port > 0 {
		str += fmt.Sprintf(" addr=%s:%v", this.host, this.port)
	}
	if name := this.fn; name != "" {
		str += fmt.Sprintf(" name=%q", name)
	}
	if model := this.md; model != "" {
		str += fmt.Sprintf(" model=%q", model)
	}
	if app := this.rs; app != "" {
		str += fmt.Sprintf(" app=%q", app)
	}
	str += fmt.Sprintf(" state=%v", this.st)
	return str + ">"
}

/*
////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *Cast) Id() string {
	return this.id
}

// Name returns the readable name for a chromecast
func (this *Cast) Name() string {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()
	return this.fn
}

// Model returns the reported model information
func (this *Cast) Model() string {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()
	return this.md
}

func (this *Cast) service() string {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()
	return this.rs
}

// Service returns the currently running service
func (this *Cast) Service() string {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()
	if this.app != nil && this.app.DisplayName != "" {
		return this.app.DisplayName
	} else {
		return this.rs
	}
}

// State returns 0 if backdrop (no app running), else returns 1
func (this *Cast) State() uint {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()
	return this.st
}

// Return volume or nil if volume is not known
func (this *Cast) volume() *Volume {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()
	if this.vol == nil {
		return nil
	} else {
		// Make a copy of volume
		vol := *this.vol
		return &vol
	}
}

// Return volume or (0,false) if volume is not known
func (this *Cast) Volume() (float32, bool) {
	if vol := this.volume(); vol == nil {
		return 0, false
	} else if vol.Level == 0.0 {
		return 0, true
	} else {
		return vol.Level, vol.Muted
	}
}

func (this *Cast) App() *App {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()
	if this.app == nil {
		return nil
	} else {
		// Make a copy of app
		app := *this.app
		return &app
	}
}

////////////////////////////////////////////////////////////////////////////////
// METHODS

func (this *Cast) Equals(other *Cast) gopi.CastFlag {
	flags := gopi.CAST_FLAG_NONE
	if other == nil {
		return flags
	}
	// Any change to Name, Model or Id
	if this.Id() != other.Id() || this.Name() != other.Name() || this.Model() != other.Model() {
		flags |= gopi.CAST_FLAG_NAME
	}
	// Any change to service or state
	if this.service() != other.service() {
		fmt.Println(" service changed")
		flags |= gopi.CAST_FLAG_APP
	}
	if this.State() != other.State() {
		fmt.Println(" state changed")
		flags |= gopi.CAST_FLAG_APP
	}
	// Return changed flags
	return flags
}

func (this *Cast) ConnectWithTimeout(ch gopi.Publisher, timeout time.Duration) (*Conn, error) {
	// Use hostname to connect
	addr := fmt.Sprintf("%v:%v", this.host, this.port)

	// Update state
	this.vol = nil
	this.app = nil

	// Perform the connection
	return NewConnWithTimeout(this.id, addr, timeout, ch)
}

func (this *Cast) Disconnect(conn *Conn) error {
	// Update state
	this.vol = nil
	this.app = nil

	// Close connection
	return conn.Close()
}

func (this *Cast) UpdateState(state *State) gopi.CastFlag {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()

	fmt.Println("=>updateState")

	if state.apps != nil {
		return this.updateStateApps(state)
	} else if state.media != nil {
		return this.updateStateMedia(state)
	} else {
		return gopi.CAST_FLAG_NONE
	}
}

func (this *Cast) updateStateApps(state *State) gopi.CastFlag {
	// Changes in volume
	flags := gopi.CAST_FLAG_NONE
	if this.vol == nil || state.volume.Equals(*this.vol) == false {
		this.vol = &state.volume
		flags |= gopi.CAST_FLAG_VOLUME
	}

	// Changes in app
	if this.app == nil && len(state.apps) > 0 {
		this.app = &state.apps[0]
		flags |= gopi.CAST_FLAG_APP
	} else if len(state.apps) == 0 && this.app != nil {
		this.app = nil
		flags |= gopi.CAST_FLAG_APP
	} else if this.app != nil && len(state.apps) > 0 && this.app.Equals(state.apps[0]) == false {
		this.app = &state.apps[0]
		flags |= gopi.CAST_FLAG_APP
	}

	fmt.Println("   => updateStateApps", flags, this.app)

	// Return any changed state
	return flags
}

func (this *Cast) updateStateMedia(state *State) gopi.CastFlag {
	fmt.Println("   => updateStateMedia", state.media)
	flags := gopi.CAST_FLAG_NONE
	return flags
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *Cast) updateFrom(other *Cast) {
	this.RWMutex.Lock()
	other.RWMutex.RLock()
	defer this.RWMutex.Unlock()
	defer other.RWMutex.RUnlock()

	// This seems dumb
	this.id = other.id
	this.fn = other.fn
	this.md = other.md
	this.rs = other.rs
	this.st = other.st
	this.host = other.host
	this.ips = other.ips
	this.port = other.port
}

func txtToMap(txt []string) map[string]string {
	result := make(map[string]string, len(txt))
	for _, r := range txt {
		if kv := strings.SplitN(r, "=", 2); len(kv) == 2 {
			result[kv[0]] = kv[1]
		} else if len(kv) == 1 {
			result[kv[0]] = ""
		}
	}
	return result
}
	if app := this.rs; app != "" {
		str += fmt.Sprintf(" app=%q", app)
	}
*/
