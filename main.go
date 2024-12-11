package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Config struct definition
type Config struct {
	FilePath   string   `json:"file_path"`
	VideoTypes []string `json:"video_types"`
	ProxyAddr  string   `json:"proxy_addr"`
}

// New function to handle config loading
func loadConfig(configFile string) (Config, error) {
	// Default config values
	defaultConfig := Config{
		FilePath:   "./",
		VideoTypes: []string{".mp4", ".mkv", ".avi"},
		ProxyAddr:  "",
	}

	configData, err := os.ReadFile(configFile)
	if err != nil {
		// Config file doesn't exist, create one with default values
		configData, err = json.MarshalIndent(defaultConfig, "", "    ")
		if err != nil {
			return Config{}, fmt.Errorf("error creating default config: %v", err)
		}

		err = os.WriteFile(configFile, configData, 0644)
		if err != nil {
			return Config{}, fmt.Errorf("error writing config file: %v", err)
		}

		fmt.Println("Created new config file with default values")
		return defaultConfig, nil
	}

	// Parse existing config file
	var config Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return Config{}, fmt.Errorf("error parsing config file: %v", err)
	}
	fmt.Println("Loaded existing config file")
	return config, nil
}

// extractMovieCode extracts the movie code from filename
func extractMovieCode(filename string) string {
	// Remove path, keep only filename
	base := filepath.Base(filename)

	// Remove common prefixes/suffixes and URLs
	// Common patterns: [XXX], (XXX), xxx-com, xxx.com
	cleaned := regexp.MustCompile(`\[.*?\]|\(.*?\)|[-_](com|net|org|xyz)[^.]*`).ReplaceAllString(base, "")

	// Extract movie code pattern (letters followed by numbers)
	// Including optional -c or -uc suffix (case insensitive)
	if matches := regexp.MustCompile(`(?i)([a-zA-Z]+-\d+(?:-(?:c|uc))?)`).FindString(cleaned); matches != "" {
		// Get extension from original filename
		ext := filepath.Ext(base)
		return strings.ToUpper(matches) + ext
	}

	return base
}

func getUniqueFilePath(targetPath string) string {
	if _, err := os.Stat(targetPath); err != nil {
		// 文件不存在，可以直接使用
		return targetPath
	}

	dir := filepath.Dir(targetPath)
	ext := filepath.Ext(targetPath)
	nameWithoutExt := strings.TrimSuffix(filepath.Base(targetPath), ext)

	counter := 1
	for {
		newPath := filepath.Join(dir, fmt.Sprintf("%s_%d%s", nameWithoutExt, counter, ext))
		if _, err := os.Stat(newPath); err != nil {
			// 找到一个不存在的文件名
			return newPath
		}
		counter++
	}
}

func main() {
	fmt.Println("ScrapeMovieData v0.0.0")
	fmt.Println("hello world")

	// Load config
	config, err := loadConfig("config.json")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	fmt.Printf("Using config: %+v\n", config)

	// Walk through the directory and find all video files
	var videoFiles []string
	err = filepath.Walk(config.FilePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip directories
		if info.IsDir() {
			return nil
		}
		// Check if file extension matches any of the video types
		for _, ext := range config.VideoTypes {
			if strings.HasSuffix(strings.ToLower(path), ext) {
				videoFiles = append(videoFiles, path)
				break
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		return
	}

	fmt.Printf("Found %d video files\n", len(videoFiles))
	for _, file := range videoFiles {
		movieCode := extractMovieCode(file)
		newPath := filepath.Join(filepath.Dir(file), movieCode)

		if file != newPath {
			uniquePath := getUniqueFilePath(newPath)
			err := os.Rename(file, uniquePath)
			if err != nil {
				fmt.Printf("Error renaming %s to %s: %v\n", file, uniquePath, err)
				continue
			}
			fmt.Printf("Renamed: %s -> %s\n", file, filepath.Base(uniquePath))
		} else {
			fmt.Printf("Skipped: %s (already named correctly)\n", file)
		}
	}
}
