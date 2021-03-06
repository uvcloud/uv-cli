package cmd

import (
	"log"

	"github.com/spf13/cobra"
	ybApi "github.com/yottab/proto-api/proto"
)

var (
	flagTLS bool
)

// DomainList .
func DomainList(appName string, index int32) (*ybApi.DomainListRes, error) {
	req := getRequestIndexForApp(appName, index)
	client := grpcConnect()
	defer client.Close()
	return client.V2().DomainList(client.Context(), req)
}
func domainList(cmd *cobra.Command, args []string) {
	appName := getCliArg(args, 0, "")
	res, err := DomainList(appName, flagIndex)
	uiCheckErr("Could not List the domain", err)
	uiList(res)
}

func domainInfo(cmd *cobra.Command, args []string) {
	res, err := DomainInfo(
		getCliRequiredArg(args, 0))
	uiCheckErr("Could not get the Domains", err)
	uiDomainStatus(res)
}

// DomainInfo get domain informatin
func DomainInfo(name string) (*ybApi.DomainStatusRes, error) {
	req := getRequestIdentity(name)
	client := grpcConnect()
	defer client.Close()
	return client.V2().DomainInfo(client.Context(), req)
}

// DomainCreate create domain
func DomainCreate(domain string, tls bool) (*ybApi.DomainStatusRes, error) {
	req := new(ybApi.DomainCreateReq)
	req.Domain = domain
	req.Tls = tls

	client := grpcConnect()
	defer client.Close()
	return client.V2().DomainCreate(client.Context(), req)
}
func domainCreate(cmd *cobra.Command, args []string) {
	res, err := DomainCreate(
		getCliRequiredArg(args, 0),
		flagTLS)

	uiCheckErr("Could not Create the Domain", err)
	uiDomainStatus(res)
}

func domainDelete(cmd *cobra.Command, args []string) {
	req := getCliRequestIdentity(args, 0)
	client := grpcConnect()
	defer client.Close()
	_, err := client.V2().DomainDelete(client.Context(), req)
	uiCheckErr("Could not Delete the Domain", err)
	log.Println("Task is done.")
}
