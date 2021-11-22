package main

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	hostFilePath := detectHostFilePath()
	hostList := backupOtherHosts(hostFilePath)

	domainNames := []string{"github.com", "github.global.ssl.fastly.net", "raw.githubusercontent.com", "assets-cdn.github.com"}
	hostList = append(hostList, "\n# GitHub")
	for i := 0; i < len(domainNames); i++ {
		hostIP := resolveIPAddress("https://websites.ipaddress.com/" + domainNames[i])
		hostList = append(hostList, hostIP+" "+domainNames[i])
	}

	writeHostToFile(hostFilePath, hostList)

	flushDNS()
}

func detectHostFilePath() string {
	osType := runtime.GOOS
	log.Println("Detecting your operating system, it is", osType)

	hostFilePath := ""
	if osType == "windows" {
		hostFilePath = "C:\\Windows\\System32\\drivers\\etc\\hosts"
	} else if osType == "linux" || osType == "mac" {
		hostFilePath = "/etc/hosts"
	} else {
		log.Fatal("Sorry, this operating system is not supported at this moment: ", osType)
	}
	return hostFilePath
}

func backupOtherHosts(hostFilePath string) []string {
	hostFile, err := os.OpenFile(hostFilePath, os.O_RDONLY, 0666)
	if err != nil {
		log.Fatal("Error occurred when attempt to open 'hosts' file, error detail: ", err)
	}
	defer hostFile.Close()

	scanner := bufio.NewScanner(hostFile)

	otherHosts := []string{}
	for scanner.Scan() {
		if !strings.Contains(strings.ToLower(scanner.Text()), "github") {
			otherHosts = append(otherHosts, scanner.Text())
		}
	}
	return otherHosts
}

func resolveIPAddress(url string) string {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal("Error occurred when executing goquery.NewDocument(url), error detail: ", err)
	}
	ip := doc.Find("ul.comma-separated > li").First().Text()
	return ip
}

func writeHostToFile(hostFilePath string, hostList []string) {
	hostFile, err := os.OpenFile(hostFilePath, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal("Error occurred when attempt to open 'hosts' file, error detail: ", err)
	}
	defer hostFile.Close()

	writer := bufio.NewWriter(hostFile)
	writer.WriteString(strings.Join(hostList, "\n"))
	writer.Flush()
}

func flushDNS() {
	log.Println("Ready to flush DNS cache")
	osType := runtime.GOOS
	if osType == "windows" {
		exec.Command("ipconfig", "/flushdns").Run()
	} else if osType == "linux" {
		exec.Command("service", "nscd", "restart").Run()
	} else if osType == "mac" {
		exec.Command("killall", "-HUP", "mDNSResponder").Run()
	} else {
		log.Fatal("Sorry, this operating system is not supported at this moment: ", osType)
	}
}
