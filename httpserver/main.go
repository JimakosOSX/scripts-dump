/* A program that aims to return a few basic system info
 * Targeted towards Linux. Tested under Ubuntu.
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

// some global variables
var user_config = read_Config_file("config.ini")
var debug_mode, _ = user_config.Section("").Key("debug_mode").Bool()
var key_server_addr = user_config.Section("").Key("server_address").String()

// read configuration files.
func read_Config_file(configfile string) *ini.File {
	cfg, err := ini.Load(configfile)

	if err != nil {
		handle_Errors(err)
	}

	return cfg
}

// remove spaces from a string
func remove_Space(s string) string {
	rr := make([]rune, 0, len(s))
	for _, r := range s {
		if !unicode.IsSpace(r) {
			rr = append(rr, r)
		}
	}
	return string(rr)
}

// a standard way to handle errors
func handle_Errors(err error) {
	fmt.Fprintf(os.Stderr, "Failed: %s\n", err)
	os.Exit(1)
}

// redirect http traffic to https
func redirect(w http.ResponseWriter, r *http.Request) {

	http.Redirect(w, r, "https://"+key_server_addr, 301)
}

// main starts
func main() {

	var target_env string

	if debug_mode {
		target_env = "devel"
		fmt.Println("</> Debugging mode </>")

	} else {
		target_env = "prod"
		fmt.Println("Production mode")
	}

	https_port := user_config.Section(target_env).Key("https_port").String()
	tls_path_crt := user_config.Section(target_env).Key("tls_path_crt").String()
	tls_path_key := user_config.Section(target_env).Key("tls_path_key").String()

	mux_tls := http.NewServeMux()
	mux_tls.HandleFunc("/", get_webRoot)

	ctx_tls := context.Background()

	server_tls := &http.Server{
		Addr:    ":" + https_port,
		Handler: mux_tls,
		BaseContext: func(l net.Listener) context.Context {
			ctx_tls = context.WithValue(ctx_tls, key_server_addr, l.Addr().String())
			return ctx_tls
		},
	}

	server_err_tls := server_tls.ListenAndServeTLS(tls_path_crt, tls_path_key)

	if errors.Is(server_err_tls, http.ErrServerClosed) {
		fmt.Printf("server closed\n")

	} else if server_err_tls != nil {
		handle_Errors(server_err_tls)
	}

}

// here, every info we need gets collected. Then, they mershal inside a json
func collect_info() []byte {

	start := time.Now()

	// hostname
	hostname, _ := os.Hostname()

	// cpu
	before, err := cpu.Get()
	if err != nil {
		handle_Errors(err)
	}

	time.Sleep(time.Duration(1) * time.Second)

	after, err := cpu.Get()
	if err != nil {
		handle_Errors(err)
	}

	total := float64(after.Total - before.Total)
	cpu_user := fmt.Sprintf("%.2f%%", float64(after.User-before.User)/total*100)
	cpu_system := fmt.Sprintf("%.2f%%", float64(after.System-before.System)/total*100)
	cpu_idle := fmt.Sprintf("%.2f%%", float64(after.Idle-before.Idle)/total*100)

	// memory
	memory, err := memory.Get()
	if err != nil {
		handle_Errors(err)
	}

	memory_total := fmt.Sprintf("%.2f MB", (float64(memory.Total) * float64(0.000001)))
	memory_used := fmt.Sprintf("%.2f MB", (float64(memory.Used) * float64(0.000001)))
	memory_cached := fmt.Sprintf("%.2f MB", (float64(memory.Cached) * float64(0.000001)))
	memory_free := fmt.Sprintf("%.2f MB", (float64(memory.Free) * float64(0.000001)))

	// loadavg
	loadavg, err := loadavg.Parse()
	if err != nil {
		handle_Errors(err)
	}

	// OS Release
	OS_release := read_Config_file("/etc/os-release").Section("").Key("PRETTY_NAME").String()

	// Disk size
	target_mountpoint := user_config.Section("").Key("target_mountpoint").String()

	root_fs_info := make(map[string]string)

	parts, _ := disk.Partitions(true)
	for _, p := range parts {

		device := p.Mountpoint
		s, _ := disk.Usage(device)

		// if partition is unused, or partition is not root, partition is skipped.
		if s.Total == 0 || p.Mountpoint != target_mountpoint {
			continue
		}

		percent := fmt.Sprintf("%2.f%%", s.UsedPercent)

		root_fs_info = map[string]string{

			"size_total":   human_df.Bytes(s.Total),
			"size_used":    human_df.Bytes(s.Used),
			"size_free":    human_df.Bytes(s.Free),
			"percent_used": percent,
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
		handle_Errors(err)
	}

	// Read SimpleX chat fingerprint address
	simplex_addr, err := os.ReadFile("/etc/opt/simplex/fingerprint")

	if err != nil {
		handle_Errors(err)
	}

	simplex_fingerprint := remove_Space(string(simplex_addr))
	simplex_full_address := "smp://" + simplex_fingerprint + ":PASSWORD@" + hostname

	formatted_date := fmt.Sprintf("%02d/%02d/%02d %02d:%02d", start.Day(), start.Month(), start.Year(), start.Hour(), start.Minute())

	// convert all data to json
	data := map[string]interface{}{
		"hostname":      hostname,
		"cpu_user":      cpu_user,
		"cpu_system":    cpu_system,
		"cpu_idle":      cpu_idle,
		"memory_total":  memory_total,
		"memory_used":   memory_used,
		"memory_cached": memory_cached,
		"memory_free":   memory_free,
		"load_avg_1":    loadavg.LoadAverage1,
		"load_avg_5":    loadavg.LoadAverage5,
		"load_avg_10":   loadavg.LoadAverage10,
		"OS":            OS_release,
		"Root_FS":       root_fs_info,
		"SimpleX":       simplex_full_address,
		"Date":          formatted_date,
		"IP Addresses":  ip_addresses,
	}

	jsonData, err := json.Marshal(data)

	if err != nil {
		handle_Errors(err)
	}

	// finally, return the data
	return jsonData
}

// handles any web requests under /
func get_webRoot(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	get_password := r.URL.Query().Get("password")

	body, body_err := io.ReadAll(r.Body)

	if body_err != nil {
		handle_Errors(body_err)
	}

	if debug_mode {
		fmt.Printf("(DEBUG) %s: got / request. password=%s, body: %s\n",
			ctx.Value(key_server_addr),
			get_password,
			body)
	}

	json_result := collect_info()
	io.WriteString(w, fmt.Sprintf("%s\n", json_result))
}
