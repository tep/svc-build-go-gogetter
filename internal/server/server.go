package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/google/go-github/v25/github"
	"github.com/gorilla/mux"

	"toolman.org/base/log/v2"
	"toolman.org/net/http/httperr"

	"toolman.org/svc/build/go/gogetter/internal/config"
	"toolman.org/svc/build/go/gogetter/internal/xlat"
)

type Server struct {
	trans *xlat.Translator
	*config.Config
}

func New(cfg *config.Config, translator *xlat.Translator) *Server {
	return &Server{trans: translator, Config: cfg}
}

// TODO: ListenAndServe should accept a context for shutdown
func (s *Server) ListenAndServe() error {
	r := mux.NewRouter()

	r.Queries("go-get", "1").Handler(httperr.Handler(s.reroute))
	r.Handle("/hook", httperr.Handler(s.receiveHook))

	if s.Socket != "" {
		return s.fcgiServe(r)
	}

	// TODO: Add alternate support for FastCGI over unix-domain sockets.
	//       (See ~gto/gosrc-redirector/main.go?func=fcgiServe)
	addr := s.addr()
	log.Infof("HTTP Server Ready: %s", addr)
	return http.ListenAndServe(addr, r)
}

func (s *Server) Shutdown(ctx context.Context) {
	log.Warning("Server shutdown not yet implemented")
}

func (s *Server) fcgiServe(hndlr http.Handler) error {
	var (
		lis net.Listener
		err error
		del bool
	)

	if strings.ToLower(s.Socket) == "systemd" {
		lis, err = systemdListener()
	} else {
		if lis, err = net.Listen("unix", s.Socket); err == nil {
			del = true
		}
	}

	if err != nil {
		return err
	}

	defer func() {
		lis.Close()
		if del {
			os.Remove(s.Socket)
		}
	}()

	log.Infof("FCGI Server Ready: %s", s.Socket)
	return fcgi.Serve(lis, hndlr)
}

func systemdListener() (net.Listener, error) {
	lpid, lfds := os.Getenv("LISTEN_PID"), os.Getenv("LISTEN_FDS")
	if lpid == "" || lfds == "" {
		return nil, errors.New("systemd socket not found")
	}

	pid := os.Getpid()

	if i, err := strconv.Atoi(lpid); err != nil || i != pid {
		if err == nil {
			err = fmt.Errorf("systemd socket pid mismatch: got %d; wanted %d", i, pid)
		}
		return nil, err
	}

	if i, err := strconv.Atoi(lfds); err != nil || i != 1 {
		if err == nil {
			err = fmt.Errorf("systemd socket count mismatch: got %d; wanted 1", i)
		}
		return nil, err
	}

	return net.FileListener(os.NewFile(uintptr(3), "systemd:socket"))
}

func (s *Server) addr() string {
	return (&net.TCPAddr{Port: int(s.Port)}).String()
}

func (s *Server) reroute(w http.ResponseWriter, r *http.Request) error {
	log.V(1).Infof("GOGET: host=%q  uri=%q", r.Host, r.URL.Path)
	if repo := s.trans.Lookup(path.Join(r.Host, r.URL.Path)); repo != nil {
		repo.WriteImportTags(w)
	}
	return nil
}

func (s *Server) receiveHook(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return httperr.LogErrorf("bad request method: %s", r.Method).WithOptions(httperr.Status(http.StatusMethodNotAllowed))
	}

	enam := r.Header.Get("X-GitHub-Event")

	if enam == "integration_installation" || enam == "integration_installation_repositories" {
		log.V(1).Infof("Skipping deprecated event: %s", enam)
		return nil
	}

	log.Infof("Recieved event: %s", enam)

	payload, err := github.ValidatePayload(r, []byte(s.HookSecret))
	if err != nil {
		return httperr.LogErrorf("Failed payload validation: %v", err)
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		return httperr.LogErrorf("Bad webhook payload: %v", err)
	}

	switch evt := event.(type) {
	// InstallationEvent is triggered when a GitHub App has been
	// installed or uninstalled.
	// https://developer.github.com/v3/activity/events/types/#installationevent
	case *github.InstallationEvent:
		log.Infof("InstallationEvent: installation=%d sender=%q action=%q",
			evt.Installation.GetID(), evt.Sender.GetLogin(), evt.GetAction())
		for _, repo := range evt.Repositories {
			log.Infof("    REPO: id=%d %s", repo.GetID(), repo.GetFullName())
		}

	// InstallationRepositoriesEvent is triggered when a repository
	// is added or removed from an installation.
	// https://developer.github.com/v3/activity/events/types/#installationrepositoriesevent
	case *github.InstallationRepositoriesEvent:
		log.Infof("InstallationRepositoriesEvent: installation=%d sender=%q action=%q",
			evt.GetInstallation().GetID(), evt.GetSender().GetName(), evt.GetAction())
		for _, repo := range evt.RepositoriesAdded {
			log.Infof("    ADD: id=%d %s", repo.GetID(), repo.GetFullName())
		}

		for _, repo := range evt.RepositoriesRemoved {
			log.Infof("    REM: id=%d %s", repo.GetID(), repo.GetFullName())
		}

	// RepositoryEvent is triggered when a repository is created, archived,
	// unarchived, renamed, edited, transferred, made public, or made private.
	// (Organization hooks are also trigerred when a repository is deleted.)
	// https://developer.github.com/v3/activity/events/types/#repositoryevent
	case *github.RepositoryEvent:
		// TODO: Tell Translator to refresh this Repo (based on evt.GetAction())
		//
		if log.V(1) {
			log.Infof("RepositoryEvent: installation=%d repo=%q id=%d action=%q",
				evt.GetInstallation().GetID(), evt.GetRepo().GetFullName(), evt.GetRepo().GetID(), evt.GetAction())
		}

		s.trans.UpdateRepo(evt.GetRepo(), evt.GetAction() == "deleted")

	default:
		log.Warningf("Unhandled Event[%T]: %v", event, event)
	}

	return nil
}
