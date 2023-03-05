/* big thanks to
 * https://www.digitalocean.com/community/tutorials/how-to-make-an-http-server-in-go
 * https://www.digitalocean.com/community/tutorials/how-to-use-json-in-go
TODO
   - all info as functions
   - better writing
   - remove unused stuff, add comments
   - return JSON result
*/

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
	"unicode"

	human_df "github.com/dustin/go-humanize"
	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
	"github.com/mikoim/go-loadavg"
	"github.com/shirou/gopsutil/disk"
	"gopkg.in/ini.v1"
)

var user_config = ReadConfigFile("config.ini")

var keyServerAddr = user_config.Section("server").Key("server_address").String()

func ReadConfigFile(configfile string) *ini.File {
	cfg, err := ini.Load(configfile)

	if err != nil {
		handleErrors(err)
	}

	return cfg
}

func removeSpace(s string) string {
	rr := make([]rune, 0, len(s))
	for _, r := range s {
		if !unicode.IsSpace(r) {
			rr = append(rr, r)
		}
	}
	return string(rr)
}

func handleErrors(err error) {
	fmt.Fprintf(os.Stderr, "Failed: %s\n", err)
	os.Exit(1)
}

func redirect(w http.ResponseWriter, r *http.Request) {

	http.Redirect(w, r, "https://"+keyServerAddr, 301)
}

func main() {

	debug_mode, _ := user_config.Section("").Key("debug_mode").Bool()
	//	http_port := user_config.Section("server").Key("http_port").String()
	https_port := user_config.Section("server").Key("https_port").String()
	tls_path_crt := user_config.Section("server").Key("tls_path_crt").String()
	tls_path_key := user_config.Section("server").Key("tls_path_key").String()

	if debug_mode {
		fmt.Println("</> Debugging enabled </>")
	}

	mux_tls := http.NewServeMux()
	mux_tls.HandleFunc("/", getRoot)
	mux_tls.HandleFunc("/nick", getNick)

	ctx_tls := context.Background()

	server_tls := &http.Server{
		Addr:    ":" + https_port,
		Handler: mux_tls,
		BaseContext: func(l net.Listener) context.Context {
			ctx_tls = context.WithValue(ctx_tls, keyServerAddr, l.Addr().String())
			return ctx_tls
		},
	}

	server_err_tls := server_tls.ListenAndServeTLS(tls_path_crt, tls_path_key)

	if errors.Is(server_err_tls, http.ErrServerClosed) {
		fmt.Printf("server closed\n")

	} else if server_err_tls != nil {
		handleErrors(server_err_tls)
	}

}

func getRoot(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")
	ctx := r.Context()
	get_password := r.URL.Query().Get("password")

	body, body_err := io.ReadAll(r.Body)

	if body_err != nil {
		handleErrors(body_err)
	}

	debug_mode, _ := user_config.Section("").Key("debug_mode").Bool()

	// START --- Gathering information --- START
	start := time.Now()

	// hostname
	hostname, _ := os.Hostname()

	// cpu
	before, err := cpu.Get()
	if err != nil {
		handleErrors(err)
	}

	time.Sleep(time.Duration(1) * time.Second)

	after, err := cpu.Get()
	if err != nil {
		handleErrors(err)
	}

	total := float64(after.Total - before.Total)

	// memory
	memory, err := memory.Get()
	if err != nil {
		handleErrors(err)
	}

	// loadavg
	loadavg, err := loadavg.Parse()
	if err != nil {
		handleErrors(err)
	}

	// OS Release
	OS_release := ReadConfigFile("/etc/os-release").Section("").Key("PRETTY_NAME")

	// Disk size
	target_mountpoint := user_config.Section("").Key("target_mountpoint").String()
	root_fs_info := []string{}

	parts, _ := disk.Partitions(true)
	for _, p := range parts {
		device := p.Mountpoint
		s, _ := disk.Usage(device)

		if s.Total == 0 {
			continue
		}

		percent := fmt.Sprintf("%2.f%%", s.UsedPercent)

		if p.Mountpoint == target_mountpoint {

			root_fs_info = append(root_fs_info,
				s.Fstype,
				human_df.Bytes(s.Total),
				human_df.Bytes(s.Used),
				human_df.Bytes(s.Free),
				percent,
				p.Mountpoint,
			)

		}
	}

	// IP Addresses
	netAddrs, err := net.InterfaceAddrs()

	if err != nil {
		handleErrors(err)
	}
	// Read SimpleX chat fingerprint address
	simplex_addr, err := os.ReadFile("/etc/opt/simplex/fingerprint")

	if err != nil {
		handleErrors(err)
	}

	simplex_fingerprint := removeSpace(string(simplex_addr))
	simplex_full_address := "smp://" + simplex_fingerprint + ":PASSWORD@" + hostname

	// Execution time
	elapsed := time.Since(start)
	// END --- Gathering information --- END

	if debug_mode {
		fmt.Printf("(DEBUG) %s: got / request. password=%s, body: %s\n",
			ctx.Value(keyServerAddr),
			get_password,
			body)
	}

	io.WriteString(w, fmt.Sprintf("hostname: %s\n", hostname))
	io.WriteString(w, fmt.Sprintf("\n---ROOTFS%vROOTFS---\n", root_fs_info))
	io.WriteString(w, fmt.Sprintf("cpu user: %.2f %%\n", float64(after.User-before.User)/total*100))
	io.WriteString(w, fmt.Sprintf("cpu system: %.2f %%\n", float64(after.System-before.System)/total*100))
	io.WriteString(w, fmt.Sprintf("cpu idle: %.2f %%\n", float64(after.Idle-before.Idle)/total*100))

	io.WriteString(w, fmt.Sprintf("memory total: %.2f MB\n", (float64(memory.Total)*float64(0.000001))))
	io.WriteString(w, fmt.Sprintf("memory used: %.2f MB\n", (float64(memory.Used)*float64(0.000001))))
	io.WriteString(w, fmt.Sprintf("memory cached: %.2f MB\n", (float64(memory.Cached)*float64(0.000001))))
	io.WriteString(w, fmt.Sprintf("memory free: %.2f MB\n", (float64(memory.Free)*float64(0.000001))))
	io.WriteString(w, fmt.Sprintf("load average: %.2f %.2f %.2f\n", loadavg.LoadAverage1, loadavg.LoadAverage5, loadavg.LoadAverage10))
	io.WriteString(w, fmt.Sprintf("OS: %s\n", OS_release))

	for _, ip_addr := range netAddrs {
		io.WriteString(w, fmt.Sprintf("IP: %s\n", ip_addr))
	}

	io.WriteString(w, fmt.Sprintf("Time to gather info: %s\n", elapsed))
	io.WriteString(w, fmt.Sprintf("Date: %02d/%02d/%02d %02d:%02d\n", start.Day(), start.Month(), start.Year(), start.Hour(), start.Minute()))
	io.WriteString(w, fmt.Sprintf("SimpleX: %s\n", simplex_full_address))
}

func getNick(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	debug_mode, _ := user_config.Section("").Key("debug_mode").Bool()

	if debug_mode {
		fmt.Printf("(DEBUG) %s: got /nick request\n", ctx.Value(keyServerAddr))
	}

	/*
		myName := r.PostFormValue("myName")
		if myName == "" {
			w.Header().Set("x-missing-field", "myName")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	*/
	nick_task, err := os.ReadFile("/opt/nick_task")

	if err != nil {
		handleErrors(err)
	}

	io.WriteString(w, fmt.Sprintf("%s\n", nick_task))
}
