package main

// crypto.go — couche 0 (sécurité) du bus csend.
//
// Primitives AUDITÉES uniquement (§38 : jamais de crypto maison), toutes dans la
// stdlib Go 1.24 → zéro dépendance externe :
//   · signature       Ed25519                 (crypto/ed25519)
//   · KEM classique   X25519                  (crypto/ecdh)
//   · KEM post-quant. ML-KEM-768              (crypto/mlkem)        ← hybride PQC
//   · AEAD            AES-256-GCM             (crypto/aes+cipher)
//   · KDF             HKDF-SHA256 / PBKDF2    (crypto/hkdf, pbkdf2)
//
// Un message est chiffré DE BOUT EN BOUT : le bus/relais ne voit que du chiffré
// signé (zero-trust). Le secret de session dérive de DEUX KEM combinés (X25519 ⊕
// ML-KEM) — il faut casser les deux pour lire (résistance « Harvest Now Decrypt
// Later », §38.7).

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/hkdf"
	"crypto/mlkem"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
)

const hkdfInfo = "csend/v1 hybrid-seal"

// Identity is an agent's long-term key material. The private halves live ONLY in
// the vault (SealVault); peers need only the PublicBundle to send to this identity.
type Identity struct {
	Sign   ed25519.PrivateKey
	X25519 *ecdh.PrivateKey
	MLKEM  *mlkem.DecapsulationKey768
}

// PublicBundle is the shareable public identity (what a sender needs).
type PublicBundle struct {
	SignPub   []byte `json:"sign_pub"`   // 32
	X25519Pub []byte `json:"x25519_pub"` // 32
	MLKEMPub  []byte `json:"mlkem_pub"`  // 1184
}

// SealedMessage is the on-wire E2E envelope. None of it reveals the plaintext.
type SealedMessage struct {
	EphX25519 []byte `json:"eph_x25519"` // ephemeral X25519 public key
	MLKEMCt   []byte `json:"mlkem_ct"`   // ML-KEM ciphertext
	Nonce     []byte `json:"nonce"`      // AES-GCM nonce
	Ct        []byte `json:"ct"`         // AES-GCM ciphertext+tag
	SenderPub []byte `json:"sender_pub"` // sender Ed25519 public key
	Sig       []byte `json:"sig"`        // Ed25519 signature over the transcript
}

// NewIdentity generates a fresh hybrid identity (classical + post-quantum).
func NewIdentity() (*Identity, error) {
	_, signPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	xPriv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	dk, err := mlkem.GenerateKey768()
	if err != nil {
		return nil, err
	}
	return &Identity{Sign: signPriv, X25519: xPriv, MLKEM: dk}, nil
}

// Public returns the shareable public bundle.
func (id *Identity) Public() PublicBundle {
	return PublicBundle{
		SignPub:   id.Sign.Public().(ed25519.PublicKey),
		X25519Pub: id.X25519.PublicKey().Bytes(),
		MLKEMPub:  id.MLKEM.EncapsulationKey().Bytes(),
	}
}

// transcript is the exact byte sequence that is both AEAD-bound and signed.
func transcript(ephPub, mlkemCt, nonce, ct []byte) []byte {
	t := make([]byte, 0, len(ephPub)+len(mlkemCt)+len(nonce)+len(ct))
	t = append(t, ephPub...)
	t = append(t, mlkemCt...)
	t = append(t, nonce...)
	t = append(t, ct...)
	return t
}

func deriveKey(xShared, mlShared []byte) ([]byte, error) {
	secret := append(append([]byte{}, xShared...), mlShared...)
	return hkdf.Key(sha256.New, secret, nil, hkdfInfo, 32)
}

// Seal encrypts plaintext to `to` and signs it as `from`. Hybrid KEM: the AEAD key
// derives from X25519 ⊕ ML-KEM, so confidentiality holds unless BOTH are broken.
func Seal(to PublicBundle, from *Identity, plaintext []byte) (*SealedMessage, error) {
	eph, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	toX, err := ecdh.X25519().NewPublicKey(to.X25519Pub)
	if err != nil {
		return nil, fmt.Errorf("clé X25519 destinataire invalide: %w", err)
	}
	xShared, err := eph.ECDH(toX)
	if err != nil {
		return nil, err
	}
	toML, err := mlkem.NewEncapsulationKey768(to.MLKEMPub)
	if err != nil {
		return nil, fmt.Errorf("clé ML-KEM destinataire invalide: %w", err)
	}
	mlShared, mlCt := toML.Encapsulate()

	key, err := deriveKey(xShared, mlShared)
	if err != nil {
		return nil, err
	}
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	ct := gcm.Seal(nil, nonce, plaintext, nil)

	ephPub := eph.PublicKey().Bytes()
	sig := ed25519.Sign(from.Sign, transcript(ephPub, mlCt, nonce, ct))
	return &SealedMessage{
		EphX25519: ephPub, MLKEMCt: mlCt, Nonce: nonce, Ct: ct,
		SenderPub: from.Sign.Public().(ed25519.PublicKey), Sig: sig,
	}, nil
}

