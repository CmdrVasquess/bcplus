package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	l "github.com/fractalqb/qblog"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/ssh/terminal"
)

var credsKey []byte

const prompt = "credentials masterkey: "

func promptCredsKey() []byte {
	fmt.Print("enter credentials masterkey: ")
	res, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
	return res
}

type CmdrCreds struct {
	CmpnjUser  string `json:",omitempty"`
	CmpnjPass  string `json:",omitempty"`
	EdsmApiKey string `json:",omitempty"`
}

func (cc *CmdrCreds) Clear() {
	if cc != nil {
		cc.EdsmApiKey = ""
	}
}

func (cc *CmdrCreds) Write(cmdr string, key []byte) error {
	glog.Logf(l.Info, "save credentials to %s/%s.pgp", dataDir, cmdr)
	filenm := filepath.Join(dataDir, cmdr+".pgp~")
	f, err := os.Create(filenm)
	if err != nil {
		return err
	}
	defer f.Close()
	arm, err := armor.Encode(f, "PGP MESSAGE", nil)
	if err != nil {
		return err
	}
	cwr, err := openpgp.SymmetricallyEncrypt(arm, key, nil, nil)
	if err != nil {
		return err
	}
	jenc := json.NewEncoder(cwr)
	jenc.Encode(cc)
	cwr.Close()
	arm.Close()
	f.Close()
	os.Rename(filenm, filepath.Join(dataDir, cmdr+".pgp"))
	return nil
}

func (cc *CmdrCreds) Read(cmdr string, key []byte) error {
	filenm := filepath.Join(dataDir, cmdr+".pgp")
	glog.Logf(l.Info, "load credentials from %s", filenm)
	if _, err := os.Stat(filenm); os.IsNotExist(err) {
		glog.Logf(lNotice, "commander %s's credentials do not exists", cmdr)
		return nil
	}
	f, err := os.Open(filenm)
	if err != nil {
		return err
	}
	defer f.Close()
	arm, err := armor.Decode(f)
	if err != nil {
		return err
	}
	md, err := openpgp.ReadMessage(
		arm.Body,
		nil,
		func(keys []openpgp.Key, symm bool) ([]byte, error) {
			if key == nil {
				return nil, errors.New("wrong password")
			} else {
				tmp := key
				key = nil
				return tmp, nil
			}
		},
		nil)
	if err != nil {
		return err
	}
	jdec := json.NewDecoder(md.UnverifiedBody)
	err = jdec.Decode(cc)
	if err != nil {
		return err
	}
	return nil
}
