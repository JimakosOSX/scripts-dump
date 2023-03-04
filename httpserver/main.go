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
    "github.com/mackerelio/go-osstat/cpu"
    "github.com/mackerelio/go-osstat/memory"
//    "github.com/mackerelio/go-osstat/loadavg"
)

const keyServerAddr = "clusters.gr"
var DEBUG, IS_DEBUG = os.LookupEnv("DEBUG")
/*
   { "DISTRO": distro_pretty },
   { "HDD": hdd },
   { "LOAD": loadavg_pretty },
   { "IP": ip_list_prettier },
   { "EXEC_TIME": execution_time_str },
   { "USER_AGENT": user_agent },
   { "DATE": time_pretty },
   { "SMP_ADDRESS": smp_server_address },
*/

func main() {

    if DEBUG != "" { 
        fmt.Printf("DEBUG = %s\n", DEBUG)
    }

	mux := http.NewServeMux()
	mux.HandleFunc("/", getRoot)
	mux.HandleFunc("/hello", getHello)

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

func getRoot(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	get_fqdn := r.URL.Query().Get("FQDN")

	body, err := io.ReadAll(r.Body)

	if err != nil {
		fmt.Printf("Could not read body: %s\n", err)
	}

    // START --- Gathering information --- START

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
}

func getHello(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

    
    if IS_DEBUG {
        fmt.Printf("(DEBUG) %s: got /hello request\n", ctx.Value(keyServerAddr))
    }

	myName := r.PostFormValue("myName")
	if myName == "" {
		w.Header().Set("x-missing-field", "myName")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	io.WriteString(w, fmt.Sprintf("hello, %s\n", myName))
}
