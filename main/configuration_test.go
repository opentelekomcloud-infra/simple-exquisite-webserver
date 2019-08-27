package main_test

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/outcatcher/simple-exquisite-webserver/main"
)

func logerr(n int, err error) {
	if err != nil {
		log.Printf("Write failed: %v", err)
	}
}

func validRandomPath() string {
	randBytes := make([]byte, 10)
	logerr(rand.Read(randBytes))
	return filepath.Join(os.TempDir(), "/"+hex.EncodeToString(randBytes)+".yml")
}

func invalidRandomPath() string {
	randBytes := make([]byte, 10)
	logerr(rand.Read(randBytes))
	return filepath.Join("/" + hex.EncodeToString(randBytes))
}

//WriteConfig test with invalid path
func TestWriteConfigPanicOnInvalidPath(t *testing.T) {
	var config main.Configuration
	var path = invalidRandomPath()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	// The following is the code under test
	config.WriteConfiguration(path, true)
}

//LoadConfig test with invalid path
func TestLoadConfigPanicOnInvalidFilepath(t *testing.T) {
	var config main.Configuration
	var path = invalidRandomPath()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	// The following is the code under test
	config.LoadConfiguration(path)
}

//Load debug=true config test with valid path
func TestWriteAndLoadDebugConfigValidPath(t *testing.T) {
	g := NewGomegaWithT(t)
	var config main.Configuration
	var path = validRandomPath()

	config.WriteConfiguration(path, true)
	config.LoadConfiguration(path)

	g.Expect(config.Debug).To(Equal(true))
}

//Load debug=false config test with valid path
func TestWriteAndLoadConfigValidPath(t *testing.T) {
	g := NewGomegaWithT(t)
	var config main.Configuration
	var path = validRandomPath()

	config.WriteConfiguration(path, false)
	config.LoadConfiguration(path)

	g.Expect(config.Debug).To(Equal(false))
	g.Expect(config.PgDatabase).To(Equal("entities"))
	g.Expect(config.PgDbURL).To(Equal("localhost:9999"))
	g.Expect(config.PgUsername).To(Equal("entities"))
	g.Expect(config.PgPassword).To(Equal(""))
}
