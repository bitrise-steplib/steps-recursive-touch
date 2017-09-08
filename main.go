package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bitrise-io/go-utils/fileutil"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-tools/go-steputils/cache"
	"github.com/bitrise-tools/go-steputils/input"
)

// ConfigsModel ...
type ConfigsModel struct {
	Path              string
	TouchTime         string
	TimestoreFilePath string
}

func createConfigsModelFromEnvs() ConfigsModel {
	return ConfigsModel{
		Path:              os.Getenv("directory_path"),
		TouchTime:         os.Getenv("touch_time"),
		TimestoreFilePath: os.Getenv("time_store_file_path"),
	}
}

func (configs ConfigsModel) print() {
	log.Infof("Configs:")
	log.Printf("- DirectoryPath: %s", configs.Path)
	log.Printf("- TouchTime: %s", configs.TouchTime)
	log.Printf("- TimestoreFilePath: %s", configs.TimestoreFilePath)
}

func (configs ConfigsModel) validate() error {
	if err := input.ValidateIfNotEmpty(configs.TimestoreFilePath); err != nil {
		return fmt.Errorf("TimestoreFilePath: %s", err)
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
	// Input validation
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
	touchTime := time.Now()

	if configs.TouchTime != "" {
		parsedTime, err := time.Parse(time.RFC3339, configs.TouchTime)
		if err != nil {
			log.Errorf("Failed to parse time(%s), error: %s", configs.TouchTime, err)
			os.Exit(1)
		}
		touchTime = parsedTime
	} else {
		timestoreFileExists, err := pathutil.IsPathExists(configs.TimestoreFilePath)
		if err != nil {
			log.Errorf("Failed to check if path(%s) exists, error: %s", configs.TimestoreFilePath, err)
			os.Exit(1)
		}

		if timestoreFileExists {
			log.Printf("- Timestore file found")
			timestoreFileContent, err := fileutil.ReadStringFromFile(configs.TimestoreFilePath)
			if err != nil {
				log.Errorf("Failed to read file content from(%s), error: %s", configs.TimestoreFilePath, err)
				os.Exit(1)
			}

			parsedTime, err := time.Parse(time.RFC3339, timestoreFileContent)
			if err != nil {
				log.Errorf("Failed to parse time(%s), error: %s", timestoreFileContent, err)
				os.Exit(1)
			}

			touchTime = parsedTime
		} else {
			log.Printf("- Timestore file created")
			if err := fileutil.WriteStringToFile(configs.TimestoreFilePath, touchTime.Format(time.RFC3339)); err != nil {
				log.Errorf("Failed to frite to file(%s), error: %s", configs.TimestoreFilePath, err)
				os.Exit(1)
			}
		}

		timestoreCache := cache.New()
		timestoreCache.IncludePath(configs.TimestoreFilePath)

		if err := timestoreCache.Commit(); err != nil {
			log.Warnf("Failed to commit cache paths, error: %s", err)
		}
	}

	log.Printf("- Using: %s", touchTime.Format(time.RFC3339))

	fmt.Println()

	log.Infof("Touch files...")

	filesCount := 0
	pathWalkStartTime := time.Now()

	err := filepath.Walk(configs.Path, func(path string, f os.FileInfo, err error) error {
		err = os.Chtimes(path, touchTime, touchTime)
		if err != nil {
			log.Warnf("Failed to touch file(%s), error: %s", path, err)
			return nil
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
