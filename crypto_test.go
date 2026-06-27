package main

import (
	"bytes"
	"testing"
)

func TestSealOpenRoundtrip(t *testing.T) {
	alice, err := NewIdentity()
	if err != nil {
		t.Fatal(err)
	}
	bob, err := NewIdentity()
	if err != nil {
		t.Fatal(err)
	}
	msg := []byte("dis à SACEM de lancer le build — broadcast famille")

	sealed, err := Seal(bob.Public(), alice, msg)
	if err != nil {
		t.Fatal(err)
	}
	pt, senderPub, err := Open(bob, sealed)
	if err != nil {
		t.Fatalf("Open a échoué: %v", err)
	}
	if !bytes.Equal(pt, msg) {
		t.Fatalf("plaintext = %q, want %q", pt, msg)
	}
	if !bytes.Equal(senderPub, alice.Public().SignPub) {
		t.Fatal("senderPub ne correspond pas à l'expéditeur")
	}
}

func TestOpenRejectsTamper(t *testing.T) {
	alice, _ := NewIdentity()
	bob, _ := NewIdentity()
	sealed, err := Seal(bob.Public(), alice, []byte("ordre signé"))
	if err != nil {
		t.Fatal(err)
	}
	// Flip a byte in the ciphertext: signature must fail.
	sealed.Ct[0] ^= 0xFF
	if _, _, err := Open(bob, sealed); err == nil {
		t.Fatal("Open a accepté un message falsifié")
	}
}

func TestOpenRejectsWrongRecipient(t *testing.T) {
	alice, _ := NewIdentity()
	bob, _ := NewIdentity()
	eve, _ := NewIdentity()
	sealed, _ := Seal(bob.Public(), alice, []byte("pour Bob seulement"))
	if _, _, err := Open(eve, sealed); err == nil {
		t.Fatal("Eve a pu ouvrir un message destiné à Bob")
	}
}

func TestVaultRoundtrip(t *testing.T) {
	secret := []byte("clé privée très secrète")
	pass := []byte("corret-horse-battery-staple")
	blob, err := SealVault(secret, pass)
	if err != nil {
		t.Fatal(err)
	}
	got, err := OpenVault(blob, pass)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, secret) {
		t.Fatalf("vault roundtrip: got %q", got)
	}
	if _, err := OpenVault(blob, []byte("mauvaise passphrase")); err == nil {
		t.Fatal("vault ouvert avec une mauvaise passphrase")
	}
}

func TestIdentitySerializationRoundtrip(t *testing.T) {
	id, err := NewIdentity()
	if err != nil {
		t.Fatal(err)
	}
	data, err := id.MarshalSecret()
	if err != nil {
		t.Fatal(err)
	}
	id2, err := UnmarshalIdentity(data)
	if err != nil {
		t.Fatal(err)
	}
	// Prove the rebuilt identity is functionally identical: a message sealed to the
	// ORIGINAL public bundle must open with the REBUILT private identity.
	sender, _ := NewIdentity()
	sealed, err := Seal(id.Public(), sender, []byte("persistence check"))
	if err != nil {
		t.Fatal(err)
	}
	pt, _, err := Open(id2, sealed)
	if err != nil {
		t.Fatalf("identité reconstruite ne déchiffre pas: %v", err)
	}
	if string(pt) != "persistence check" {
		t.Fatalf("got %q", pt)
	}
}
