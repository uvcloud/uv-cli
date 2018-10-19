package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"

	"google.golang.org/grpc/codes"

	"github.com/Sirupsen/logrus"
	uvApi "github.com/uvcloud/uv-api-go/proto"

	"google.golang.org/grpc/status"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	//For windows support
	"github.com/mattn/go-colorable"
)

var log = logrus.New()

func init() {
	//Configure logging formatter
	customFormatter := new(logrus.TextFormatter)
	customFormatter.ForceColors = true
	customFormatter.FullTimestamp = false
	customFormatter.DisableColors = false
	log.Formatter = customFormatter
	
	//Windows color support
	log.SetOutput(colorable.NewColorableStdout())
}

func uiStringArray(title string, arr []string) {
	if len(arr) == 0 {
		log.Printf("%s: None", title)
		return
	}
	log.Printf("%s:", title)
	for _, stri := range arr {
		log.Printf("\t %s", stri)
	}
}

//FIXME:
func uiList(list interface{}) {
	log.Printf("%v", list)
	// log.Printf("Count: %d,\t Next: %d,\t  Previous:%d ", list.Count, list.Next, list.Previous)
	// for i, v := range list.Names {
	// 	log.Printf("%d. %s", i, v)
	// }
}

//FIXME:
func uiImageInfo(res *uvApi.ImgStatusRes) {
	log.Printf("%v", res)
}

func uiServicStatus(srv *uvApi.SrvStatusRes) {
	log.Printf("Service Name: %s ", srv.Name)
	log.Printf("Plan Name: %s ", srv.Plan)
	log.Printf("State: %v ", srv.State.String())
	log.Printf("Created: %v,\t Updated: %v ", srv.Created, srv.Updated)
	uiMap(srv.Variable, "Variable")
	uiStringArray("List of endpoints", srv.Endpoints)
	uiAttachedDomains(srv.Domains)
}

func uiPortforward(in *uvApi.PortforwardRes) {
	bearer := string(in.Token)
	proxyURL, _ := url.Parse(in.ProxyHost)
	conf := &rest.Config{
		BearerToken:     bearer,
		Host:            fmt.Sprintf("%s://%s", proxyURL.Scheme, proxyURL.Host),
		TLSClientConfig: rest.TLSClientConfig{Insecure: true},
	}
	transport, upgrader, err := spdy.RoundTripperFor(conf)
	if err != nil {
		panic(err.Error())
	}

	var done = make(chan struct{}, 1)
	var rdy = make(chan struct{})
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	defer signal.Stop(signals)

	go func() {
		fmt.Printf("Forwarding ports: %s", in.Ports)
		<-signals
		fmt.Print("closing the opened ports...")
		if done != nil {
			close(done)
		}
		os.Exit(1)
	}()
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, proxyURL)
	pf, err := portforward.New(dialer, in.Ports, done, rdy, &stdout, &stderr)
	if err != nil {
		panic(err.Error())
	}
	err = pf.ForwardPorts()
	if err != nil {
		panic(err.Error())
	}
}

func uiPlan(plan []*uvApi.Plan) {
	log.Println("Plan:")
	for i, p := range plan {
		log.Printf("%d. ", i)
		log.Printf("\t name: %s ", p.Name)
		//Deprecated
		// log.Printf("\t price: %v, off: %v ", p.Price, p.Off)
		log.Printf("\t Description: %v ", p.Description)
	}
}
func uiMap(mapVar map[string]string, name string) {
	if len(mapVar) == 0 {
		log.Printf("%s: None", name)
		return
	}
	log.Printf("%s:", name)
	for k, v := range mapVar {
		log.Printf("\t %s: %s ", k, v)
	}
}

func uiProduct(prd *uvApi.ProductRes) {
	log.Printf("Product Name: %s ", prd.Name)
	log.Printf("Description: %s ", prd.Description)
	uiPlan(prd.Plan)
	uiMap(prd.VariableHints, "Variable Hints")
}

func uiSettingByDetail(set *uvApi.SettingRes) {
	log.Printf("Setting Name: %s ", set.Name)
	log.Printf("Application: %s Path: %s", set.App, set.Path)
	log.Print("value: ")
	log.Print(set.File)
}

func uiSetting(set *uvApi.SettingRes) {
	log.Println(set.File)
}

func uiApplicationStatus(app *uvApi.AppStatusRes) {
	log.Printf("Service Name: %s ", app.Name)
	log.Printf("State: %v ", app.State)
	log.Printf("Config: %v ", app.Config)
	log.Printf("Created: %v,\t Updated: %v ", app.Created, app.Updated)
	log.Printf("VCAP_SERVICES: ")
	log.Print(jsonPrettyPrint(app.VcapServices))
	uiMap(app.EnvironmentVariables, "Environment variables")
	// uiAttachedDomains(app.Domains)
}

func uiAttachedDomains(domains []*uvApi.AttachedDomainInfo) {
	if len(domains) == 0 {
		log.Println("Attached domains: None")
		return
	}
	log.Println("Attached domains:")
	log.Println("Domain | Endpoint | Type")
	for i, d := range domains {
		log.Printf("%d. %s | %s | %s ", i, d.Domain, d.Endpoint, d.EndpointType)
	}
}

func uiApplicationLog(client uvApi.UV_AppLogClient) {
	var byteRecieved = 0
	for {
		c, err := client.Recv()
		if err != nil {
			if status.Code(err) == codes.OutOfRange {
				log.Printf("Transfer of %d bytes done", byteRecieved)
				return
			}
			log.Fatal(err)
		}
		byteRecieved += len(c.Chunk)
		log.Printf(string(c.Chunk))
	}
}

func uiDomainStatus(dom *uvApi.DomainStatusRes) {
	log.Printf("Domain Name: %s ", dom.Domain)
	log.Printf("Created: %v ,Update: %v", dom.Created, dom.Updated)
	log.Printf("AttachedTo: %s ", dom.AttachedTo)
	log.Printf("TLS: %s ", dom.Tls)
}

func uiVolumeSpec(vol *uvApi.VolumeSpec) {
	log.Printf("Volume Spec Name: %s ", vol.Name)
	log.Printf("\t Spec Class: %s ", vol.Class)
	log.Printf("\t Spec Size: %v ", vol.Size)
}

func uiVolumeStatus(vol *uvApi.VolumeStatusRes) {
	log.Printf("Volume Name: %s ", vol.Name)
	log.Printf("Created: %v ,Update: %v", vol.Created, vol.Updated)
	log.Printf("AttachedTo: %s ", vol.AttachedTo)
	uiVolumeSpec(vol.Spec)
}

func uiCheckErr(info string, err error) {
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func jsonPrettyPrint(in string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(in), "", "\t")
	if err != nil {
		return in
	}
	return out.String()
}
