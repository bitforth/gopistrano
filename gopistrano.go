package main

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/goconf/conf"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {

	if len(os.Args) == 1 {
		fmt.Println("This is not the right way to use this program! use gopistrano deploy or gopistrano deploy:setup")
		return
	}
	args := os.Args[1]

	// reading config file
	c, err := conf.ReadConfigFile("Gopfile")
	user, err := c.GetString("", "username")
	pass, err := c.GetString("", "password")
	hostname, err := c.GetString("", "hostname")
	repository, err := c.GetString("", "repository")
	path, err := c.GetString("", "path")
	timestamp := time.Now().Local()
	releases := path + "/releases"
	shared := path + "/shared"
	utils := path + "/utils"
	release := path + "/" + timestamp.Format("20060102150405")
	use_sudo, err := c.GetBool("false", "use_sudo")
	keep_releases, err := c.GetInt("3", "keep_releases")
	deployment_script, err := ioutil.ReadFile("utils/deploy.sh")

	fmt.Println(repository)
	fmt.Println(release)
	fmt.Println(keep_releases)
	fmt.Println(utils)

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

		fmt.Println("running scp connection")
		scpSession, err := client.NewSession()
		if err != nil {
			panic("Failed to create session: " + err.Error())
		}
		defer scpSession.Close()

		go func() {
			w, _ := scpSession.StdinPipe()
			defer w.Close()
			content := string(deployment_script)
			fmt.Fprintln(w, "C0755 "+strconv.Itoa(len(content))+" deploy.sh")
			fmt.Fprint(w, content)
			fmt.Fprint(w, "\x00")
		}()
		if err := scpSession.Run("/usr/bin/scp -t " + utils); err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println("Cool Beans! Gopistrano created the structure correctly!")
		return
	} else if strings.EqualFold(args, "deploy") {
		// check that the releases directory exists
	} else {
		fmt.Println("You have to run deploy or deploy:setup")
		return
	}
	//	fmt.Println(b.String())
}
