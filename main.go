package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/axgle/mahonia"
	flag "github.com/axgle/pflag"
	"github.com/axgle/util"
	"github.com/kardianos/service"
)

var (
	timeout            *int
	number             *int
	pingtime           string
	result             bool
	install            *bool
	uninstall          *bool
	host               *string
	hostfile           *string
	serviceName        *string
	serviceDescription *string
	webhook            = "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key="
	programName        = strings.Split(os.Args[0], "\\")[len(strings.Split(os.Args[0], "\\"))-1]
)

type Program struct{}

func (p *Program) Start(s service.Service) error {
	log.Println("开始服务")
	go p.doWork()
	return nil
}

func (p *Program) Stop(s service.Service) error {
	log.Println("停止服务")
	return nil
}

func (p *Program) doWork() {

	ips := util.IPparse(*host, *hostfile)
	for {
		result = false
		for _, ip := range ips {
			result = ping(ip)
			if result {
				sendIp(ip)
				os.Exit(0)
			}
		}
	}
}

func httpclient(url string, method string, body string) string {
	var reqest *http.Request

	if method == "GET" {
		reqest, _ = http.NewRequest("GET", url, nil)
	} else if method == "POST" {
		reqest, _ = http.NewRequest("POST", url, strings.NewReader(body))
		reqest.Header.Add("Content-Type", "application/json")
	}

	// handle reqest
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// handle response
	response, err := client.Do(reqest)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		panic(err)
	}

	resbody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	resbodyData := string(resbody)
	return resbodyData
}

func ping(dstIP string) bool {
	var OS = runtime.GOOS
	var command *exec.Cmd
	if OS == "windows" {
		command = exec.Command("cmd", "/c", "ping -n "+strconv.Itoa(*number)+" -w "+strconv.Itoa(*timeout)+" "+dstIP+" && echo true || echo false")
	} else if OS == "linux" {
		command = exec.Command("/bin/bash", "-c", "ping -c "+strconv.Itoa(*number)+" -w "+strconv.Itoa(*timeout)+" "+dstIP+" && echo true || echo false")
	} else if OS == "darwin" {
		command = exec.Command("/bin/bash", "-c", "ping -c "+strconv.Itoa(*number)+" -W "+strconv.Itoa(*timeout)+" "+dstIP+" && echo true || echo false")
	}
	outinfo := bytes.Buffer{}
	command.Stdout = &outinfo
	err := command.Start()
	if err != nil {
		return false
	}
	if err = command.Wait(); err != nil {
		return false
	} else if strings.Contains(outinfo.String(), "true") {
		resping := mahonia.NewDecoder("gbk").ConvertString(string(outinfo.String()))
		if strings.Index(resping, "ms") > 0 {
			res := strings.Split(resping, "ms")[0]

			if strings.Index(res, "<") > 0 {
				pingtime = strings.Split(res, "<")[len(strings.Split(res, "<"))-1]
			} else {
				pingtime = strings.Split(res, "=")[len(strings.Split(res, "="))-1]
			}
			fmt.Printf("[ping]  %-16v [%vms]\n", dstIP, pingtime)
		}
		return true
	}

	return false
}

func sendIp(content string) {
	data := fmt.Sprintf(`{
        "msgtype": "text",
        "text": {
            "content": "%s",
            "mentioned_list":["@all"],
        }
    }`, content)

	httpclient(webhook, "POST", data)

}

func main() {
	f1 := flag.NewFlagSet("f1", flag.ContinueOnError)

	host = f1.StringP("host", "h", "", "target ip, eg: 192.168.1.1 192.168.1.1/24 192.168.1.1-100 192.168.1-20.100")
	hostfile = f1.StringP("hostfile", "H", "", "target ip file, eg: ./ip.txt")
	number = f1.IntP("number", "n", 1, "Number of packets sent")
	timeout = f1.IntP("timeout", "w", 1, "IP Timeout for reply")
	install = f1.BoolP("install", "i", false, "install service")
	uninstall = f1.BoolP("uninstall", "u", false, "uninstall service")
	serviceName = f1.StringP("servicename", "I", "Network", "install/uninstall service displayName")
	serviceDescription = f1.StringP("description", "D", "", "install/uninstall service description")

	f1.Usage = func() {
		fmt.Fprintf(os.Stderr, `Name:
  ping Tools

Usage:
  %s [options] [arguments...]

Options:
`, programName)
		f1.PrintDefaults()
	}
	f1.Parse(os.Args[1:])

	if *host == "" && *hostfile == "" {
		f1.Usage()
		fmt.Printf("\nRequired Options: \"host/hostfile\" not set!\n\n")
		os.Exit(0)
	}

	var serviceConfig = &service.Config{
		Name:        *serviceName,
		DisplayName: *serviceName,
		Description: *serviceDescription,
		Arguments:   []string{"-h", *host},
	}

	// 构建服务对象
	prog := &Program{}
	s, err := service.New(prog, serviceConfig)
	if err != nil {
		log.Fatal(err)
	}

	// 用于记录系统日志
	logger, err := s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	if *install {
		err = s.Install()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("安装成功")
	}

	if *uninstall {
		err = s.Uninstall()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("卸载成功")
	}

	if *host != "" && !*install && !*uninstall {
		//Run 调用会调用 prog.start
		err = s.Run()
		if err != nil {
			logger.Error(err)
		}
		return
	}

}
