// qst - run things quickly
//
// Given a file or directory, guesses the project type and runs
// it for you. Restarts on changes. Intended for small experiments
// and working with unfamilar build systems.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/heyLu/qst/detect"
	"github.com/heyLu/qst/fileutil"
)

var delay = flag.Duration("delay", 1*time.Second, "time to wait until restart")
var autoRestart = flag.Bool("autorestart", false, "automatically restart after command exists")
var command = flag.String("command", "", "command to run ({file} will be substituted)")
var projectType = flag.String("type", "", "project type to use (autodetected if not present)")
var step = flag.String("step", "run", "which step to run (build, run or test)")
var justDetect = flag.Bool("detect", false, "detect the project type and exit")
var remote = flag.Bool("remote", false, "fetch and run a remote project")
var version = flag.Bool("v", false, "display version and exit")

var VERSION = "v0.1.0"

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <file>\n\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nSupported project types: \n")
		for _, project := range detect.ProjectTypes {
			paddedId := fmt.Sprintf("%s%s", project.Id, strings.Repeat(" ", 30-len(project.Id)))
			fmt.Fprintf(os.Stderr, "\t%s- %v\n", paddedId, project.Commands[*step])
		}
	}

	flag.Parse()
	args := flag.Args()

	if *version {
		fmt.Printf("qst %s\n", VERSION)
		os.Exit(0)
	}

	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	file := args[0]

	if *remote {
		f, err := fetchFromGit(file)
		if err != nil {
			log.Fatal(err)
		}
		file = *f
	}

	if *justDetect {
		projects := detect.DetectAll(file)
		if len(projects) == 0 {
			log.Fatal("unknown project type")
		}
		for _, project := range projects {
			fmt.Println(project.Id)
		}
		os.Exit(0)
	}

	var cmd string
	if !flagEmpty(command) {
		cmd = *command
	} else {
		project, err := detect.Detect(file)
		if !flagEmpty(projectType) {
			project = detect.GetById(*projectType)
			if project == nil {
				log.Fatalf("unknown type: `%s'", *projectType)
			}
		}
		if err != nil {
			log.Fatal("error: ", err)
		}
		log.Printf("detected a %s project", project.Id)
		projectCmd, found := project.Commands[*step]
		if !found {
			log.Fatalf("%s doesn't support `%s'", project.Id, *step)
		}
		cmd = projectCmd
	}
	file, _ = filepath.Abs(file)
	cmd = strings.Replace(cmd, "{file}", file, -1)
	if err := os.Chdir(fileutil.Dir(file)); err != nil {
		log.Fatal(err)
	}
	log.Printf("command to run: `%s'", cmd)

	runner := MakeRunner(cmd, *autoRestart)
	go runCmd(file, runner)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	s := <-c
	log.Printf("got signal: %s, exiting...", s)
	runner.Stop()
}

func flagEmpty(stringFlag *string) bool {
	return stringFlag == nil || strings.TrimSpace(*stringFlag) == ""
}

func fetchFromGit(file string) (*string, error) {
	if !strings.HasPrefix(file, "github.com") {
		return nil, errors.New("only github repos supported for now, sorry")
	}
	repo, dir, err := splitGithubUrl(file)
	if err != nil {
		return nil, err
	}
	log.Printf("cloning %s", *repo)
	cloneCmd := exec.Command("git", "clone", *repo)
	cloneCmd.Stderr = os.Stderr
	cloneCmd.Stdout = os.Stdout
	err = cloneCmd.Run()
	if err != nil {
		return nil, err
	}
	return dir, nil
}

func splitGithubUrl(repo string) (*string, *string, error) {
	repoWithPath := strings.TrimPrefix(strings.TrimPrefix(repo, "git://"), "github.com/")
	parts := strings.SplitN(repoWithPath, "/", 3)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return nil, nil, errors.New(fmt.Sprint("error: invalid repo: ", repo))
	}
	repoUrl := fmt.Sprintf("git://github.com/%s/%s", parts[0], parts[1])
	dir := parts[1]
	if len(parts) == 3 {
		dir = path.Join(dir, parts[2])
	}
	return &repoUrl, &dir, nil
}

func runCmd(file string, runner *Runner) {
	runner.Start()
	lastMtime := time.Now()
	for {
		info, err := os.Stat(file)
		if os.IsNotExist(err) {
			log.Fatalf("`%s' disappeared, exiting", file)
		}

		mtime := info.ModTime()
		if mtime.After(lastMtime) {
			log.Printf("`%s' changed, trying to restart", file)
			runner.Restart()
		}

		lastMtime = mtime
		time.Sleep(1 * time.Second)
	}
}

type Runner struct {
	cmd         *exec.Cmd
	shellCmd    string
	started     bool
	autoRestart bool
	restarting  bool
}

func MakeRunner(shellCmd string, autoRestart bool) *Runner {
	return &Runner{nil, shellCmd, false, autoRestart, false}
}

func (r *Runner) Start() error {
	if r.started {
		return errors.New("already started, use Restart()")
	}

	r.started = true
	go func() {
		for {
			log.Printf("[runner] starting command")
			r.cmd = exec.Command("sh", "-c", r.shellCmd)
			r.cmd.Stderr = os.Stderr
			r.cmd.Stdout = os.Stdout
			r.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			err := r.cmd.Run()
			var result interface{}
			if err != nil {
				result = err
			} else {
				result = r.cmd.ProcessState
			}
			log.Printf("[runner] finished: %s", result)

			time.Sleep(*delay)
			if !r.restarting && !r.autoRestart {
				r.started = false
				break
			}

			r.restarting = false
		}
	}()

	return nil
}

func (r *Runner) Kill() error {
	pgid, err := syscall.Getpgid(r.cmd.Process.Pid)
	if err == nil {
		syscall.Kill(-pgid, syscall.SIGTERM)
	}
	return err
}

func (r *Runner) Restart() error {
	r.restarting = true
	if r.started {
		return r.Kill()
	} else {
		return r.Start()
	}
}

func (r *Runner) Stop() {
	r.autoRestart = false
	r.Kill()
}

func isFile(file string) bool {
	info, err := os.Stat(file)
	return err == nil && !info.IsDir()
}
