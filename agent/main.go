package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"encoding/json"
	"github.com/kelseyhightower/envconfig"
	"github.com/parnurzeal/gorequest"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ConfigOptions struct {
	ServerAddress string `envconfig:"server_address"`
	PrivateKey    string `envconfig:"private_key"`
	TenantID      string `envconfig:"tenant_id"`
}

type Endpoints struct {
	API  string `json:"api"`
	SSH  string `json:"ssh"`
	MQTT string `json:"mqtt"`
}

func (e *Endpoints) buildAPIUrl(uri string) string {
	return fmt.Sprintf("http://%s/api/%s", e.API, uri)
}

func sendAuthRequest(endpoints *Endpoints, identity *DeviceIdentity, attributes *DeviceAttributes, pubKey *rsa.PublicKey, tenantID string, sessions []string) (*AuthResponse, error) {
	var auth AuthResponse

	_, _, errs := gorequest.New().Post(endpoints.buildAPIUrl("/devices/auth")).Send(&AuthRequest{
		Identity:   identity,
		Attributes: attributes,
		PublicKey: string(pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(pubKey),
		})),
		TenantID: tenantID,
		Sessions: sessions,
	}).EndStruct(&auth)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	return &auth, nil
}

type Information struct {
	SSHID string `json:"sshid"`
}

func getInfo(input string) string {
	info := Information{
		SSHID: input,
	}
	prettyJSON, err := json.MarshalIndent(info, "", "    ")
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Fatal("Failed to generate json")
	}
	return string(prettyJSON)
}
func main() {
	opts := ConfigOptions{}

	err := envconfig.Process("", &opts)
	if err != nil {
		logrus.Panic(err)
	}

	endpoints := Endpoints{}

	_, _, errs := gorequest.New().Get(fmt.Sprintf("%s/endpoints", opts.ServerAddress)).EndStruct(&endpoints)
	if len(errs) > 0 {
		logrus.WithFields(logrus.Fields{"err": errs[0]}).Fatal("Failed to get endpoints")
	}

	identity, err := GetDeviceIdentity()
	if err != nil {
		logrus.Fatal(err)
	}

	attributes, err := GetDeviceAttributes()
	if err != nil {
		logrus.Fatal(err)
	}

	if _, err := os.Stat(opts.PrivateKey); os.IsNotExist(err) {
		logrus.Info("Private key not found. Generating...")
		err := generatePrivateKey(opts.PrivateKey)
		if err != nil {
			logrus.Fatal(err)
		}
	}

	pubKey, err := readPublicKey(opts.PrivateKey)
	if err != nil {
		logrus.Fatal(err)
	}

	auth, err := sendAuthRequest(&endpoints, identity, attributes, pubKey, opts.TenantID, []string{})
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Panic("Failed authenticate device")
	}
	if l := len(os.Args); l > 1 && os.Args[1] == "info" {
		fmt.Println(getInfo(auth.Namespace + "." + auth.Name + "@" + strings.Split(endpoints.SSH, ":")[0]))
		return
	}

	freePort, err := getFreePort()
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "failed to get free port"))
	}

	server := NewSSHServer(opts.PrivateKey, freePort)
	client := NewSSHClient(opts.PrivateKey, endpoints.SSH, freePort)

	server.SetDeviceName(auth.Name)

	go func() {
		logrus.Fatal(server.ListenAndServe())
	}()

	b := NewBroker(endpoints.MQTT, auth.UID, auth.Token)

	b.Subscribe(fmt.Sprintf("device/%s/session/+/open", auth.UID), client.connect)
	b.Subscribe(fmt.Sprintf("device/%s/session/+/close", auth.UID), client.close)
	b.Connect()

	ticker := time.NewTicker(10 * time.Second)

	for _ = range ticker.C {
		auth, err = sendAuthRequest(&endpoints, identity, attributes, pubKey, opts.TenantID, client.Sessions)
		if err == nil {
			server.SetDeviceName(auth.Name)
		}
	}
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