// Open verifies the sender signature then decrypts. Returns the plaintext and the
// sender's Ed25519 public key (for authorization decisions by the caller).
func Open(id *Identity, m *SealedMessage) (plaintext, senderPub []byte, err error) {
	if len(m.SenderPub) != ed25519.PublicKeySize {
		return nil, nil, errors.New("clé d'expéditeur invalide")
	}
	if !ed25519.Verify(m.SenderPub, transcript(m.EphX25519, m.MLKEMCt, m.Nonce, m.Ct), m.Sig) {
		return nil, nil, errors.New("signature invalide (message falsifié ou mauvais expéditeur)")
	}
	ephPub, err := ecdh.X25519().NewPublicKey(m.EphX25519)
	if err != nil {
		return nil, nil, err
	}
	xShared, err := id.X25519.ECDH(ephPub)
	if err != nil {
		return nil, nil, err
	}
	mlShared, err := id.MLKEM.Decapsulate(m.MLKEMCt)
	if err != nil {
		return nil, nil, err
	}
	key, err := deriveKey(xShared, mlShared)
	if err != nil {
		return nil, nil, err
	}
	gcm, err := newGCM(key)
	if err != nil {
		return nil, nil, err
	}
	pt, err := gcm.Open(nil, m.Nonce, m.Ct, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("déchiffrement échoué: %w", err)
	}
	return pt, m.SenderPub, nil
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// --- Vault: identity private keys at rest, sealed by a passphrase ---

// pbkdf2Iters is deliberately high (OWASP-grade for PBKDF2-SHA256). Argon2id is the
// preferred upgrade (needs golang.org/x/crypto) — tracked in the design (§38).
const pbkdf2Iters = 600_000

// VaultBlob is the encrypted-at-rest serialization of secret material.
type VaultBlob struct {
	Salt  []byte `json:"salt"`
	Nonce []byte `json:"nonce"`
	Ct    []byte `json:"ct"`
}

// SealVault encrypts plaintext under a passphrase (PBKDF2-SHA256 → AES-256-GCM).
func SealVault(plaintext, passphrase []byte) (*VaultBlob, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	key, err := pbkdf2.Key(sha256.New, string(passphrase), salt, pbkdf2Iters, 32)
	if err != nil {
		return nil, err
	}
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return &VaultBlob{Salt: salt, Nonce: nonce, Ct: gcm.Seal(nil, nonce, plaintext, nil)}, nil
}

// OpenVault reverses SealVault.
func OpenVault(b *VaultBlob, passphrase []byte) ([]byte, error) {
	key, err := pbkdf2.Key(sha256.New, string(passphrase), b.Salt, pbkdf2Iters, 32)
	if err != nil {
		return nil, err
	}
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}
	pt, err := gcm.Open(nil, b.Nonce, b.Ct, nil)
	if err != nil {
		return nil, errors.New("vault: passphrase invalide ou coffre corrompu")
	}
	return pt, nil
}

// --- Identity serialization (private — only ever written inside a vault) ---

type identitySecret struct {
	SignSeed  []byte `json:"sign_seed"`   // 32 (ed25519 seed)
	X25519Key []byte `json:"x25519_key"`  // 32 (x25519 private scalar)
	MLKEMSeed []byte `json:"mlkem_seed"`  // 64 (ml-kem-768 seed)
}

// MarshalSecret serializes the private identity (to be wrapped by SealVault).
func (id *Identity) MarshalSecret() ([]byte, error) {
	return json.Marshal(identitySecret{
		SignSeed:  id.Sign.Seed(),
		X25519Key: id.X25519.Bytes(),
		MLKEMSeed: id.MLKEM.Bytes(),
	})
}

// UnmarshalIdentity rebuilds an Identity from its serialized secret.
func UnmarshalIdentity(data []byte) (*Identity, error) {
	var s identitySecret
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	xPriv, err := ecdh.X25519().NewPrivateKey(s.X25519Key)
	if err != nil {
		return nil, err
	}
	dk, err := mlkem.NewDecapsulationKey768(s.MLKEMSeed)
	if err != nil {
		return nil, err
	}
	return &Identity{
		Sign:   ed25519.NewKeyFromSeed(s.SignSeed),
		X25519: xPriv,
		MLKEM:  dk,
	}, nil
}
