package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	args := parseArgs()
	dryRun := flag.Bool("dry", false, "Print the command but not execute")
	flag.Parse()

	deps, err := ParseDeps(args)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	cmdStr := deps.Expand()

	fmt.Println("Running: " + cmdStr)
	if *dryRun {
		return
	}

	cmd := exec.Command("sh", "-c", cmdStr)

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("output | %s\n", scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
		os.Exit(1)
	}

	err = cmd.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error waiting for Cmd", err)
		os.Exit(1)
	}
}

func parseArgs() []string {
	args := make([]string, 0)

	for _, arg := range os.Args[1:] {
		if !strings.HasPrefix(arg, "-") {
			args = append(args, arg)
		}
	}

	if len(args) == 0 {
		help()
		return make([]string, 0)
	}

	return args
}

func help() {
	fmt.Print("This is help")
}

// Dep holds all dependency information for a module
type Dep struct {
	sys   string
	build string
	php   string
	pecl  string
	npm   string
}

var reqs = map[string]Dep{
	"git": Dep{
		sys: "git openssh-client",
	},
	"crux": Dep{
		php: "bcmath sockets",
	},
	"soap": Dep{
		sys: "libxml2-dev",
		php: "soap",
	},
	"zip": Dep{
		sys: "zlib-dev",
		php: "zip",
	},
	"mysql": Dep{
		sys: "mysql-client",
		php: "pdo_mysql",
	},
	"postgres": Dep{
		sys: "postgresql-dev",
		php: "pdo_pgsql",
	},
	"swoole": Dep{
		build: "$PHPIZE_DEPS",
		pecl:  "swoole",
	},
	"pm2": Dep{
		sys: "nodejs",
		npm: "pm2",
	},
}

// Deps holds a list of deps to be installed
type Deps map[string]Dep

func (ds Deps) Expand() string {
	var sys, php, build, pecl []string

	for _, d := range ds {
		if d.sys != "" {
			sys = append(sys, d.sys)
		}
		if d.php != "" {
			php = append(php, d.php)
		}
		if d.build != "" {
			build = append(build, d.build)
		}
		if d.pecl != "" {
			pecl = append(pecl, d.pecl)
		}
	}

	var cmd []string
	var sysCmd, phpCmd, buildCmd, peclCmd string

	if len(sys) > 0 {
		sysCmd = "apk add --no-cache " + strings.Join(sys, " ")
		cmd = append(cmd, sysCmd)
	}

	if len(build) > 0 {
		buildCmd = "apk add --no-cache --virtual .build " + strings.Join(build, " ")
		cmd = append(cmd, buildCmd)
	}

	if len(pecl) > 0 {
		peclCmd = fmt.Sprintf("pecl install %s && docker-php-ext-enable %s", strings.Join(pecl, " "), strings.Join(pecl, " "))
		cmd = append(cmd, peclCmd)
	}

	if len(php) > 0 {
		phpCmd = "docker-php-ext-install " + strings.Join(php, " ")
		cmd = append(cmd, phpCmd)
	}

	if len(build) > 0 {
		cmd = append(cmd, "apk del .build")
	}

	return strings.Join(cmd, " && ")
}

func ParseDeps(vs []string) (Deps, error) {
	deps := make(Deps)

	for _, v := range vs {
		d, ok := reqs[v]
		if !ok {
			return deps, fmt.Errorf("%s is not recogized", v)
		}

		deps[v] = d
	}

	return deps, nil
}
