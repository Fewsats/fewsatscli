package macaroons

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"
	"gopkg.in/macaroon.v2"
)

type Macaroon struct {
	ID          string   `json:"ID"`
	Version     uint16   `json:"version"`
	PaymentHash string   `json:"payment_hash"`
	TokenID     string   `json:"token_id"`
	Location    string   `json:"location"`
	Caveats     []string `json:"caveats"`
}

var byteOrder = binary.BigEndian

var decodeCommand = &cli.Command{
	Name:      "decode",
	Usage:     "Decode a macaroon token",
	Action:    decode,
	ArgsUsage: "[token]",
}

// DecodeMacIdentifier decodes the macaroon identifier into its version,
// payment hash and user ID.
func DecodeMacIdentifier(id []byte) (uint16, [32]byte, [32]byte, error) {
	r := bytes.NewReader(id)

	var version uint16
	if err := binary.Read(r, byteOrder, &version); err != nil {
		return 0, [32]byte{}, [32]byte{}, err
	}

	switch version {
	// A version 0 identifier consists of its linked payment hash, followed
	// by the user ID.
	case 0:
		var paymentHash [32]byte
		if _, err := r.Read(paymentHash[:]); err != nil {
			return 0, [32]byte{}, [32]byte{}, err
		}
		var tokenID [32]byte
		if _, err := r.Read(tokenID[:]); err != nil {
			return 0, [32]byte{}, [32]byte{}, err
		}

		return version, paymentHash, tokenID, nil
	}

	return 0, [32]byte{}, [32]byte{}, fmt.Errorf("unkown version: %d", version)
}

func decode(c *cli.Context) error {
	if c.Args().Len() != 1 {
		return cli.Exit("A single macaroon token is required", 1)
	}

	token := c.Args().Get(0)
	macBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Unable to decode base64 macaroon: %v", err), 1)
	}

	mac := &macaroon.Macaroon{}
	if err := mac.UnmarshalBinary(macBytes); err != nil {
		return cli.Exit(fmt.Sprintf("Failed to unmarshal macaroon: %v", err), 1)
	}

	version, paymentHash, tokenID, err := DecodeMacIdentifier(mac.Id())
	if err != nil {
		fmt.Printf("{\"Error\": \"Failed to decode identifier, printing raw ID: %s\"}\n", mac.Id())
		return nil
	}

	decoded := Macaroon{
		ID:          fmt.Sprintf("%x", mac.Id()),
		Version:     version,
		PaymentHash: fmt.Sprintf("%x", paymentHash),
		TokenID:     fmt.Sprintf("%x", tokenID),
		Location:    mac.Location(),
		Caveats:     []string{},
	}

	for _, caveat := range mac.Caveats() {
		decoded.Caveats = append(decoded.Caveats, string(caveat.Id))
	}

	jsonOutput, _ := json.MarshalIndent(decoded, "", "  ")
	fmt.Println(string(jsonOutput))

	return nil
}
