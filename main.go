package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-steputils/input"
)

// ConfigsModel ...
type ConfigsModel struct {
	Path      string
	TouchTime string
}

func createConfigsModelFromEnvs() ConfigsModel {
	return ConfigsModel{
		Path:      os.Getenv("directory_path"),
		TouchTime: os.Getenv("touch_time"),
	}
}

func (configs ConfigsModel) print() {
	log.Infof("Configs:")
	log.Printf("- DirectoryPath: %s", configs.Path)
	log.Printf("- TouchTime: %s", configs.TouchTime)
}

func (configs ConfigsModel) validate() error {
	if err := input.ValidateIfNotEmpty(configs.TouchTime); err != nil {
		return fmt.Errorf("TouchTime: %s", err)
	}
	if err := input.ValidateIfNotEmpty(configs.Path); err != nil {
		return fmt.Errorf("DirectoryPath: %s", err)
	}
	if err := input.ValidateIfDirExists(configs.Path); err != nil {
		return fmt.Errorf("DirectoryPath: %s", err)
	}

	return nil
}

func main() {
	configs := createConfigsModelFromEnvs()

	fmt.Println()
	configs.print()

	if err := configs.validate(); err != nil {
		log.Errorf("Issue with input: %s", err)
		os.Exit(1)
	}

	fmt.Println()

	//
	// Main
	log.Infof("Check time...")

	parsedTime, err := time.Parse(time.RFC3339, configs.TouchTime)
	if err != nil {
		log.Errorf("Failed to parse time(%s), error: %s", configs.TouchTime, err)
		os.Exit(1)
	}
	touchTime := parsedTime

	log.Printf("- Using: %s", touchTime.Format(time.RFC3339))

	fmt.Println()
	log.Infof("Touch files...")

	filesCount := 0
	pathWalkStartTime := time.Now()

	err = filepath.Walk(configs.Path, func(path string, f os.FileInfo, err error) error {
		if f.Mode()&os.ModeSymlink != 0 {
			time := f.ModTime().Format("0601021504.05")
			if err := command.New("touch", "-ht", time, path).Run(); err != nil {
				log.Warnf("Failed to touch file(%s), error: %s", path, err)
				return nil
			}
		} else {
			err = os.Chtimes(path, touchTime, touchTime)
			if err != nil {
				log.Warnf("Failed to touch file(%s), error: %s", path, err)
				return nil
			}
		}

		filesCount++
		return nil
	})
	if err != nil {
		log.Errorf("Failed to walk directory(%s), error: %s", configs.Path, err)
		os.Exit(1)
	}

	log.Printf("- %d files touched in: %s", filesCount, time.Now().Sub(pathWalkStartTime))
}
