package available

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

func doCheck(host string) (map[string]any, error) {
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		},
		Timeout: time.Second,
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/version", host), http.NoBody)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(context.Background())
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	var resJSON = map[string]any{}
	if err := json.NewDecoder(res.Body).Decode(&resJSON); err != nil {
		return nil, err
	}
	return resJSON, nil
}

func Main(cmd *cobra.Command, args []string) {
	host := "https://localhost:8080"
	if len(args) > 0 {
		host = args[0]
	}

	doWait, _ := cmd.Flags().GetBool("wait")
	if !doWait {
		res, err := doCheck(host)
		if err != nil {
			fmt.Println("Server not available")
			os.Exit(1)
		}
		fmt.Println(res)
		os.Exit(1)
	}

	for {
		res, err := doCheck(host)
		if err == nil {
			fmt.Println("Server", host, "is available")
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(res)
			os.Exit(1)
		}
		fmt.Println("Server", host, "is not available")
		time.Sleep(time.Second)
	}
}
