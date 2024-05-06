package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type JenkinsConfig struct {
	Jenkins struct {
		URL  string `yaml:"url"`
		Auth struct {
			Username string `yaml:"username"`
			Token    string `yaml:"token"`
			Cookie   string `yaml:"cookie"`
			Crump    string `yaml:"crump"`
		} `yaml:"auth"`
	} `yaml:"jenkins"`
}

var (
	USERNAME            string = ""
	AUTHORIZATION_TOKEN string = ""
	JENKINS_URL         string = ""
	COOKIE              string = ""
	CRUMP               string = ""
)

func clen(s string) int {
	count := 0
	for _, char := range s {
		if char == '\n' {
			count++
		}
	}
	return count
}

func checkForInputRequested(consoleOutput string) bool {
	username, _ := base64.StdEncoding.DecodeString(USERNAME)
	lines := strings.Split(consoleOutput, "\n")
	for _, line := range lines {
		lightGray := "\033[90m"
		reset := "\033[0m"
		fmt.Printf("%s[%s@Jenkins] %s%s\n", lightGray, strings.Trim(string(username), "\n"), line, reset)
		if strings.Contains(line, "Input requested") {
			return true
		}
	}
	return false
}

func fullyAutomatedPipelineRun(jenkinsURL string) {
	myCookie := ""
	myCrumpi := ""
	// fmt.Printf("[pacJenker.go] Enter cookie used to login into Jenkins : ")
	// fmt.Scanf("%s", &myCookie)
	// fmt.Printf("[pacJenker.go] Enter cookie used to login into Jenkins : ")
	// fmt.Scanf("%s", &myCrumpi)
	username, _ := base64.StdEncoding.DecodeString(USERNAME)
	fmt.Printf("[pacJenker.go] Attempting to login as %s\n", string(username))

	triggerPipelineBuild(jenkinsURL)
	time.Sleep(370 * time.Millisecond) // Jenkins Cooldown timer
	buttonIDs := []string{"BuildForProductionInput", "BuildForProductionInput"}
	buttonParams := []string{"Build and deploy for QA?", "Build and deploy for production?"}
	buildNum, _ := getLastBuildNumber(jenkinsURL)
	lastLogScroll := 0
	elapsedTime := 0
	for i, buttonID := range buttonIDs {
		for {
			consoleLogs, _ := getConsoleLogs(jenkinsURL, buildNum, lastLogScroll)
			lastLogScroll += clen(consoleLogs)
			if checkForInputRequested(consoleLogs) {
				fmt.Printf("%s[INFO] Found an Input request .. Trying %s : %s%s\n", Blue, buttonID, buttonParams[i], Reset)
				err := approveJenkinsButton(jenkinsURL, buildNum, buttonIDs[i], buttonParams[i], myCookie, myCrumpi)
				if err != nil {
					fmt.Printf("[pacJenker.go] 🟣 ~ Aborted pipeline\n")
					stopPipelineBuild(jenkinsURL, buildNum)
					return
				}
				elapsedTime += 2
				time.Sleep(2 * time.Second)
				break
			}
			fmt.Printf("[INFO] Waiting for Button ... Elapsed %ds\n", elapsedTime)
			time.Sleep(5 * time.Second)
			elapsedTime += 5
		}
	}
	fmt.Printf("%s[pacJenker.go] Pipeline completed%s\n", Green, Reset)

}

func main() {
	data, err := ioutil.ReadFile("pac.yaml")
	if err != nil {
		fmt.Printf("%s[pacJenker.go] Missing pac.yaml%s\n", Red, Reset)
	}
	var jenkinsConfig JenkinsConfig
	err = yaml.Unmarshal(data, &jenkinsConfig)
	if err != nil {
		fmt.Printf("%s[pacJenker.go] Error unmarshal-ing pac.yaml%s\n", Red, Reset)
	}

	JENKINS_URL = jenkinsConfig.Jenkins.URL
	USERNAME = jenkinsConfig.Jenkins.Auth.Username
	AUTHORIZATION_TOKEN = jenkinsConfig.Jenkins.Auth.Token
	COOKIE = jenkinsConfig.Jenkins.Auth.Cookie
	CRUMP = jenkinsConfig.Jenkins.Auth.Crump

	fullyAutomatedPipelineRun("https://jenkins.something.org/job/FOLDER1/job/JOB1/job/MyBranch")
}
