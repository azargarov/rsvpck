package main

import (
	"github.com/azargarov/go-utils/autostr"
	"github.com/azargarov/rsvpck/internal/adapters/dns"
	"github.com/azargarov/rsvpck/internal/adapters/hostinfo"
	"github.com/azargarov/rsvpck/internal/adapters/http"
	"github.com/azargarov/rsvpck/internal/adapters/httpx"
	"github.com/azargarov/rsvpck/internal/adapters/icmp"
	"github.com/azargarov/rsvpck/internal/adapters/render/text"
	"github.com/azargarov/rsvpck/internal/adapters/tcp"
	"github.com/azargarov/rsvpck/internal/config"
	"github.com/azargarov/rsvpck/internal/app"
	"github.com/azargarov/rsvpck/internal/domain"
	//"github.com/azargarov/rsvpck/internal/adapters/speedtest"

	"context"
	"fmt"
	"os"
	"time"
	"io"
	"runtime"
	"bufio"
)
//TODO: if TLS is stuck thre rest fails due to timeout.
var version = "dev"

const (
    applicationName = "RSvP connectivity checker"
	totalTimeout = 300*time.Second
)

func main() {
	
	rsvpConf := parseFlagsToConfig()
	if rsvpConf.printVersion{
		fmt.Printf("%s, version %s\n", applicationName, version)
		return
	}

	printHeader()
	

	var renderer domain.Renderer
	renderConf := text.NewRenderConfig(text.WithForceASCII(rsvpConf.forceASCII))
	
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), totalTimeout)
	defer cancel()

	//if rsvpConf.speedtest{
	//	res :=runSpeedTest(ctx)
	//	if res != nil{
	//		fmt.Println(res.String())
	//	}
	//	return
	//}

	testConfig, err := config.LoadEmbedded()
	if err != nil {
		fmt.Printf("Invalid config: %v", err)
		return
	}
	
	tcpChecker := &tcp.Checker{}
	dnsChecker := &dns.Checker{}
	httpChecker := &http.Checker{}
	icmpChecker := &icmp.Checker{}

	prober := app.NewCompositeProber(tcpChecker,dnsChecker, httpChecker, icmpChecker)
	stopSpinner := startAnimatedSpinner(os.Stdout, ctx, 120 * time.Millisecond)
	executor := app.NewExecutor(prober, domain.PolicyExhaustive)
	result := executor.Run(ctx, testConfig)
	stopSpinner()

	if result.IsConnected {
		h := hostinfo.GetCRMInfo(ctx)
		autostrCfg := autostr.Config{Separator: autostr.Ptr("\n"), FieldValueSeparator: autostr.Ptr(" : "), PrettyPrint: true}

		text.PrintBlock(os.Stdout, "SYSTEM INFORMATION", autostr.String(h, autostrCfg), renderConf)
		h.TLSCert, err = httpx.GetCertificatesSmart(ctx, "insite-eu.gehealthcare.com:443", "insite-eu.gehealthcare.com", testConfig.VPNIPs)
		if err == nil {
			text.PrintList(os.Stdout, "TLS certificates, eu-insite.gehealthcare.com\n", h.TLSCert, renderConf)
		} else {
			fmt.Println("Failed fetching certificates")
		}
	}


	if rsvpConf.textRender {
		renderer = text.NewRenderer(renderConf)
		if err := renderer.Render(os.Stdout, result); err != nil {
			fmt.Printf("Failed to render: %v", err)
		}
	} else {
		renderer = text.NewTableRenderer(renderConf)
		if err := renderer.Render(os.Stdout, result); err != nil {
			fmt.Printf("Failed to render: %v", err)
		}
	}
	waitForEnterOnWindows()
}

func startAnimatedSpinner(w io.Writer, parent context.Context, interval time.Duration) (stop func()) {
	ctx, cancel := context.WithCancel(parent)
	done := make(chan struct{})

	//frames := []string{"|", "/", "-", "\\"} 
	frames := []string{"      ", ".", "..", "...", "....",".....","......", " .....", "  ....", "   ...","    ..","     ."}

	// Hide cursor
	fmt.Fprint(w, "\x1b[?25l")

	go func() {
		t := time.NewTicker(interval)
		defer func() {
			t.Stop()
			fmt.Fprint(w, "\r \r") // clear spinner
			fmt.Fprint(w, "\x1b[?25h")
			close(done)
		}()
		i := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				fmt.Fprintf(w, "\r%s", frames[ i % len(frames) ])
				i++
			}
		}
	}()
	return func() {
		cancel()
		<-done
	}
}

func printHeader() {
	fmt.Println("\nRSVP CHECK - Connectivity Diagnostics")
	fmt.Println("-------------------------------------")
}

func waitForEnterOnWindows() {
	if runtime.GOOS == "windows" {
		fmt.Println("\nPress Enter to exit...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
}

//func runSpeedTest(ctx context.Context) *speedtest.SpeedtestResult{
//	fmt.Println("Test Network Speed")
//	spTest := speedtest.SpeedtestChecker{}
//	result, err :=spTest.Run(ctx)
//	if err != nil{
//		return nil
//	}
//
//	return result
//}

//docker run --rm -ti -v "$PWD":/app -w/app golang:1.23-alpine sh
//apk add build-base
//CGO_ENABLE=1 go build -tags netgo -o rsvpck ./cmd/rsvpck/*.go

//docker run --rm -ti -v "$PWD":/app -w /app golang:1.23-alpine sh -lc '
//  apk add --no-cache upx zip git &&
//  CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=v0.1.0" -o dist/rsvpck &&
//  upx --best --lzma dist/rsvpck || true &&
//  tar czf dist/rsvpck_linux_amd64.tar.gz -C dist rsvpck &&
//  sha256sum dist/* > dist/SHA256SUMS.txt &&
//  ls -lh dist
//'
