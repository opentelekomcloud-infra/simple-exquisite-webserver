package main_test

import (
	"encoding/hex"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opentelekomcloud-infra/simple-exquisite-webserver/main"
)

/**
 * Helper functions
 */
func logErr(_ int, err error) {
	if err != nil {
		log.Printf("Read failed: %v", err)
	}
}

func validRandomPath() string {
	randBytes := make([]byte, 10)
	logErr(rand.Read(randBytes))
	return filepath.Join(os.TempDir(), "/"+hex.EncodeToString(randBytes)+".yml")
}

func invalidRandomPath() string {
	randBytes := make([]byte, 10)
	logErr(rand.Read(randBytes))
	return filepath.Join("/" + hex.EncodeToString(randBytes))
}

func errorOnDiff(expected interface{}, actual interface{}, t *testing.T) {
	diff := cmp.Diff(actual, expected)
	if diff != "" {
		t.Errorf("Actual and expected differs: \n%s", diff)
	}
}

/**
 * Test functions
 */
//WriteConfig test with invalid path
func TestConfiguration_WriteConfigErrorOnInvalidPath(t *testing.T) {
	var config main.Configuration
	var path = invalidRandomPath()
	// The following is the code under test
	err := config.WriteConfiguration(path)
	if err == nil {
		t.Errorf("No error on writing to invalid path")
	}
}

//LoadConfig test with invalid path
func TestConfiguration_LoadConfigErrorOnInvalidFilepath(t *testing.T) {
	var path = invalidRandomPath()
	_, err := main.LoadConfiguration(path)
	if err == nil {
		t.Errorf("No exception on reading from invalid path")
	}
}

var r = rand.New(rand.NewSource(0))

func strConfigTemplate(src *main.Configuration) []string {
	res := []string{
		fmt.Sprintf("debug: %v", src.Debug),
		fmt.Sprintf("server_port: %v", src.ServerPort),
	}
	if src.Postgres != nil {
		res = append(res,
			"postgres:",
			fmt.Sprintf("  db_url: %v", src.Postgres.DbURL),
			fmt.Sprintf("  database: %s", src.Postgres.Database),
			fmt.Sprintf("  username: %s", src.Postgres.Username),
			fmt.Sprintf("  password: %s", src.Postgres.Password),
		)
		if src.Postgres.Initial != nil {
			res = append(res,
				"  initial_data:",
				fmt.Sprintf("    count: %d", src.Postgres.Initial.Count),
				fmt.Sprintf("    size: %d", src.Postgres.Initial.Size),
			)
		}
	}
	return res
}

var simpleCS = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
var testDataSet = map[string]main.Configuration{
	"Without Postgres": {
		Debug:      true,
		ServerPort: r.Intn(0xffff),
	},
	"With Postgres": {
		Debug:      false,
		ServerPort: r.Intn(0xffff),
		Postgres: &main.PostgresConfig{
			DbURL:    fmt.Sprintf("localhost:%v", r.Intn(0xffff)),
			Database: main.RandomString(15, "", simpleCS),
			Username: main.RandomString(15, "", simpleCS),
			Password: main.RandomString(10, "", simpleCS),
		},
	},
	"With Postgres And Initial Data": {
		Debug:      false,
		ServerPort: r.Intn(0xffff),
		Postgres: &main.PostgresConfig{
			DbURL:    fmt.Sprintf("localhost:%v", r.Intn(0xffff)),
			Database: main.RandomString(15, "", simpleCS),
			Username: main.RandomString(15, "", simpleCS),
			Password: main.RandomString(10, "", simpleCS),
			Initial: &main.InitialData{
				Count: 100,
				Size:  20,
			},
		},
	},
}

func TestConfiguration_WriteConfigValidPathPg(t *testing.T) {
	for name, cfg := range testDataSet {
		t.Run(name, func(t *testing.T) {
			var path = validRandomPath()
			err := cfg.WriteConfiguration(path)
			if err != nil {
				t.Errorf("Can't write configuration")
			} else {
				t.Logf("Config written to %s", path)
			}

			targetFile, _ := os.Open(path)
			buffer, err := ioutil.ReadAll(targetFile)
			if err != nil {
				t.Errorf("Can't read configuration file")
			}
			strBuf := string(buffer)

			data := strings.Split(strings.TrimSpace(strBuf), "\n")

			expected := strConfigTemplate(&cfg)
			errorOnDiff(expected, data, t)
		})
	}
}
func TestConfiguration_LoadConfiguration(t *testing.T) {
	for name, cfg := range testDataSet {
		t.Run(name, func(t *testing.T) {
			expected := strConfigTemplate(&cfg)
			path := validRandomPath()
			file, err := os.Create(path)
			if err != nil {
				t.Errorf("Can't open configuration file")
				return
			}

			_, err = file.WriteString(strings.Join(expected, "\n"))
			if err != nil {
				t.Errorf("Can't write configuration file")
				return
			}

			res, err := main.LoadConfiguration(path)
			if err != nil {
				t.Errorf("Can't load configuration")
				return
			}
			errorOnDiff(cfg, *res, t)
		})
	}

}
