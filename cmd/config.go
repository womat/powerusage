package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/womat/debug"

	"powerusage/global"
	"powerusage/pkg/config"
)

// defaultInterval defines the default of dataCollectionInterval and backupInterval (in seconds)
const (
	defaultDataCollectionInterval = 60
	defaultPort                   = 4000
	defaultDataFile               = "/opt/womat/data/" + global.MODULE + ".yaml"
	defaultDebugFile              = "stderr"
	defaultDebugFlag              = "standard"
)

type yamlDebug struct {
	File string `yaml:"file"`
	Flag string `yaml:"flag"`
}

type yamlWebserver struct {
	Port        int             `yaml:"port"`
	Webservices map[string]bool `yaml:"webservices"`
}

type yamlStruct struct {
	DataCollectionInterval int           `yaml:"datacollectioninterval"`
	Debug                  yamlDebug     `yaml:"debug"`
	WebServer              yamlWebserver `yaml:"webserver"`
	InverterMeterURL       string        `yaml:"invertermeterurl"`
	PowerMeterURL          string        `yaml:"powermeterurl"`
}

func init() {
	var err error
	var flags = config.Flag{
		"version":   {FlagType: config.FlagBool, Usage: "print version and exit", DefaultValue: false},
		"debugFile": {FlagType: config.FlagString, Usage: "log file eg. " + filepath.Join("/opt/womat/log/"+global.MODULE+".log") + `(default "stderr")`, DefaultValue: ""},
		"debugFlag": {FlagType: config.FlagString, Usage: `"enable debug information (standard | trace | debug) (default "standard")`, DefaultValue: ""},
		"config":    {FlagType: config.FlagString, Usage: "Config File", DefaultValue: filepath.Join("/opt/womat/conf/" + global.MODULE + ".yaml")},
	}
	var configFile = yamlStruct{
		DataCollectionInterval: defaultDataCollectionInterval,
		Debug:                  yamlDebug{File: defaultDebugFile, Flag: defaultDebugFlag},
		WebServer:              yamlWebserver{Port: defaultPort, Webservices: map[string]bool{"version": false, "currentdata": false}},
	}

	config.Parse(flags)
	if flags.Bool("version") {
		fmt.Printf("Version: %v\n", global.VERSION)
		os.Exit(0)
	}

	if err := readConfigFile(flags.String("config"), &configFile); err != nil {
		log.Fatalf("Error reading config file, %v", err)
	}
	if global.Config.Debug.File, global.Config.Debug.Flag, err = getDebugConfig(flags, configFile.Debug); err != nil {
		log.Fatalf("unable to open debug file, %v", err)
	}
	global.Config.DataCollectionInterval = time.Duration(configFile.DataCollectionInterval) * time.Second
	global.Config.InverterMeterURL = configFile.InverterMeterURL
	global.Config.PowerMeterURL = configFile.PowerMeterURL
	global.Config.Webserver.Port = configFile.WebServer.Port
	for s, b := range configFile.WebServer.Webservices {
		global.Config.Webserver.Webservices[s] = b
	}
}

func getDebugConfig(f config.Flag, d yamlDebug) (writer io.WriteCloser, flag int, err error) {
	var fileName, flagString string

	if s := f.String("debugFile"); s != "" {
		fileName = s
	} else {
		fileName = d.File
	}

	if s := f.String("debugFlag"); s != "" {
		flagString = s
	} else {
		flagString = d.Flag
	}

	// defines Debug section of global.Config
	switch flagString {
	case "trace":
		flag = debug.Full
	case "debug":
		flag = debug.Warning | debug.Info | debug.Error | debug.Fatal | debug.Debug
	case "standard":
		flag = debug.Standard
	}

	switch fileName {
	case "stderr":
		writer = os.Stderr
	case "stdout":
		writer = os.Stdout
	default:
		if writer, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
			return
		}
	}

	return
}

func readConfigFile(fileName string, c *yamlStruct) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err = decoder.Decode(c); err != nil {
		return err
	}

	return nil
}
