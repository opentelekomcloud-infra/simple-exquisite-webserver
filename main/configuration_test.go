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
func TestWriteConfigErrorOnInvalidPath(t *testing.T) {
	var config main.Configuration
	var path = invalidRandomPath()
	// The following is the code under test
	err := config.WriteConfiguration(path)
	if err == nil {
		t.Errorf("No error on writing to invalid path")
	}
}

//LoadConfig test with invalid path
func TestLoadConfigErrorOnInvalidFilepath(t *testing.T) {
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
			"postgres:", fmt.Sprintf("  db_url: %v", src.Postgres.DbURL),
			fmt.Sprintf("  database: %s", src.Postgres.Database),
			fmt.Sprintf("  username: %s", src.Postgres.Username),
			fmt.Sprintf("  password: %s", src.Postgres.Password),
		)
	}
	return res
}

func TestWriteConfigValidPathPg(t *testing.T) {
	src := main.Configuration{
		Debug:      1 == r.Intn(1),
		ServerPort: r.Intn(0xffff),
		Postgres: &main.PostgresConfig{
			DbURL:    fmt.Sprintf("localhost:%v", r.Intn(0xffff)),
			Database: "edlkjsfd",
			Username: "sfdjnsfdjlkjsfd",
			Password: "opoxgdp[koiujiklililhkjg",
		},
	}
	var path = validRandomPath()
	err := src.WriteConfiguration(path)
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

	expected := strConfigTemplate(&src)
	errorOnDiff(expected, data, t)
}

func TestLoadConfigValidPathPg(t *testing.T) {
	src := main.Configuration{
		Debug:      1 == r.Intn(1),
		ServerPort: r.Intn(0xffff),
		Postgres: &main.PostgresConfig{
			DbURL:    fmt.Sprintf("localhost:%v", r.Intn(0xffff)),
			Database: "edlkjsfd",
			Username: "sfdjnsfdjlkjsfd",
			Password: "opoxgdp[koiujiklililhkjg",
		},
	}
	expected := strConfigTemplate(&src)
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
	errorOnDiff(src, *res, t)
}
