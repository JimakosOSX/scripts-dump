/* big thanks to 
 * https://www.digitalocean.com/community/tutorials/how-to-make-an-http-server-in-go 
 * https://www.digitalocean.com/community/tutorials/how-to-use-json-in-go
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
    "github.com/mackerelio/go-osstat/cpu"
    "github.com/mackerelio/go-osstat/memory"
    "github.com/mikoim/go-loadavg"
    human_df "github.com/dustin/go-humanize"
    "github.com/shirou/gopsutil/disk"
    "github.com/go-ini/ini"
)

const keyServerAddr = "clusters.gr"
var DEBUG, IS_DEBUG = os.LookupEnv("DEBUG")
/* 
TODO
   { "USER_AGENT": user_agent },
   - all info as functions
   - better writing
   - remove unused stuff, add comments
   - return JSON result
   - read an ini file to get parameters eg debug
*/

    
func removeSpace(s string) string {
	rr := make([]rune, 0, len(s))
	for _, r := range s {
		if !unicode.IsSpace(r) {
			rr = append(rr, r)
		}
	}
	return string(rr)
}

func main() {
    
    if DEBUG != "" { 
        fmt.Printf("DEBUG = %s\n", DEBUG)
    }

	mux := http.NewServeMux()
	mux.HandleFunc("/", getRoot)
	mux.HandleFunc("/nick", getNick)

	ctx := context.Background()
	server := &http.Server{
		Addr:    ":80",
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			ctx = context.WithValue(ctx, keyServerAddr, l.Addr().String())
			return ctx
		},
	}

	err := server.ListenAndServe()

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")

	} else if err != nil {
		fmt.Printf("error listening for server: %s\n", err)
	}

    	// err := http.ListenAndServe(":80", mux)

	// if errors.Is(err, http.ErrServerClosed) {
	// 	fmt.Printf("server closed\n")

	// } else if err != nil {
	// 	fmt.Printf("error starting server: %s\n", err)
	// 	os.Exit(1)
	// }
}
 
func ReadOSRelease(configfile string, target_key string) string {
    cfg, err := ini.Load(configfile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fail to read file: %s\n", err)
        os.Exit(1)
    }

    ConfigParams := make(map[string]string)
    ConfigParams[target_key] = cfg.Section("").Key(target_key).String()

    return ConfigParams[target_key]
}

func getRoot(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	get_fqdn := r.URL.Query().Get("FQDN")

	body, err := io.ReadAll(r.Body)

	if err != nil {
		fmt.Printf("Could not read body: %s\n", err)
	}

    // START --- Gathering information --- START
    start := time.Now()

    // hostname
    hostname, _ := os.Hostname()

    // cpu
    before, err := cpu.Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}

	time.Sleep(time.Duration(1) * time.Second)
	after, err := cpu.Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	total := float64(after.Total - before.Total)
    
    // memory
    memory, err := memory.Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}

    // loadavg
    loadavg, err := loadavg.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
        return 
	}

    // OS Release
    OSRelease := ReadOSRelease("/etc/os-release", "PRETTY_NAME")

    // Disk size
    formatter := "%-14s %7s %7s %7s %4s %s\n"
    io.WriteString(w, fmt.Sprintf(formatter, "Filesystem", "Size", "Used", "Avail", "Use%", "Mounted on"))

    parts, _ := disk.Partitions(true)
    for _, p := range parts {
        device := p.Mountpoint
        s, _ := disk.Usage(device)

        if s.Total == 0 {
            continue
        }

        percent := fmt.Sprintf("%2.f%%", s.UsedPercent)

        if p.Mountpoint == "/" {

            io.WriteString(w, fmt.Sprintf(formatter,
                s.Fstype,
                human_df.Bytes(s.Total),
                human_df.Bytes(s.Used),
                human_df.Bytes(s.Free),
                percent,
                p.Mountpoint,
            ))
        }
    }

   // IP Addresses
   netAddrs, err := net.InterfaceAddrs()
 
   if err != nil {
       fmt.Fprintf(os.Stderr, "Error getting IP addresses: %s\n", err)
       os.Exit(1)
   }
   // Read SimpleX chat fingerprint address
   simplex_addr, err := os.ReadFile("/etc/opt/simplex/fingerprint")

   if err != nil {
       fmt.Fprintf(os.Stderr, "SimpleX: Error opening file: %s\n", err)
       os.Exit(1)
   }

   simplex_fingerprint := removeSpace( string(simplex_addr) )
   simplex_full_address := "smp://" + simplex_fingerprint + ":PASSWORD@" + hostname

   // Execution time
   elapsed := time.Since(start)
   // END --- Gathering information --- END

    if IS_DEBUG {
        fmt.Printf("(DEBUG) %s: got / request. first=%s, body: %s\n",
        ctx.Value(keyServerAddr),
		get_fqdn,
		body)
    }
     

	io.WriteString(w, fmt.Sprintf("hostname = %s\n", hostname))

    io.WriteString(w, fmt.Sprintf("cpu user: %.2f %%\n", float64(after.User-before.User)/total*100))
	io.WriteString(w, fmt.Sprintf("cpu system: %.2f %%\n", float64(after.System-before.System)/total*100))
	io.WriteString(w, fmt.Sprintf("cpu idle: %.2f %%\n", float64(after.Idle-before.Idle)/total*100))

    io.WriteString(w, fmt.Sprintf("memory total: %.2f MB\n", (float64(memory.Total) * float64(0.000001))))
	io.WriteString(w, fmt.Sprintf("memory used: %.2f MB\n", (float64(memory.Used) * float64(0.000001))))
	io.WriteString(w, fmt.Sprintf("memory cached: %.2f MB\n", (float64(memory.Cached) * float64(0.000001))))
	io.WriteString(w, fmt.Sprintf("memory free: %.2f MB\n", (float64(memory.Free) * float64(0.000001))))
	io.WriteString(w, fmt.Sprintf("load average: %.2f %.2f %.2f\n", loadavg.LoadAverage1, loadavg.LoadAverage5, loadavg.LoadAverage10))
    io.WriteString(w, fmt.Sprintf("OS: %s\n", OSRelease))

    for _, ip_addr := range netAddrs {
        io.WriteString(w, fmt.Sprintf("IP: %s\n", ip_addr))
      }

    io.WriteString(w, fmt.Sprintf("Time to gather info: %s\n", elapsed))
    io.WriteString(w, fmt.Sprintf("Date: %02d/%02d/%02d %02d:%02d\n", start.Day(), start.Month(), start.Year(), start.Hour(), start.Minute()))
    io.WriteString(w, fmt.Sprintf("SimpleX: %s\n", simplex_full_address))
}

func getNick(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

    
    if IS_DEBUG {
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
       fmt.Fprintf(os.Stderr, "Error: %s\n", err)
       os.Exit(1)
   }


   io.WriteString(w, fmt.Sprintf("%s\n", nick_task))
}
