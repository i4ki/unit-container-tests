package main

import (
    "fmt"
    "net/http"
    "nginx/unit"
    "os"
    "path/filepath"
    "syscall"
)

func abortonerr(ctx string, err error) {
    if err != nil {
        panic(fmt.Errorf("%s: %s", ctx, err))
    }
}

func changeRoot(rootfs string) {
    err := syscall.Mount("proc", filepath.Join(rootfs, "/proc"), "proc",
        uintptr(0), "")
    abortonerr("mount proc", err)

    err = syscall.Mount(rootfs, rootfs, "", syscall.MS_BIND|syscall.MS_REC, "")
    abortonerr("bind mount", err)

    err = os.Chdir(rootfs)
    abortonerr("chdir", err)

    err = syscall.Mount("", ".", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
    abortonerr("recursive private mount", err)

    err = syscall.PivotRoot(".", ".")
    abortonerr("pivot_root: %s", err)

    err = syscall.Unmount(".", syscall.MNT_DETACH)
    abortonerr("unmount", err)

    err = os.Chdir("/")
    abortonerr("cd /", err)
}

func handler(w http.ResponseWriter, r *http.Request) {
    wdir, err := os.Getwd()
    abortonerr("cwd", err)

    rootfs := filepath.Join(wdir, "rootfs")

    if _, err := os.Stat(filepath.Join(rootfs, "bin", "sh")); err == nil {
        changeRoot(rootfs)
    }

    fmt.Fprintf(w, "PID: %d\n", os.Getpid())
    fmt.Fprintf(w, "UID: %d\n", os.Getuid())
    fmt.Fprintf(w, "GID: %d\n", os.Getgid())
}

func main() {
    http.HandleFunc("/", handler)
    unit.ListenAndServe(":7080", nil)
}
