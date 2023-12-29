package main

import (
	"compress/gzip"
	"flag"
	"io"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var BlockListURL = "https://raw.githubusercontent.com/scriptzteam/IP-BlockList-v4/master/ips.txt"

func main() {
	var daemonize bool
	flag.StringVar(&BlockListURL, "b", BlockListURL, "blocklist url")
	flag.BoolVar(&daemonize, "d", false, "daemonize")
	flag.Parse()

	// check if we have ability to call iptables
	cmd := exec.Command("iptables", "-L")
	cmd.Run()
	switch cmd.ProcessState.ExitCode() {
	case 4:
		log.Fatal("Error: no permission to run iptables")
	case 127:
		log.Fatal("Error: iptables command not found")
	case 0:
		// all good
	default:
		log.Fatal("Error: unknown error running iptables")
	}

	// check ipset is installed and callablle
	cmd = exec.Command("ipset", "list")
	cmd.Run()
	switch cmd.ProcessState.ExitCode() {
	case 127:
		log.Fatal("Error: ipset command not found")
	case 0:
		// all good
	default:
		log.Fatal("Error: ipset command failed (need root?)")
	}

	if daemonize {
		loop()
		return
	}

	pullAndSet(BlockListURL)
}

func loop() {
	for {
		time.Sleep(time.Hour * 24 * 7)
		pullAndSet(BlockListURL)
	}
}

func pullAndSet(url string) {
	log.Println("Pulling IPs from", url)
	data := pull(url)
	setIptables(parseIps(data))
	log.Println("Done")
}

func pull(url string) []byte {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Error creating http request: ", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:78.0) Gecko/20100101 Firefox/78.0")
	req.Header.Set("Accept-Encoding", "gzip")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("Error making http request: ", err)
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatal("Bad status code: ", res.StatusCode)
	}

	if enc := res.Header.Get("Content-Encoding"); enc != "gzip" {
		log.Fatal("Bad Content-Encoding: ", enc)
	}

	r, err := gzip.NewReader(res.Body)
	if err != nil {
		log.Fatal("Error creating gzip reader: ", err)
	}

	data, err := io.ReadAll(r)
	if err != nil {
		log.Fatal("Error reading gzip data: ", err)
	}

	return data
}

func parseIps(data []byte) []string {
	var ips []string
	lines := strings.Split(string(data), "\n")
	log.Printf("Found %d lines", len(lines))

	omittedLevels := regexp.MustCompile(`\s[1-2]$`)
	var omittedLevelCount int
	for _, line := range lines {
		// omit comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		// omit empty lines
		if line == "" {
			continue
		}

		// omit lines matching regex that contain a 1 or a 2 on their own
		if omittedLevels.Match([]byte(line)) {
			omittedLevelCount++
			continue
		}

		// pull the first word from the string
		ips = append(ips, strings.Fields(line)[0])
	}

	log.Printf("Omitted %d IPs due to low level", omittedLevelCount)
	log.Printf("Found %d IPs to block", len(ips))
	return ips
}

func setIptables(ips []string) {
	name := "blocklist"
	log.Println("Flushing and recreating IP setP", name)

	if err := exec.Command("ipset", "-q", "list", name).Run(); err == nil {
		if err := exec.Command("ipset", "-q", "flush", name).Run(); err != nil {
			log.Fatal("Error flushing ipset: ", err)
		}
	}

	if err := exec.Command("ipset", "-q", "create", name, "hash:net").Run(); err != nil {
		log.Fatal("Error creating ipset: ", err)
	}

	log.Printf("Adding %d IPs to the IP set", len(ips))
	for _, ip := range ips {
		if err := exec.Command("ipset", "-q", "add", name, ip).Run(); err != nil {
			log.Fatal("Error adding ip to ipset: ", err)
			break
		}
	}

	if err := exec.Command("iptables", "-I", "INPUT", "-m", "set", "--match-set", name, "src", "-j", "DROP").Run(); err != nil {
		log.Fatal("Error adding iptables rule: ", err)
	}
}
