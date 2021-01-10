package powerusage

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/womat/debug"
)

const httpRequestTimeout = 10 * time.Second

const (
	Export State = "export"
	Import State = "import"
)

type State string

type Measurements struct {
	sync.RWMutex
	Timestamp      time.Time
	State          State
	Power          float64
	PowerFromGrid  float64
	EnergyToGrid   float64
	EnergyFromGrid float64
	config         struct {
		powerMeterURL    string
		inverterMeterURL string
	}
}

type meterURLBody struct {
	Timestamp time.Time `json:"Time"`
	Runtime   float64   `json:"Runtime"`
	Measurand struct {
		E         float64 `json:"e"`
		EfromGrid float64 `json:"e_grid"`
		P         float64 `json:"p"`
	} `json:"Measurand"`
}

func New() *Measurements {
	return &Measurements{}
}

func (m *Measurements) SetInverterMeterURL(url string) {
	m.config.inverterMeterURL = url
}

func (m *Measurements) SetPowerMeterURL(url string) {
	m.config.powerMeterURL = url
}

func (m *Measurements) Read() (err error) {
	var wg sync.WaitGroup

	data := New()
	data.SetInverterMeterURL(m.config.inverterMeterURL)
	data.SetPowerMeterURL(m.config.powerMeterURL)

	start := time.Now()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if e := data.readInverterMeter(); e != nil {
			err = e
		}

		debug.TraceLog.Printf("runtime to request inverter meter data: %vs", time.Since(start).Seconds())
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if e := data.readPowerMeter(); e != nil {
			err = e
		}

		debug.TraceLog.Printf("runtime to request primary meter data: %vs", time.Since(start).Seconds())
	}()

	wg.Wait()
	debug.DebugLog.Printf("runtime to request data: %vs", time.Since(start).Seconds())

	m.Lock()
	defer m.Unlock()
	if data.PowerFromGrid > 0 {
		m.State = Import
	} else {
		m.State = Export
	}
	m.Power = data.PowerFromGrid + data.Power
	m.PowerFromGrid = data.PowerFromGrid
	m.EnergyFromGrid = data.EnergyFromGrid
	m.EnergyToGrid = data.EnergyToGrid
	m.Timestamp = time.Now()
	return
}

func (m *Measurements) readInverterMeter() (err error) {
	var r meterURLBody

	if err = read(m.config.inverterMeterURL, &r); err != nil {
		return
	}

	m.Lock()
	defer m.Unlock()

	m.Power = r.Measurand.P
	return
}

func (m *Measurements) readPowerMeter() (err error) {
	var r meterURLBody

	if err = read(m.config.powerMeterURL, &r); err != nil {
		return
	}

	m.Lock()
	defer m.Unlock()

	m.PowerFromGrid = r.Measurand.P
	m.EnergyToGrid = r.Measurand.E
	m.EnergyFromGrid = r.Measurand.EfromGrid

	return
}

func read(url string, data interface{}) (err error) {
	done := make(chan bool, 1)
	go func() {
		// ensures that data is sent to the channel when the function is terminated
		defer func() {
			select {
			case done <- true:
			default:
			}
			close(done)
		}()

		debug.TraceLog.Printf("performing http get: %v\n", url)

		var resp *http.Response
		if resp, err = http.Get(url); err != nil {
			return
		}

		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if err = json.Unmarshal(bodyBytes, data); err != nil {
			return
		}
	}()

	// wait for API Data
	select {
	case <-done:
	case <-time.After(httpRequestTimeout):
		err = errors.New("timeout during receive data")
	}

	if err != nil {
		debug.ErrorLog.Println(err)
		return
	}
	return
}
