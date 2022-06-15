// Copyright 2018-2020 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/debug"
	"github.com/gopcua/opcua/errors"
	"github.com/gopcua/opcua/ua"
)

var (
	endpoint    = flag.String("endpoint", "opc.tcp://localhost:4840", "OPC UA Endpoint URL")
	policy      = flag.String("sec-policy", "auto", "Security Policy URL or one of None, Basic128Rsa15, Basic256, Basic256Sha256")
	mode        = flag.String("sec-mode", "auto", "Security Mode: one of None, Sign, SignAndEncrypt")
	auth        = flag.String("auth-mode", "Anonymous", "Authentication Mode: one of Anonymous, UserName, Certificate")
	list        = flag.Bool("list", false, "List the policies supported by the endpoint and exit")
	username    = flag.String("user", "", "Username to use in auth-mode UserName; will prompt for input if omitted")
	password    = flag.String("pass", "", "Password to use in auth-mode UserName; will prompt for input if omitted")
	vegaAddress = flag.String("url", "http://127.0.0.1/val.csv", "URL of vegascan")
)

func main() {
	ReadTagsJson()

	flag.BoolVar(&debug.Enable, "debug", false, "enable debug logging")
	flag.Parse()
	log.SetFlags(0)

	ctx := context.Background()

	// Get a list of the endpoints for our target server
	endpoints, err := opcua.GetEndpoints(ctx, *endpoint)
	if err != nil {
		log.Fatal(err)
	}

	// User asked for just the list of options: print and quit
	if *list {
		printEndpointOptions(endpoints)
		return
	}

	// Get the options to pass into the client based on the flags passed into the executable
	opts := clientOptsFromFlags(endpoints)

	// Create a Client with the selected options
	c := opcua.NewClient(*endpoint, opts...)
	if err := c.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer c.CloseWithContext(ctx)

	// id, err := ua.ParseNodeID("ns=3;s=AirConditioner_2.TemperatureSetPoint")
	// id, err := ua.ParseNodeID("ns=3;s=FTCA01_TTL.Value")
	// if err != nil {
	// 	log.Fatalf("invalid node id: %v", err)
	// }

	// val, err := ua.NewVariant(float64(43.4))
	// if err != nil {
	// 	log.Fatalf("invalid value: %v", err)
	// }

	// req := &ua.WriteRequest{
	// 	NodesToWrite: []*ua.WriteValue{
	// 		{
	// 			NodeID:      id,
	// 			AttributeID: ua.AttributeIDValue,
	// 			Value: &ua.DataValue{
	// 				EncodingMask: ua.DataValueValue,
	// 				Value:        val,
	// 			},
	// 		},
	// 	},
	// }

	// resp, err := c.WriteWithContext(ctx, req)
	// if err != nil {
	// 	log.Fatalf("Write failed: %s", err)
	// }
	// log.Printf("%v", resp.Results[0])

	go func() {
		for {

			for _, t := range tags {
				writeTag(ctx, c, t)
			}
			time.Sleep(time.Second)
		}
	}()

	VegaToVars()

}

func clientOptsFromFlags(endpoints []*ua.EndpointDescription) []opcua.Option {
	opts := []opcua.Option{}

	var secPolicy string
	switch {
	case *policy == "auto":
		// set it later
	case strings.HasPrefix(*policy, ua.SecurityPolicyURIPrefix):
		secPolicy = *policy
		*policy = ""
	case *policy == "None" || *policy == "Basic128Rsa15" || *policy == "Basic256" || *policy == "Basic256Sha256" || *policy == "Aes128_Sha256_RsaOaep" || *policy == "Aes256_Sha256_RsaPss":
		secPolicy = ua.SecurityPolicyURIPrefix + *policy
		*policy = ""
	default:
		log.Fatalf("Invalid security policy: %s", *policy)
	}

	// Select the most appropriate authentication mode from server capabilities and user input
	authMode, authOption := authFromFlags([]byte{})
	opts = append(opts, authOption)

	var secMode ua.MessageSecurityMode
	switch strings.ToLower(*mode) {
	case "auto":
	case "none":
		secMode = ua.MessageSecurityModeNone
		*mode = ""
	case "sign":
		secMode = ua.MessageSecurityModeSign
		*mode = ""
	case "signandencrypt":
		secMode = ua.MessageSecurityModeSignAndEncrypt
		*mode = ""
	default:
		log.Fatalf("Invalid security mode: %s", *mode)
	}

	// Allow input of only one of sec-mode,sec-policy when choosing 'None'
	if secMode == ua.MessageSecurityModeNone || secPolicy == ua.SecurityPolicyURINone {
		secMode = ua.MessageSecurityModeNone
		secPolicy = ua.SecurityPolicyURINone
	}

	// Find the best endpoint based on our input and server recommendation (highest SecurityMode+SecurityLevel)
	var serverEndpoint *ua.EndpointDescription
	switch {
	case *mode == "auto" && *policy == "auto": // No user selection, choose best
		for _, e := range endpoints {
			if serverEndpoint == nil || (e.SecurityMode >= serverEndpoint.SecurityMode && e.SecurityLevel >= serverEndpoint.SecurityLevel) {
				serverEndpoint = e
			}
		}

	case *mode != "auto" && *policy == "auto": // User only cares about mode, select highest securitylevel with that mode
		for _, e := range endpoints {
			if e.SecurityMode == secMode && (serverEndpoint == nil || e.SecurityLevel >= serverEndpoint.SecurityLevel) {
				serverEndpoint = e
			}
		}

	case *mode == "auto" && *policy != "auto": // User only cares about policy, select highest securitylevel with that policy
		for _, e := range endpoints {
			if e.SecurityPolicyURI == secPolicy && (serverEndpoint == nil || e.SecurityLevel >= serverEndpoint.SecurityLevel) {
				serverEndpoint = e
			}
		}

	default: // User cares about both
		fmt.Println("secMode: ", secMode, "secPolicy:", secPolicy)
		for _, e := range endpoints {
			if e.SecurityPolicyURI == secPolicy && e.SecurityMode == secMode && (serverEndpoint == nil || e.SecurityLevel >= serverEndpoint.SecurityLevel) {
				serverEndpoint = e
			}
		}
	}

	if serverEndpoint == nil { // Didn't find an endpoint with matching policy and mode.
		log.Printf("unable to find suitable server endpoint with selected sec-policy and sec-mode")
		printEndpointOptions(endpoints)
		log.Fatalf("quitting")
	}

	secPolicy = serverEndpoint.SecurityPolicyURI
	secMode = serverEndpoint.SecurityMode

	// Check that the selected endpoint is a valid combo
	err := validateEndpointConfig(endpoints, secPolicy, secMode, authMode)
	if err != nil {
		log.Fatalf("error validating input: %s", err)
	}

	opts = append(opts, opcua.SecurityFromEndpoint(serverEndpoint, authMode))

	log.Printf("Using config:\nEndpoint: %s\nSecurity mode: %s, %s\nAuth mode : %s\n", serverEndpoint.EndpointURL, serverEndpoint.SecurityPolicyURI, serverEndpoint.SecurityMode, authMode)
	return opts
}

