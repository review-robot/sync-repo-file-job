package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/opensourceways/community-robot-lib/logrusutil"
	"github.com/opensourceways/sync-repo-file/client"
	"github.com/opensourceways/sync-repo-file/server"
	"github.com/sirupsen/logrus"
)

type options struct {
	platform  string
	endpoint  string
	fileNames string
	orgRepos  string
}

func (o options) validate() error {
	if o.endpoint == "" {
		return fmt.Errorf("endpoint must be set")
	}

	if o.orgRepos == "" {
		return fmt.Errorf("orgRepos must be set")
	}

	return nil
}

func gatherOptions(fs *flag.FlagSet, args ...string) options {
	var o options

	fs.StringVar(&o.platform, "platform", "gitee", "the platform of the repository that needs to sync files")
	fs.StringVar(&o.fileNames, "fileNames", "OWNERS", "the file name that needs to be synchronized, multiple separated ,")
	fs.StringVar(&o.endpoint, "endpoint", "", "grpc connection address")
	fs.StringVar(&o.orgRepos, "orgRepos", "", "the full path of the repository that needs to be synchronized, e.g: org/repo. multiple separated with ,")

	_ = fs.Parse(args)

	return o
}

func main() {
	logrusutil.ComponentInit("sync-repo-file-job")

	o := gatherOptions(flag.NewFlagSet(os.Args[0], flag.ExitOnError), os.Args[1:]...)
	if err := o.validate(); err != nil {
		logrus.WithError(err).Error("parse option")
		return
	}

	syncConf := transOpt2SyncConfig(o)
	logrus.Info(syncConf)

	clients := initClients(o.platform, o.endpoint)
	if len(clients) == 0 {
		return
	}

	defer func() {
		for _, cli := range clients {
			_ = cli.Stop()
		}
	}()

	clis := map[string]server.SyncFileClient{}
	for k, v := range clients {
		clis[k] = v
	}
	wait, cancel := server.DoOnce(clis, syncConf, 10)

	run(wait, cancel)
}

func transOpt2SyncConfig(o options) []server.SyncFileConfig {
	sfc := server.SyncFileConfig{}

	sfc.Platform = o.platform
	sfc.FileNames = strings.Split(o.fileNames, ",")

	rrs := strings.Split(o.orgRepos, ",")
	mr := make(map[string][]string, 0)

	for _, v := range rrs {
		ts := strings.Split(v, "/")
		tOrg, tRepo := "", ""

		if len(ts) == 0 {
			continue
		}

		if len(ts) == 1 {
			tOrg = ts[0]
		} else {
			tOrg = ts[0]
			tRepo = ts[1]
		}

		if tOrg == "" {
			continue
		}

		if _, ok := mr[tOrg]; ok {
			if tRepo != "" {
				mr[tOrg] = append(mr[tOrg], tRepo)
			}
		} else {
			if tRepo != "" {
				mr[tOrg] = []string{tRepo}
			} else {
				mr[tOrg] = []string{}
			}
		}
	}

	for k, v := range mr {
		sfc.OrgRepos = append(sfc.OrgRepos, server.OrgRepos{
			Org:   k,
			Repos: v,
		})
	}

	return []server.SyncFileConfig{sfc}
}

func initClients(platform, endpoint string) map[string]*client.SyncFileClient {
	clients := map[string]*client.SyncFileClient{}

	cli, err := client.NewSyncFileClient(endpoint)
	if err != nil {
		logrus.WithField("enpoint", endpoint).WithError(err).Infof(
			"init sync file client for platform:%s", platform,
		)
	}

	clients[platform] = cli

	return clients
}

func run(wait, cancel func()) {
	done := make(chan struct{})
	sig := make(chan os.Signal, 1)

	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup
	defer func() {
		wg.Wait()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		select {
		case <-done:
			logrus.Info("receive done. exit normally")
			return
		case <-sig:
			logrus.Info("receive exit signal")
			cancel()
			return
		}
	}()

	wait()
	close(done)
}
