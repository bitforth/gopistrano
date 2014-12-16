package main

import (
	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/goconf/conf"
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	user,
	pass,
	hostname,
	repository,
	path,
	releases,
	shared,
	utils,
	keep_releases string
)

func init() {
	var err error
	// reading config file

	c, err := conf.ReadConfigFile("Gopfile")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	user, err = c.GetString("", "username")
	pass, err = c.GetString("", "password")
	hostname, err = c.GetString("", "hostname")
	repository, err = c.GetString("", "repository")
	path, err = c.GetString("", "path")
	releases = path + "/releases"
	shared = path + "/shared"
	utils = path + "/utils"

	keep_releases, err = c.GetString("", "keep_releases")

	//just log whichever we get; let the user re-run the program to see all errors... for now
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func main() {
	flag.Parse()

	action := flag.Arg(0)

	if action == "" {
		fmt.Println("Error: use gopistrano deploy or gopistrano deploy:setup")
		return
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
		fmt.Println("Failed to dial: " + err.Error())
		return
	}

	switch strings.ToLower(action) {
	case "deploy:setup":
		err = deploySetup(client)
	case "deploy":
		err = deploy(client)
	default:
		fmt.Println("Invalid command!")
	}

	if err != nil {
		fmt.Println(err.Error())
	}
}

// runs the deployment script remotely
func deploy(client *ssh.Client) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	deployCmd := "if [ ! -d " + releases + " ]; then exit 1; fi &&" +
		"if [ ! -d " + shared + " ]; then exit 1; fi &&" +
		"if [ ! -d " + utils + " ]; then exit 1; fi &&" +
		"if [ ! -f " + utils + "/deploy.sh ]; then exit 1; fi &&" +
		"bash " + utils + "/deploy.sh " + path + " " + repository + " " + keep_releases

	if err := session.Run(deployCmd); err != nil {
		return err
	}

	//send through
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	fmt.Println("Project Deployed!")
	return nil
}

// sets up directories for deployment a la capistrano
func deploySetup(client *ssh.Client) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}

	//send through to main stdout, stderr
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	defer session.Close()

	cdPathCmd := "if [ ! -d " + releases + " ]; then mkdir " + releases + "; fi &&" +
		"if [ ! -d " + shared + " ]; then mkdir " + shared + "; fi &&" +
		"if [ ! -d " + utils + " ]; then mkdir " + utils + "; fi &&" +
		"chmod g+w " + releases + " " + shared + " " + path + " " + utils

	if err := session.Run(cdPathCmd); err != nil {
		return err
	}

	fmt.Println("running scp connection")

	scpSession, err := client.NewSession()
	if err != nil {
		return err
	}

	defer scpSession.Close()

	cpy := `echo -n '` + string(deployment_script) + `' > ` + utils + `/deploy.sh ; chmod +x ` + utils + `/deploy.sh`

	if err := scpSession.Run(cpy); err != nil {
		return err
	}

	fmt.Println("Cool Beans! Gopistrano created the structure correctly!")
	return nil
}
