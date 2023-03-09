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
	"encoding/json"
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
var debug_mode, _ = user_config.Section("").Key("debug_mode").Bool()
var keyServerAddr = user_config.Section("").Key("server_address").String()

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

	var https_port, tls_path_crt, tls_path_key string

	if debug_mode {
		https_port = user_config.Section("devel").Key("https_port").String()
		tls_path_crt = user_config.Section("devel").Key("tls_path_crt").String()
		tls_path_key = user_config.Section("devel").Key("tls_path_key").String()
		fmt.Println("</> Debugging mode </>")
	} else {
		https_port = user_config.Section("prod").Key("https_port").String()
		tls_path_crt = user_config.Section("prod").Key("tls_path_crt").String()
		tls_path_key = user_config.Section("prod").Key("tls_path_key").String()
		fmt.Println("Production mode")
	}

	mux_tls := http.NewServeMux()
	mux_tls.HandleFunc("/", getRoot)

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
	cpu_user := fmt.Sprintf("%.2f%%", float64(after.User-before.User)/total*100)
	cpu_system := fmt.Sprintf("%.2f%%", float64(after.System-before.System)/total*100)
	cpu_idle := fmt.Sprintf("%.2f%%", float64(after.Idle-before.Idle)/total*100)

	// memory
	memory, err := memory.Get()
	if err != nil {
		handleErrors(err)
	}

	memory_total := fmt.Sprintf("%.2f MB", (float64(memory.Total) * float64(0.000001)))
	memory_used := fmt.Sprintf("%.2f MB", (float64(memory.Used) * float64(0.000001)))
	memory_cached := fmt.Sprintf("%.2f MB", (float64(memory.Cached) * float64(0.000001)))
	memory_free := fmt.Sprintf("%.2f MB", (float64(memory.Free) * float64(0.000001)))

	// loadavg
	loadavg, err := loadavg.Parse()
	if err != nil {
		handleErrors(err)
	}

	// OS Release
	OS_release := ReadConfigFile("/etc/os-release").Section("").Key("PRETTY_NAME").String()

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
	var ip_addresses []net.IP

	for _, ip_addr := range netAddrs {
		if ipnet, ok := ip_addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			// check if IPv4 or IPv6 is not nil
			if ipnet.IP.To4() != nil || ipnet.IP.To16 != nil {
				ip_addresses = append(ip_addresses, ipnet.IP)
			}
		}
	}

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

	formatted_date := fmt.Sprintf("%02d/%02d/%02d %02d:%02d", start.Day(), start.Month(), start.Year(), start.Hour(), start.Minute())

	// END --- Gathering information --- END

	if debug_mode {
		fmt.Printf("(DEBUG) %s: got / request. password=%s, body: %s\n",
			ctx.Value(keyServerAddr),
			get_password,
			body)
	}

	// convert to json
	data := map[string]interface{}{
		"hostname":       hostname,
		"cpu user":       cpu_user,
		"cpu system":     cpu_system,
		"cpu idle":       cpu_idle,
		"memory total":   memory_total,
		"memory used":    memory_used,
		"memory cached":  memory_cached,
		"memory free":    memory_free,
		"load average1":  loadavg.LoadAverage1,
		"load average5":  loadavg.LoadAverage5,
		"load average10": loadavg.LoadAverage10,
		"OS":             OS_release,
		"RootFS":         root_fs_info,
		"SimpleX":        simplex_full_address,
		"Date":           formatted_date,
		"IP Addresses":   ip_addresses,
	}

	jsonData, err := json.Marshal(data)

	if err != nil {
		handleErrors(err)
	}

	io.WriteString(w, fmt.Sprintf("%s\n", jsonData))

}
