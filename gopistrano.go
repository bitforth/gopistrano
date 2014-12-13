package main

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/goconf/conf"
	"fmt"
	"github.com/laher/scp-go/scp"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	args := os.Args[1]

	// reading config file
	c, err := conf.ReadConfigFile("Gopfile")
	user, err := c.GetString("", "username")
	pass, err := c.GetString("", "password")
	hostname, err := c.GetString("", "hostname")
	repository, err := c.GetString("", "repository")
	path, err := c.GetString("", "path")
	timestamp := time.Now().Local()
	releases := path + "releases"
	shared := path + "shared"
	utils := path + "utils"
	release := path + timestamp.Format("20060102150405")
	use_sudo, err := c.GetBool("false", "use_sudo")
	keep_releases, err := c.GetInt("3", "keep_releases")

	sudo := ""
	if use_sudo {
		sudo = "sudo "
	}

	fmt.Println("SSH-ing into " + hostname)
	// initialize the structure with the configuration for ssh packat.
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
	}

	client, err := ssh.Dial("tcp", hostname+":22", config)
	// Do panic if the dial fails
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}

	if strings.EqualFold(args, "deploy:setup") {
		// Checking if setup folders exist. If they don't create them.
		fmt.Println("Setting up gopistrano structure")

		session, err := client.NewSession()
		if err != nil {
			panic("Failed to create session: " + err.Error())
		}
		defer session.Close()
		cdPathCmd := sudo + "if [ ! -d " + releases + " ]; then mkdir " + releases + "; fi &&" +
			"if [ ! -d " + shared + " ]; then mkdir " + shared + "; fi &&" +
			"if [ ! -d " + utils + " ]; then mkdir " + utils + "; fi &&" +
			"chmod g+w " + releases + " " + shared + " " + path + " " + utils
		if err := session.Run(cdPathCmd); err != nil {
			panic("Failed to run: " + err.Error())
		}
		var b bytes.Buffer
		session.Stdout = &b

		scpArgs := []string{"utils/*", user + "@" + hostname + ":" + utils}
		// Creating another connection to scp the file
		err, status := scp.ScpCli(scpArgs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(status)
		}
	} else if strings.EqualFold(args, "deploy") {
		// check that the releases directory exists
		fmt.Println("Checking if " + releases + " exists")
		cdPathCmd := sudo + "if [ ! -d " + releases + " ]; then exit 1; fi"
		if err := session.Run(cdPathCmd); err != nil {
			fmt.Println("You have to run deploy:setup before running deploy!")
			return
		}

		fmt.Println(releases + " exists")

		fmt.Println("Counting directories...")
		directoryCountCmd := sudo + "ls -lR " + releases + " | grep ^d | wc -l"
		fmt.Println(directoryCountCmd)
		if err := session.Run(directoryCountCmd); err != nil {
			panic("Failed to run: " + err.Error())
		}
		// counting number of directories in releases
		directoryCount, err := strconv.Atoi(b.String())
		if err != nil {
			panic("Failed conversion to integer: " + err.Error())
		}
		if directoryCount == keep_releases {
			// removing oldest directory
			removeOldestDirectoryCmd := sudo + "rm -Rf $(ls -td " + releases + "/* | cut -d' ' -f1 | tail -1)"
			if err := session.Run(removeOldestDirectoryCmd); err != nil {
				panic("Failed to run: " + err.Error())
			}

		} else {
			// creating directory for new release
			timestampDirCmd := sudo + "mkdir " + release
			if err := session.Run(timestampDirCmd); err != nil {
				panic("Failed to run: " + err.Error())
			}

			// cloning repository into directory
			gitCloneCmd := sudo + "git clone " + repository + " " + release
			if err := session.Run(gitCloneCmd); err != nil {
				panic("Failed to run: " + err.Error())
			}

			//symbolic link the latest release
			symLinkReleaseCmd := sudo + "ln -s " + release + " current"
			if err := session.Run(symLinkReleaseCmd); err != nil {
				panic("Failed to run: " + err.Error())
			}
		}
	} else {
		fmt.Println("You have to run deploy or deploy:setup")
		return
	}
	//	fmt.Println(b.String())
}
