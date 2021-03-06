package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/koron/netupvim/netup"
)

var (
	version = "none"
)

var (
	targetDir  = "."
	sourceName = "release"
	cpu        string
	restore    bool
	selfUpdate = true
)

func setup() error {
	conf, err := loadConfig("netupvim.ini")
	if err != nil {
		return err
	}

	// Parse options.
	var (
		helpOpt    = flag.Bool("h", false, "show this message")
		targetOpt  = flag.String("t", conf.getTargetDir(), "target dir to upgrade/install")
		sourceOpt  = flag.String("s", conf.getSource(), "source of update: release,develop,canary")
		restoreOpt = flag.Bool("restore", false, "force download & extract all files")
		versionOpt = flag.Bool("version", false, "show version")
	)
	flag.Parse()
	if *helpOpt {
		showHelp()
		os.Exit(1)
	}
	if *versionOpt {
		showVersion()
		os.Exit(1)
	}

	// setup context.
	targetDir = *targetOpt
	sourceName = *sourceOpt
	restore = *restoreOpt
	cpu = conf.CPU
	selfUpdate = !conf.DisableSelfUpdate

	netup.Version = version
	netup.DownloadTimeout = conf.getDownloadTimeout()
	netup.GithubUser = conf.getGithubUser()
	netup.GithubToken = conf.getGithubToken()
	netup.GithubVerbose = conf.GithubVerbose
	if conf.LogRotateCount > 0 {
		netup.LogRotateCount = conf.LogRotateCount
	}
	if conf.ExeRotateCount > 0 {
		netup.ExeRotateCount = conf.ExeRotateCount
	}

	return nil
}

func shouldSelfUpdate() bool {
	if !selfUpdate {
		return false
	}
	_, err := os.Stat(filepath.Join(targetDir, "netupvim.exe"))
	return err == nil
}

func run() error {
	if err := setup(); err != nil {
		return err
	}
	workDir := filepath.Join(targetDir, "netupvim")
	// update vim
	vimPack, ok := vimSet[sourceName]
	if !ok {
		return fmt.Errorf("invalid source: %s", sourceName)
	}
	err := netup.Update(
		targetDir,
		workDir,
		vimPack,
		netup.Arch{Name: cpu, Hint: "vim.exe"},
		restore)
	if err != nil {
		return err
	}
	// try to update netupvim
	if shouldSelfUpdate() {
		netup.LogInfo("trying to update netupvim")
		err := netup.Update(
			targetDir,
			workDir,
			netupPack,
			netup.Arch{Name: "X86"},
			restore)
		if err != nil {
			netup.LogInfo("failed to udate netupvim: %s", err)
		}
	}
	return nil
}

func showHelp() {
	fmt.Fprintf(os.Stderr, `%[1]s is tool to upgrade/install Vim (+kaoriya) in/to target dir.

Usage: %[1]s [options]

Options are:
`, filepath.Base(os.Args[0]))
	flag.PrintDefaults()
}

func showVersion() {
	fmt.Fprintf(os.Stderr, "netupvim version %s\n", version)
}

func main() {
	if err := run(); err != nil {
		netup.LogFatal(err)
	}
}
