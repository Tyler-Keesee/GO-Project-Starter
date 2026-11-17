package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/briandowns/spinner"
	"github.com/ncruces/zenity"
)

type ProjectConfig struct {
	Name         string
	Language     string
	Author       string
	Description  string
	Path         string
	License      string
	GitInit      bool
	Dependencies []string
	Version      string
	TestsEnabled bool
	Docker       bool
}

type GitHubLicense struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Body string `json:"body"`
}

const (
	goTemplate = `package main

import (
    "fmt"
    "log"
    "os"
    "time"
)

// Version holds the current application version
var Version = "{{ .Version }}"

// Config holds application settings
type Config struct {
    LogLevel string
    Port     int
    Env      string
}

func init() {
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    log.Printf("Initializing {{ .Name }} v%s", Version)
}

func main() {
    startTime := time.Now()
    defer func() {
        log.Printf("Execution time: %v", time.Since(startTime))
    }()

    cfg := Config{
        LogLevel: os.Getenv("LOG_LEVEL"),
        Port:     8080,
        Env:      os.Getenv("APP_ENV"),
    }

    log.Printf("Starting {{ .Name }} with config: %+v", cfg)
    fmt.Println("{{ .Name }} is ready to rock! 🚀")
}
`

	pythonTemplate = `#!/usr/bin/env python3

import logging
import os
import sys
import time
from dataclasses import dataclass
from typing import Optional

VERSION = "{{ .Version }}"

@dataclass
class Config:
    """Application configuration."""
    log_level: str
    port: int
    env: str

def setup_logging() -> None:
    """Configure logging for the application."""
    logging.basicConfig(
        level=os.getenv("LOG_LEVEL", "INFO"),
        format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
    )

def main() -> None:
    """Main application entry point."""
    start_time = time.time()
    setup_logging()
    logger = logging.getLogger(__name__)
    
    config = Config(
        log_level=os.getenv("LOG_LEVEL", "INFO"),
        port=8080,
        env=os.getenv("APP_ENV", "development")
    )
    
    logger.info(f"Starting {{ .Name }} v{VERSION}")
    logger.info(f"Config: {config}")
    print("{{ .Name }} is ready to rock! 🚀")
    
    logger.info(f"Execution time: {time.time() - start_time:.2f}s")

if __name__ == "__main__":
    main()
`

	rustTemplate = `use log::{info, LevelFilter};
use env_logger;
use std::env;
use std::time::Instant;
use serde::{Serialize, Deserialize};

const VERSION: &str = "{{ .Version }}";

#[derive(Debug, Serialize, Deserialize)]
struct Config {
    log_level: String,
    port: u16,
    env: String,
}

fn main() {
    let start = Instant::now();
    
    env_logger::builder()
        .filter_level(LevelFilter::Info)
        .init();
    
    let config = Config {
        log_level: env::var("LOG_LEVEL").unwrap_or_else(|_| "INFO".to_string()),
        port: 8080,
        env: env::var("APP_ENV").unwrap_or_else(|_| "development".to_string()),
    };
    
    info!("Starting {{ .Name }} v{}", VERSION);
    info!("Config: {:?}", config);
    println!("{{ .Name }} is ready to rock! 🚀");
    
    info!("Execution time: {:?}", start.elapsed());
}
`

	javascriptTemplate = `const express = require('express');
const winston = require('winston');

const VERSION = "{{ .Version }}";
const app = express();
const port = process.env.PORT || 8080;

const logger = winston.createLogger({
    level: process.env.LOG_LEVEL || 'info',
    format: winston.format.json(),
    transports: [new winston.transports.Console()]
});

logger.info("Starting {{ .Name }} v" + VERSION);
console.log("{{ .Name }} is ready to rock! 🚀");

app.listen(port, () => {
    logger.info("Server running on port " + port);
});`

	typescriptTemplate = `import express from 'express';
import winston from 'winston';

const VERSION: string = "{{ .Version }}";
const app = express();
const port: number = parseInt(process.env.PORT || "8080");

const logger = winston.createLogger({
    level: process.env.LOG_LEVEL || 'info',
    format: winston.format.json(),
    transports: [new winston.transports.Console()]
});

logger.info("Starting {{ .Name }} v" + VERSION);
console.log("{{ .Name }} is ready to rock! 🚀");

app.listen(port, () => {
    logger.info("Server running on port " + port);
});`

	dockerfileTemplate = `FROM {{ .BaseImage }}

WORKDIR /app
COPY . .

{{ .BuildCommands }}

CMD ["{{ .EntryCommand }}"]
`

    htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ .Name }}</title>
    <link rel="stylesheet" href="style.css">
