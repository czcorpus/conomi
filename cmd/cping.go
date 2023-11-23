package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/czcorpus/cnc-gokit/fs"
	"github.com/czcorpus/conomi/client"
	"github.com/czcorpus/conomi/cnf"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Conomi Ping - testing server availability\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [config.json] [auth token]\n\t", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()
	if !fs.PathExists(flag.Arg(0)) {
		fmt.Fprintf(os.Stderr, "Failed to load Conomi config\n")
		flag.Usage()
		os.Exit(1)
	}
	conf := cnf.LoadConfig(flag.Arg(0))
	ct := client.NewConomiClient(client.ConomiClientConf{
		Server:   fmt.Sprintf("http://%s:%d", conf.ListenAddress, conf.ListenPort),
		App:      "test",
		Instance: "cmd",
		APIToken: flag.Arg(1),
	})
	err := ct.Ping()
	fmt.Println("--------------------------------------------------")
	if err != nil {
		fmt.Printf("finished with error: %s\n", err)
		return
	}
	fmt.Println("finished without errors")
	fmt.Println("--------------------------------------------------")
}
