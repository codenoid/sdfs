package main

import (
    "os"
    "fmt"
    "strings"
    "os/exec"
    "net/http"
    "path/filepath"
)

import (
    "./helper"
    _"./slave"
)

func symlink(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()

    fpath := r.FormValue("path")

    if fpath != "" {
        brick, err := helper.AvailableBrick()
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }
        fi, err := os.Lstat(fpath)
        
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }

        if fi.Mode() & os.ModeSymlink == os.ModeSymlink {
            // symlinked file
        } else {
            // /data/bricl1/
            bpath := brick + fpath
            
            dirOnly := strings.Split(bpath, "/")
            dirOnly  = dirOnly[:len(dirOnly)-1]

            os.MkdirAll(strings.Join(dirOnly[:],"/"), os.ModePerm)

            helper.MoveFile(fpath, bpath)

            helper.Symlink(bpath, fpath)            
        }
    }

    w.WriteHeader(200)

    fmt.Fprintln(w, ":OK")
}

func connect(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()

    path := r.FormValue("path")
    uid := r.FormValue("id")

    if path != "" && uid != "" {
        mpath := filepath.Join("/data", uid)
        err := os.MkdirAll(mpath, os.ModePerm)
        if err != nil {
            http.Error(w, "Problem with folder creation", 500)
            return
        }

        // 30.30.30.12:/storage/wamon  /mnt/storage/wamon  nfs auto,_netdev,rw,hard,intr 0 0
        fstab := "%v:%v %v nfs auto,_netdev,rw,hard,intr 0 0"
        fstab = fmt.Sprintf(fstab, strings.Split(r.RemoteAddr, ":")[0], path, mpath)

        check, err := helper.ExportsExist(fstab)

        if err == nil {
            if check == false {
                if err = helper.AppendExports(fstab); err == nil {
                    exec.Command("/bin/mount", "-a").Run()
                }
            }
        }
    }

    w.WriteHeader(200)

    fmt.Fprintln(w, ":OK")
}

func main() {
    http.HandleFunc("/api/symlink", symlink)
    http.HandleFunc("/api/connect", connect)
    
    fmt.Println("Starting server ...")

    http.ListenAndServe("0.0.0.0:2219", nil)
}