func authFromFlags(cert []byte) (ua.UserTokenType, opcua.Option) {
	var err error

	var authMode ua.UserTokenType
	var authOption opcua.Option
	switch strings.ToLower(*auth) {
	case "anonymous":
		authMode = ua.UserTokenTypeAnonymous
		authOption = opcua.AuthAnonymous()

	case "username":
		authMode = ua.UserTokenTypeUserName

		if *username == "" {
			fmt.Print("Enter username: ")
			*username, err = bufio.NewReader(os.Stdin).ReadString('\n')
			*username = strings.TrimSuffix(*username, "\n")
			if err != nil {
				log.Fatalf("error reading username input: %s", err)
			}
		}

		passPrompt := true
		flag.Visit(func(f *flag.Flag) {
			if f.Name == "pass" {
				passPrompt = false
			}
		})

		if passPrompt {
			fmt.Print("Enter password: ")
			passInput, err := terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				log.Fatalf("Error reading password: %s", err)
			}
			*password = string(passInput)
			fmt.Print("\n")
		}
		authOption = opcua.AuthUsername(*username, *password)

	case "certificate":
		authMode = ua.UserTokenTypeCertificate
		authOption = opcua.AuthCertificate(cert)

	case "issuedtoken":
		// todo: this is unsupported, fail here or fail in the opcua package?
		authMode = ua.UserTokenTypeIssuedToken
		authOption = opcua.AuthIssuedToken([]byte(nil))

	default:
		log.Printf("unknown auth-mode, defaulting to Anonymous")
		authMode = ua.UserTokenTypeAnonymous
		authOption = opcua.AuthAnonymous()

	}

	return authMode, authOption
}

func validateEndpointConfig(endpoints []*ua.EndpointDescription, secPolicy string, secMode ua.MessageSecurityMode, authMode ua.UserTokenType) error {
	for _, e := range endpoints {
		if e.SecurityMode == secMode && e.SecurityPolicyURI == secPolicy {
			for _, t := range e.UserIdentityTokens {
				if t.TokenType == authMode {
					return nil
				}
			}
		}
	}

	err := errors.Errorf("server does not support an endpoint with security : %s , %s", secPolicy, secMode)
	printEndpointOptions(endpoints)
	return err
}

func printEndpointOptions(endpoints []*ua.EndpointDescription) {
	log.Print("Valid options for the endpoint are:")
	log.Print("         sec-policy    |    sec-mode     |      auth-modes\n")
	log.Print("-----------------------|-----------------|---------------------------\n")
	for _, e := range endpoints {
		p := strings.TrimPrefix(e.SecurityPolicyURI, "http://opcfoundation.org/UA/SecurityPolicy#")
		m := strings.TrimPrefix(e.SecurityMode.String(), "MessageSecurityMode")
		var tt []string
		for _, t := range e.UserIdentityTokens {
			tok := strings.TrimPrefix(t.TokenType.String(), "UserTokenType")

			// Just show one entry if a server has multiple varieties of one TokenType (eg. different algorithms)
			dup := false
			for _, v := range tt {
				if tok == v {
					dup = true
					break
				}
			}
			if !dup {
				tt = append(tt, tok)
			}
		}
		log.Printf("%22s | %-15s | (%s)", p, m, strings.Join(tt, ","))
	}
}

func writeTag(ctx context.Context, c *opcua.Client, t Tag) {
	// id, err := ua.ParseNodeID("ns=3;s=AirConditioner_2.TemperatureSetPoint")
	id, err := ua.ParseNodeID(t.NodeId)
	if err != nil {
		log.Fatalf("invalid node id: %v", err)
	}

	val, err := ua.NewVariant(t.Value)
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
					Value:        val,
				},
			},
		},
	}

	resp, err := c.WriteWithContext(ctx, req)
	if err != nil {
		log.Fatalf("Write failed: %s", err)
	}
	log.Printf("%v", resp.Results[0])
}
