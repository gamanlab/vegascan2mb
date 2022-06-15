// package main
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/debug"
	"github.com/gopcua/opcua/ua"
)

func main() {
	// var (
	// 	endpoint = flag.String("endpoint", "opc.tcp://CIMPLICITY-JMR_BOILER@DESKTOP-N8KS9G1", "OPC UA Endpoint URL")
	// 	nodeID   = flag.String("node", "N=3;s=FTCA01_TTL.Value", "NodeID to read")
	// 	value    = flag.String("value", "12345678", "value")
	// 	username = flag.String("user", "", "user name")
	// 	password = flag.String("password", "", "password")
	// )
	var (
		endpoint = flag.String("endpoint", "opc.tcp://127.0.0.1:51800", "OPC UA Endpoint URL")
		policy   = flag.String("policy", "None", "Security policy: None, Basic128Rsa15, Basic256, Basic256Sha256. Default: auto")
		mode     = flag.String("mode", "None", "Security mode: None, Sign, SignAndEncrypt. Default: auto")
		certFile = flag.String("cert", "cert.pem", "Path to cert.pem. Required for security mode/policy != None")
		keyFile  = flag.String("key", "key.pem", "Path to private key.pem. Required for security mode/policy != None")
		nodeID   = flag.String("node", "N=3;s=FTCA01_TTL.Value", "NodeID to read")
		value    = flag.String("value", "12345678", "value")
		username = flag.String("user", "ADMINISTRATOR", "user name")
		password = flag.String("password", "", "password")
		anonym   = flag.Bool("anonym", false, "anonymouse")
	)
	flag.BoolVar(&debug.Enable, "debug", false, "enable debug logging")
	flag.Parse()
	log.SetFlags(0)

	// add an arbitrary timeout to demonstrate how to stop a subscription
	// with a context.
	d := 60 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()

	endpoints, err := opcua.GetEndpoints(ctx, *endpoint)
	if err != nil {
		log.Fatal(err)
	}
	ep := opcua.SelectEndpoint(endpoints, *policy, ua.MessageSecurityModeFromString(*mode))
	if ep == nil {
		log.Fatal("Failed to find suitable endpoint")
	}

	fmt.Println("*", ep.SecurityPolicyURI, ep.SecurityMode)

	opts := []opcua.Option{
		opcua.SecurityPolicy(*policy),
		opcua.SecurityModeString(*mode),
		opcua.CertificateFile(*certFile),
		opcua.PrivateKeyFile(*keyFile),
		opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous),
	}

	if *anonym {
		opts = append(opts, opcua.AuthAnonymous())
	} else {
		opts = append(opts, opcua.AuthUsername(*username, *password))
	}

	c := opcua.NewClient(ep.EndpointURL, opts...)
	if err := c.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer c.CloseWithContext(ctx)

	id, err := ua.ParseNodeID(*nodeID)
	if err != nil {
		log.Fatalf("invalid node id: %v", err)
	}

	v, err := ua.NewVariant(*value)
	if err != nil {
		log.Fatalf("invalid value: %v", err)
	}

	req := &ua.WriteRequest{
		NodesToWrite: []*ua.WriteValue{
			{
				NodeID:      id,
				AttributeID: ua.AttributeIDValue,
				Value: &ua.DataValue{
					EncodingMask: ua.DataValueValue,
					Value:        v,
				},
			},
		},
	}

	resp, err := c.WriteWithContext(ctx, req)
	if err != nil {
		log.Fatalf("Read failed: %s", err)
	}
	log.Printf("%v", resp.Results[0])
}
