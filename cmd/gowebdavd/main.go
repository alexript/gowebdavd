// Copyright (c) 2026 gowebdavd contributors
// SPDX-License-Identifier: MIT

package main

import (
	"flag"
	"fmt"
	"os"

	"gowebdavd/internal/daemon"
	"gowebdavd/internal/logger"
	"gowebdavd/internal/pidfile"
	"gowebdavd/internal/process"
	"gowebdavd/internal/server"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "start", "run":
		handleStartOrRun(command)

	case "stop":
		handleStop()

	case "status":
		handleStatus()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: gowebdavd <start|stop|status|run> [options]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  start   - Start WebDAV server in background")
	fmt.Println("  stop    - Stop WebDAV server")
	fmt.Println("  status  - Show service status")
	fmt.Println("  run     - Run WebDAV server in foreground")
	fmt.Println("")
	fmt.Println("Options for start/run:")
	fmt.Println("  -dir string    Directory to serve (default \".\")")
	fmt.Println("  -port int      Port to listen on (default 8080)")
	fmt.Println("  -bind string   IP address to bind to (default \"127.0.0.1\")")
	fmt.Println("  -log           Enable HTTP request logging (default: false)")
	fmt.Println("  -log-dir       Custom log directory (requires -log, must exist)")
}

func handleStartOrRun(command string) {
	startCmd := flag.NewFlagSet("start", flag.ExitOnError)
	folder := startCmd.String("dir", ".", "Directory")
	port := startCmd.Int("port", 8080, "Port")
	bind := startCmd.String("bind", "127.0.0.1", "IP")
	enableLog := startCmd.Bool("log", false, "Enable HTTP request logging")
	logDir := startCmd.String("log-dir", "", "Custom log directory (requires -log)")
	startCmd.Parse(os.Args[2:])

	if _, err := os.Stat(*folder); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Directory does not exist: %s\n", *folder)
		os.Exit(1)
	}

	if command == "start" {
		d := daemon.New(pidfile.New(), process.NewManager(), os.Args[0])
		if err := d.Start(*folder, *port, *bind, *enableLog, *logDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		var log *logger.Logger
		var err error
		if *enableLog {
			log, err = logger.New(true, *logDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
				os.Exit(1)
			}
			defer log.Close()
		}
		srv := server.New(*folder, *port, *bind, log)
		if err := srv.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	}
}

func handleStop() {
	d := daemon.New(pidfile.New(), process.NewManager(), os.Args[0])
	if err := d.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleStatus() {
	d := daemon.New(pidfile.New(), process.NewManager(), os.Args[0])
	if err := d.Status(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