</head>
<body>
    <div class="container">
        <h1>{{ .Name }}</h1>
        <p>{{ .Description }}</p>
    </div>
    <script src="main.js"></script>
</body>
</html>`

    cssTemplate = `* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: Arial, sans-serif;
    line-height: 1.6;
    padding: 20px;
}

.container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 20px;
}

h1 {
    color: #333;
    margin-bottom: 20px;
    text-align: center;
}`
)

var languageConfigs = map[string]struct {
	Extension      string
	Template       string
	BaseImage      string
	BuildCommands  string
	EntryCommand   string
	Dependencies   []string
	CreateWebFiles bool
}{
	"go": {
		Extension:      ".go",
		Template:       goTemplate,
		BaseImage:      "golang:1.21-alpine",
		BuildCommands:  "RUN go build -o app",
		EntryCommand:   "./app",
		Dependencies:   []string{"github.com/sirupsen/logrus", "github.com/spf13/cobra"},
		CreateWebFiles: false,
	},
	"python": {
		Extension:      ".py",
		Template:       pythonTemplate,
		BaseImage:      "python:3.11-slim",
		BuildCommands:  "RUN pip install -r requirements.txt",
		EntryCommand:   "python main.py",
		Dependencies:   []string{"requests", "fastapi", "pytest"},
		CreateWebFiles: false,
	},
	"rust": {
		Extension:      ".rs",
		Template:       rustTemplate,
		BaseImage:      "rust:1.70",
		BuildCommands:  "RUN cargo build --release",
		EntryCommand:   "./target/release/app",
		Dependencies:   []string{"tokio", "serde", "clap"},
		CreateWebFiles: false,
	},
	"javascript": {
		Extension:      ".js",
		Template:       javascriptTemplate,
		BaseImage:      "node:18-alpine",
		BuildCommands:  "RUN npm install",
		EntryCommand:   "node main.js",
		Dependencies:   []string{"express", "dotenv", "winston"},
		CreateWebFiles: true,
	},
	"typescript": {
		Extension:      ".ts",
		Template:       typescriptTemplate,
		BaseImage:      "node:18-alpine",
		BuildCommands:  "RUN npm install && npm run build",
		EntryCommand:   "node dist/main.js",
		Dependencies:   []string{"typescript", "express", "@types/node"},
		CreateWebFiles: false,
	},
}

func selectLanguage() string {
	fmt.Println("\n📋 Available Languages:")
	languages := []string{"go", "python", "rust", "javascript", "typescript"}
	for i, lang := range languages {
		fmt.Printf("%d. %s\n", i+1, strings.Title(lang))
	}

	for {
		choice := getUserInput("Select language (1-5)", "1")
		index, err := strconv.Atoi(choice)
		if err == nil && index >= 1 && index <= len(languages) {
			return languages[index-1]
		}
		fmt.Printf("❌ Please select a valid option (1-%d)\n", len(languages))
	}
}

func fetchLicense(license string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/licenses/%s", license)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var licenseData GitHubLicense
	if err := json.NewDecoder(resp.Body).Decode(&licenseData); err != nil {
		return "", err
	}
	return licenseData.Body, nil
}

func createGitIgnore(projectDir string, language string) error {
	gitignoreContent := map[string]string{
		"go":         "/bin\n/vendor\n.env\n",
		"python":     "*.pyc\n__pycache__\nvenv/\n.env\n",
		"rust":       "/target\nCargo.lock\n.env\n",
		"javascript": "node_modules/\n.env\n",
		"typescript": "node_modules/\ndist/\n.env\n",
	}
	return os.WriteFile(filepath.Join(projectDir, ".gitignore"), []byte(gitignoreContent[language]), 0644)
}

func createDockerfile(config ProjectConfig, projectDir string) error {
	langConfig := languageConfigs[config.Language]
	dockerConfig := struct {
		BaseImage     string
		BuildCommands string
		EntryCommand  string
	}{
		BaseImage:     langConfig.BaseImage,
		BuildCommands: langConfig.BuildCommands,
		EntryCommand:  langConfig.EntryCommand,
	}

	tmpl, err := template.New("dockerfile").Parse(dockerfileTemplate)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(projectDir, "Dockerfile"))
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, dockerConfig)
}

func createProject(config ProjectConfig) error {
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Prefix = "Creating project "
	s.Start()
	defer s.Stop()

	projectDir := config.Path
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 5)

	wg.Add(1)
	go func() {
		defer wg.Done()
		readmeContent := fmt.Sprintf("# %s\n\n%s\n\nCreated by: %s\nVersion: %s\n",
			config.Name,
			config.Description,
			config.Author,
			config.Version,
		)
		if err := os.WriteFile(filepath.Join(projectDir, "README.md"), []byte(readmeContent), 0644); err != nil {
			errChan <- fmt.Errorf("failed to create README: %w", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		tmpl, err := template.New("source").Parse(languageConfigs[config.Language].Template)
		if err != nil {
			errChan <- fmt.Errorf("failed to parse template: %w", err)
			return
		}

		mainFile := filepath.Join(projectDir, "main"+languageConfigs[config.Language].Extension)
		f, err := os.Create(mainFile)
		if err != nil {
			errChan <- fmt.Errorf("failed to create main file: %w", err)
			return
		}
		defer f.Close()

		if err := tmpl.Execute(f, config); err != nil {
			errChan <- fmt.Errorf("failed to write template: %w", err)
		}
	}()

	if config.License != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			licenseContent, err := fetchLicense(config.License)
			if err != nil {
				errChan <- fmt.Errorf("failed to fetch license: %w", err)
				return
			}
			if err := os.WriteFile(filepath.Join(projectDir, "LICENSE"), []byte(licenseContent), 0644); err != nil {
				errChan <- fmt.Errorf("failed to create license file: %w", err)
			}
		}()
	}

	if config.Docker {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := createDockerfile(config, projectDir); err != nil {
				errChan <- fmt.Errorf("failed to create Dockerfile: %w", err)
			}
		}()
	}

	if config.GitInit {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := createGitIgnore(projectDir, config.Language); err != nil {
				errChan <- fmt.Errorf("failed to create .gitignore: %w", err)
				return
			}
			if err := exec.Command("git", "init", projectDir).Run(); err != nil {
				errChan <- fmt.Errorf("failed to initialize git: %w", err)
			}
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func getUserInput(prompt string, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [%s]: ", prompt, defaultValue)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}

func main() {
	fmt.Println("🚀 Welcome to Project Generator!")

	config := ProjectConfig{
		Version: "0.1.0",
	}

	for {
		config.Name = getUserInput("Project name", "awesome-project")
		if config.Name != "" {
			break
		}
		fmt.Println("❌ Project name cannot be empty")
	}

	config.Language = selectLanguage()

	defaultPath := filepath.Join(".", config.Name)
	var projectPath string
	if dir, err := zenity.SelectFile(
		zenity.Title("Select Project Location"),
		zenity.Directory(),
	); err != nil {
		projectPath = defaultPath
	} else {
		projectPath = filepath.Join(dir, config.Name)
	}
	config.Path = projectPath

	config.Author = getUserInput("Author", os.Getenv("USER"))
	config.Description = getUserInput("Description", "A fantastic new project")
	config.License = getUserInput("License (mit, apache-2.0, gpl-3.0)", "mit")

	gitInit := getUserInput("Initialize git repository? (y/n)", "y")
	config.GitInit = strings.ToLower(gitInit) == "y"

	docker := getUserInput("Generate Dockerfile? (y/n)", "n")
	config.Docker = strings.ToLower(docker) == "y"

	tests := getUserInput("Generate test files? (y/n)", "y")
	config.TestsEnabled = strings.ToLower(tests) == "y"

	s := spinner.New(spinner.CharSets[11], 100)
	s.Prefix = "Creating project "
	s.Start()

	if err := createProject(config); err != nil {
		s.Stop()
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	s.Stop()
	fmt.Printf("\n✨ Project created successfully!\n")
	fmt.Printf("📁 Location: %s\n", config.Path)
	fmt.Printf("🎉 Ready to code! CD into your project:\n")
	fmt.Printf("   cd %s\n", config.Path)
}
