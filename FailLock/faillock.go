package faillock

import (
	config "Fly2Stats/Config"
	"errors"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

const LockFileName = "fail.lock"

var LockDir string
var LockFilePath string

func Lockup(err error) error {
	file, createErr := os.Create(LockFilePath)
	if createErr != nil {
		log.WithError(createErr).Error("error while creating lock file")
		return createErr
	}
	defer file.Close()
	file.WriteString(err.Error())
	return nil
}
func LockAndPan(err error) {
	Lockup(err)
	log.WithError(err).Panic("i'm terified ...")
}
func Check() {
	if _, err := os.Stat(LockFilePath); !os.IsNotExist(err) {
		if err == nil {
			errContent, _ := os.ReadFile(LockFilePath)
			log.WithError(errors.New(string(errContent))).Panic("a panic file exists")
		} else {
			log.WithError(err).Panic("error while checking existing panic file")
		}
	}
}
func TestUp() {
	err := Lockup(errors.New("just checking in"))
	if err != nil {
		log.WithError(err).Panic("error while creating check panic file")
	}
	err = os.Remove(LockFilePath)
	if err != nil {
		log.WithError(err).Panic("can not remove check panic file")
	}
}
func init() {
	cfg := config.Config()
	LockDir = cfg.CheckpointPath.AsString()
	LockFilePath = filepath.Join(LockDir, LockFileName)
	Check()
}
