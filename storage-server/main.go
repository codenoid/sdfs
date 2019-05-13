package main

import(
	"os"
	"fmt"
	"strings"
	"strconv"
	"os/exec"
	"net/url"
	"net/http"
	"io/ioutil"
)

func is_ipv4(host string) bool {
	parts := strings.Split(host, ".")

	if len(parts) < 4 {
		return false
	}
	
	for _,x := range parts {
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
		// connect 30.30.30.11 /var/app okloc
		if len(argsWithoutProg) == 4 {
			if is_ipv4(argsWithoutProg[1]) {
				// /path ip(options,,)
				export := "%v %v(rw,async,no_subtree_check)"
				export = fmt.Sprintf(export, argsWithoutProg[2], argsWithoutProg[1])

				check, err := exportsExist(export)

				if err == nil {
					if check == false {
						if err = appendExports(export); err == nil {
							exec.Command("/bin/systemctl", "restart", "nfs-server").Run()
						}
					}
				}

				master_url := fmt.Sprintf("http://%v:2219/api/connect", argsWithoutProg[1])

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

				fmt.Println(string(body[:]))
			}
		}
	}
}