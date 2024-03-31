package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"github.com/Luzifer/rconfig/v2"
)

var (
	cfg = struct {
		CheckDefinitionsFile string `flag:"check-definitions-file,c" default:"/etc/elb-instance-status.yml" description:"File or URL containing checks to perform for instance health"`
		UnhealthyThreshold   int64  `flag:"unhealthy-threshold" default:"5" description:"How often does a check have to fail to mark the machine unhealthy"`

		CheckInterval         time.Duration `flag:"check-interval" default:"1m" description:"How often to execute checks (do not set below 10s!)"`
		ConfigRefreshInterval time.Duration `flag:"config-refresh" default:"10m" description:"How often to update checks from definitions file / url"`

		Verbose  bool   `flag:"verbose,v" default:"false" description:"Attach stdout of the executed commands"`
		LogLevel string `flag:"log-level" default:"info" description:"Log level (debug, info, warn, error, fatal)"`

		Listen         string `flag:"listen" default:":3000" description:"IP/Port to listen on for ELB health checks"`
		VersionAndExit bool   `flag:"version" default:"false" description:"Print version and exit"`
	}{}

	version = "dev"

	checks               map[string]checkCommand
	checkResults         = map[string]*checkResult{}
	checkResultsLock     sync.RWMutex
	lastResultRegistered time.Time
)

func initApp() (err error) {
	rconfig.AutoEnv(true)
	if err = rconfig.Parse(&cfg); err != nil {
		return fmt.Errorf("parsing CLI options: %w", err)
	}

	l, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("parsing log-level: %w", err)
	}
	logrus.SetLevel(l)

	return nil
}

func main() {
	var err error
	if err = initApp(); err != nil {
		logrus.WithError(err).Fatal("initializing app")
	}

	if cfg.VersionAndExit {
		fmt.Printf("elb-instance-status %s\n", version) //nolint:forbidigo
		os.Exit(0)
	}

	if err = loadChecks(); err != nil {
		logrus.WithError(err).Fatal("reading definitions file")
	}

	c := cron.New()

	if _, err = c.AddFunc(fmt.Sprintf("@every %s", cfg.CheckInterval), spawnChecks); err != nil {
		logrus.WithError(err).Fatal("registering spawn function")
	}

	if _, err = c.AddFunc(fmt.Sprintf("@every %s", cfg.ConfigRefreshInterval), func() {
		if err := loadChecks(); err != nil {
			logrus.WithError(err).Error("refreshing checks")
		}
	}); err != nil {
		logrus.WithError(err).Fatal("registering config-refresh function")
	}

	c.Start()

	spawnChecks()

	http.HandleFunc("/status", handleELBHealthCheck)

	server := &http.Server{
		Addr:              cfg.Listen,
		Handler:           http.DefaultServeMux,
		ReadHeaderTimeout: time.Second,
	}

	if err = server.ListenAndServe(); err != nil {
		logrus.WithError(err).Fatal("listening for HTTP traffic")
	}
}

func handleELBHealthCheck(w http.ResponseWriter, _ *http.Request) {
	var (
		healthy = true
		start   = time.Now()
		buf     = new(bytes.Buffer)
	)

	checkResultsLock.RLock()
	for _, cr := range checkResults {
		state := ""
		switch {
		case cr.IsSuccess:
			state = "PASS"
		case !cr.IsSuccess && cr.Check.WarnOnly:
			state = "WARN"
		case !cr.IsSuccess && !cr.Check.WarnOnly && cr.Streak < cfg.UnhealthyThreshold:
			state = "CRIT"
		case !cr.IsSuccess && !cr.Check.WarnOnly && cr.Streak >= cfg.UnhealthyThreshold:
			state = "CRIT"
			healthy = false
		}
		fmt.Fprintf(buf, "[%s] %s\n", state, cr.Check.Name)
	}
	checkResultsLock.RUnlock()

	w.Header().Set("X-Collection-Parsed-In", strconv.FormatInt(time.Since(start).Nanoseconds()/int64(time.Microsecond), 10)+"ms")
	w.Header().Set("X-Last-Result-Registered-At", lastResultRegistered.Format(time.RFC1123))
	if healthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}

	if _, err := io.Copy(w, buf); err != nil {
		logrus.WithError(err).Error("writing HTTP response body")
	}
}
