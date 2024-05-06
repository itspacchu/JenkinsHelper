package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	Reset  = "\033[0m"
	Black  = "\033[30m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
)

func buildJenkinsUrl(project string, jobName string, branch string) string {
	endUrl := fmt.Sprintf("%s/job/%s/job/%s/job/%s", JENKINS_URL, project, jobName, branch)
	return endUrl
}

func doAuthenticatedRequest(jenkinsAPIEndpoint string, method string) (string, error) {
	req, _ := http.NewRequest(method, jenkinsAPIEndpoint, nil)
	req.Header.Set("Authorization", AUTHORIZATION_TOKEN)
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode > 299 {
		fmt.Printf("[INFO: %d] Failed to connect to jenkins\n", response.StatusCode)
	}
	body, err := io.ReadAll(response.Body)
	return string(body), err
}

func getConsoleLogs(jenkinsUrl string, buildNum int, startAt int) (string, error) {
	consoleUrl := fmt.Sprintf("%s/%d/logText/progressiveText?start=%d", jenkinsUrl, buildNum, startAt)
	body, err := doAuthenticatedRequest(consoleUrl, "GET")
	return body, err
}

func getLastBuildNumber(jenkinsURL string) (int, error) {
	apiURL := fmt.Sprintf("%s/lastBuild/buildNumber", jenkinsURL)
	body, err := doAuthenticatedRequest(apiURL, "GET")
	buildNum, _ := strconv.Atoi(string(body))
	return buildNum, err
}

func stopPipelineBuild(jenkinsUrl string, buildNum int) {
	triggerURL := jenkinsUrl + "/" + string(buildNum) + "/stop"
	_, err := doAuthenticatedRequest(triggerURL, "POST")
	if err != nil {
		fmt.Printf("%s[pacJenker.go]%s 🔴 ~ Unable to stop pipeline build : %s", Red, Reset, err)
		return
	}
	fmt.Printf("%s[pacJenker.go]%s 🟢 ~ Triggered Pipeline stop %s\n", Green, Reset, jenkinsUrl)

}

func triggerPipelineBuild(jenkinsUrl string) {
	triggerURL := jenkinsUrl + "/build"
	_, err := doAuthenticatedRequest(triggerURL, "POST")
	if err != nil {
		fmt.Printf("%s[pacJenker.go]%s 🔴 ~ Unable to trigger pipeline build : %s", Red, Reset, err)
		return
	}
	fmt.Printf("%s[pacJenker.go]%s 🟢 ~ Triggered Pipeline build for %s. Fetching Build number (waiting for 10s) ...\n", Green, Reset, jenkinsUrl)
	time.Sleep(10 * time.Second)
	buildNum, _ := getLastBuildNumber(jenkinsUrl)
	fmt.Printf("%s[pacJenker.go]%s 🟢 ~ Latest Build number : %d\n", Green, Reset, buildNum)

}

func approveJenkinsButton(jenkinsURL string, buildNum int, buttonID string, buttonParam string, cookieStr string, crumpStr string) error {
	// apiUrl = f"{JENKINS_URL}/job/{PROJECT}/job/{FOLDER}/job/{BRANCH}/{buildNumber}/wfapi/inputSubmit?inputId={buttonID}"
	buttonURL := fmt.Sprintf("%s/%d/wfapi/inputSubmit?inputId=%s", jenkinsURL, buildNum, buttonID)
	senderData := `{"parameter": [{"name": "NAMEPAR", "value": true}]}`
	senderData = strings.ReplaceAll(senderData, "NAMEPAR", buttonParam)

	req, err := http.NewRequest("POST", buttonURL, bytes.NewBuffer([]byte("json="+senderData)))
	if err != nil {
		fmt.Printf("[pacJenker.go] 🔴 ~ Unable to create a request >> %s", err)
		return err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:125.0) Gecko/20100101 Firefox/125.0")
	req.Header.Set("Accept", "text/javascript, text/html, application/xml, text/xml, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Referer", "https://jenkins.zigram.org/job/SECRET_SAUCE/job/ss-dev-helm-chart-deployment/job/test-jenkins/")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("X-Prototype-Version", "1.7")
	req.Header.Set("Content-type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Jenkins-Crumb", crumpStr)
	req.Header.Set("Origin", "https://jenkins.zigram.org")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", cookieStr)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("TE", "trailers")

	client := &http.Client{}
	response, err := client.Do(req)
	if response.StatusCode == 200 {
		fmt.Printf("[pacJenker.go] 🟢 ~ {%d} Approved Button %s (%s)\n", response.StatusCode, buttonID, buttonParam)
	} else {
		fmt.Printf("%s[pacJenker.go]%s 🟡 ~ {%d} Failed to approve Button %s (%s)\n", Yellow, Reset, response.StatusCode, buttonID, buttonParam)
		fmt.Printf("%s[pacJenker.go]%s 🟡 ~ Link to manually Approve it : %s\n[pacJenker.go] ~ Continue (y/N)", Yellow, Reset, buttonURL)
		var continueWaiting rune
		fmt.Scanf("%c", &continueWaiting)
		if continueWaiting == 'y' {
			return nil
		} else {
			fmt.Printf("%s[pacJenker.go]%s 🟣 ~ Aborting pipeline\n", Purple, Reset)
			err := errors.New("aborted pipeline")
			panic(err)
		}
	}
	defer response.Body.Close()
	if err != nil {
		fmt.Printf("%s[pacJenker.go]%s 🔴 ~ Unable to approve button\n%s", Red, Reset, err)
		return err
	}
	return nil
}
