package cmdr

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/CmdrVasquess/goEDSMc"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/ssh/terminal"
)

const prompt = "credentials masterkey: "

func PromptCredsKey(prompt string) []byte {
	if len(prompt) > 0 {
		fmt.Print(prompt)
	} else {
		fmt.Print("enter credentials masterkey: ")
	}
	res, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	return res
}

type CmdrCreds struct {
	Changed bool `json:"-"`
	Edsm    edsm.Credentials
}

func (cc *CmdrCreds) Clear() {
	if cc != nil {
		cc.Edsm.Clear()
	}
}

func (cc *CmdrCreds) Write(wr io.Writer, key []byte) error {
	//	log.Logf(l.Info, "save credentials to %s/%s.pgp", dataDir, cmdr)
	//	filenm := filepath.Join(dataDir, cmdr+".pgp~")
	//	f, err := os.Create(filenm)
	//	if err != nil {
	//		return err
	//	}
	//	defer f.Close()
	if !cc.Changed {
		return nil
	}
	arm, err := armor.Encode(wr, "PGP MESSAGE", nil)
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
	//	os.Rename(filenm, filepath.Join(dataDir, cmdr+".pgp"))
	return nil
}

func (cc *CmdrCreds) Read(rd io.Reader, key []byte) error {
	arm, err := armor.Decode(rd)
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
	cc.Changed = false
	return nil
}
