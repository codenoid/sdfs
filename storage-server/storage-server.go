package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// check is given ip address is valid ip4
func is_ipv4(host string) bool {
	parts := strings.Split(host, ".")

	if len(parts) < 4 {
		return false
	}

	for _, x := range parts {
		if i, err := strconv.Atoi(x); err == nil {
			if i < 0 || i > 255 {
				return false
			}
		} else {
			return false
		}

	}
	return true
}

func standardizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// exportsExist : check is current line config already exist
func exportsExist(line string) (bool, error) {
	b, err := ioutil.ReadFile("/etc/exports")
	if err != nil {
		return false, err
	}

	content := strings.Split(string(b), "\n")

	for _, eline := range content {
		if standardizeSpaces(eline) == line {
			return true, nil
		}
	}

	return false, nil
}

// appendExports : write line config to /etc/exports
func appendExports(line string) error {
	f, err := os.OpenFile("/etc/exports", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err = f.WriteString(line); err != nil {
		return err
	}

	return nil
}

func main() {
	argsWithoutProg := os.Args[1:]

	switch argsWithoutProg[0] {
	case "connect":
		// connect 6.6.6.28 /var/app okloc
		if len(argsWithoutProg) == 4 {
			if is_ipv4(argsWithoutProg[1]) {
				// /path ip(options,,)
				export := "%v %v(rw,async,no_subtree_check)"
				export = fmt.Sprintf(export, argsWithoutProg[2], argsWithoutProg[1])

				fmt.Println("checking if path already in /etc/export")

				check, err := exportsExist(export)

				if err != nil {
					fmt.Println(err)
					return
				}

				if check == false {

					fmt.Println("writing new line to /etc/export")

					if err = appendExports(export); err == nil {
						fmt.Println("restarting nfs-server....")
						exec.Command("/bin/systemctl", "restart", "nfs-server").Run()
					} else {
						fmt.Println(err)
						return
					}
				}

				master_url := fmt.Sprintf("http://%v:2219/api/connect", argsWithoutProg[1])

				fmt.Println("sending path information to master storage")
				resp, err := http.PostForm(master_url, url.Values{"path": {argsWithoutProg[2]}, "id": {argsWithoutProg[3]}})

				if nil != err {
					fmt.Println(err)
					return
				}

				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)

				if nil != err {
					fmt.Println(err)
					return
				}

				fmt.Println("reloading exportfs")

				exec.Command("/usr/sbin/exportfs", "-ra").Run()

				fmt.Println(string(body[:]))
			}
		}
	}
}
