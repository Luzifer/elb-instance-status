package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v3"
)

const (
	checkKillWait            = 100 * time.Millisecond
	remoteConfigFetchTimeout = 2 * time.Second
)

type (
	checkCommand struct {
		Name     string `yaml:"name"`
		Command  string `yaml:"command"`
		WarnOnly bool   `yaml:"warn_only"`
	}

	checkResult struct {
		Check     checkCommand
		IsSuccess bool
		Streak    int64
	}
)

func executeAndRegisterCheck(ctx context.Context, checkID string) {
	var (
		check  = checks[checkID]
		logger = logrus.WithField("check_id", checkID)
	)

	cmd := exec.Command("/bin/bash", "-e", "-o", "pipefail", "-c", check.Command) //#nosec G204 // Intended to run an user-defined command
	cmd.Stderr = logger.WithField("stream", "STDERR").Writer()
	if cfg.Verbose {
		cmd.Stdout = logger.WithField("stream", "STDOUT").Writer()
	}
	err := cmd.Start()

	if err == nil {
		cmdDone := make(chan error)
		go func(cmdDone chan error, cmd *exec.Cmd) { cmdDone <- cmd.Wait() }(cmdDone, cmd)
		loop := true
		for loop {
			select {
			case err = <-cmdDone:
				loop = false
			case <-ctx.Done():
				logger.Error("execution of check will be killed through context timeout")
				if err := cmd.Process.Kill(); err != nil {
					logger.WithError(err).Error("killing check command")
				}
				time.Sleep(checkKillWait)
			}
		}
	}

	success := err == nil

	checkResultsLock.Lock()

	if _, ok := checkResults[checkID]; !ok {
		checkResults[checkID] = &checkResult{
			Check: check,
		}
	}

	if success == checkResults[checkID].IsSuccess {
		checkResults[checkID].Streak++
	} else {
		checkResults[checkID].IsSuccess = success
		checkResults[checkID].Streak = 1
	}

	if !success {
		logger.WithError(err).WithField("streak", checkResults[checkID].Streak).Warn("check failed, streak increased")
	}

	lastResultRegistered = time.Now()

	checkResultsLock.Unlock()
}

func loadChecks() error {
	var rawChecks io.Reader

	if _, err := os.Stat(cfg.CheckDefinitionsFile); err == nil {
		// We got a local file, read it
		f, err := os.Open(cfg.CheckDefinitionsFile)
		if err != nil {
			return fmt.Errorf("opening file: %w", err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				logrus.WithError(err).Error("closing configfile (leaked fd)")
			}
		}()
		rawChecks = f
	} else {
		// Check whether we got an URL
		if _, err := url.Parse(cfg.CheckDefinitionsFile); err != nil {
			return errors.New("definitions file is neither a local file nor a URL")
		}

		// We got an URL, fetch and read it
		ctx, cancel := context.WithTimeout(context.TODO(), remoteConfigFetchTimeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.CheckDefinitionsFile, nil)
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("executing request: %w", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				logrus.WithError(err).Error("closing configfile request body (leaked fd)")
			}
		}()
		rawChecks = resp.Body
	}

	tmpResult := map[string]checkCommand{}
	if err := yaml.NewDecoder(rawChecks).Decode(&tmpResult); err != nil {
		return fmt.Errorf("decoding checks file: %w", err)
	}

	checks = tmpResult
	return nil
}

func spawnChecks() {
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), cfg.CheckInterval-time.Second)

	wg.Add(len(checks))
	go func() {
		// Do not block the execution function but cleanup the context after
		// all checks are done (or cancelled)
		wg.Wait()
		cancel()
	}()

	for id := range checks {
		go func(ctx context.Context, id string) {
			defer wg.Done()
			executeAndRegisterCheck(ctx, id)
		}(ctx, id)
	}
}
