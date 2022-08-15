package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	Version     = "v1.0.0"
	ServiceName = "go_delayer_test"
	HarborUrl   = "harbor.bellonline.cn/bellcloud"
	ImageNames  = HarborUrl + "/" + ServiceName

	SSHKeyPath = "/home/chenyuan/.ssh/id_rsa"
)

var cmd = fmt.Sprintf(`
		cd service_docker/go_delayer;
		docker-compose down;
		docker rmi %s;
		docker-compose up -d`, ImageNames)

func main() {
	if strings.Contains(ServiceName, "test") {
		fmt.Println("包含test，部署测试环境")
		err := deploy()
		if err != nil {
			fmt.Println(err)
		}
		developService("root", "121.196.29.23", 45012)
	} else {
		//如果部署则直接退出
		cmd := exec.Command("git", "branch")
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(err)
		}
		result := strings.Split(string(out), "\n")
		if len(result) == 0 {
			fmt.Println("分支查找失败！")
			return
		}
		for _, item := range result {
			if item == "* master" {
				//检查当前的git branch 分支是否是master //如果是则直接部署
				err := deploy()
				if err != nil {
					fmt.Println(err)
				}
				break
			}
		}
	}
}

func deploy() error {
	fmt.Println("部署中...")
	var cmd *exec.Cmd
	if strings.Contains(whichOS(), "linux") {
		cmd = exec.Command("bash", "build-docker.sh", HarborUrl, ServiceName, Version)
	} else {
		cmd = exec.Command("cmd.exe", "/c", "build-docker.bat", HarborUrl, ServiceName, Version)
	}
	printLog(cmd)
	return nil
}

func printLog(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
}

// 部署服务
func developService(user, host string, port int) {
	fmt.Println("开始部署服务器 " + host + " !")
	session, err := connect(user, host, port)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Run(cmd)
	fmt.Println("结束部署服务器 " + host + " !")
}

// 使用ssh连接
func connect(user, host string, port int) (*ssh.Session, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		client       *ssh.Client
		session      *ssh.Session
		err          error
	)
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, publicKeyAuthFunc(SSHKeyPath))

	clientConfig = &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: 30 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	addr = fmt.Sprintf("%s:%d", host, port)
	if client, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	if session, err = client.NewSession(); err != nil {
		return nil, err
	}
	return session, nil
}

func whichOS() (goos string) {
	var osb []byte
	cmd := exec.Command("go", "env", "GOOS")
	osb, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	goos = string(osb)
	return
}

func publicKeyAuthFunc(kPath string) ssh.AuthMethod {
	//keyPath, err := homedir.Expand(kPath)
	//if err != nil {
	//	log.Fatal("find key's home dir failed", err)
	//}
	key, err := ioutil.ReadFile(kPath)
	if err != nil {
		log.Fatal("ssh key file read failed", err)
	}
	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatal("ssh key signer failed", err)
	}
	return ssh.PublicKeys(signer)
}
